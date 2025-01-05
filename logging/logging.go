package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	RequestIDKey = "request-id"
	charset      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength     = 6
)

type LogManager struct {
	logger *zap.Logger
	env    string
}

func NewLogManager(env string) *LogManager {
	logger := InitLogger(env)
	return &LogManager{
		logger: logger,
		env:    env,
	}
}

func (lm *LogManager) addTraceInfo(ctx *gin.Context, fields []zap.Field) []zap.Field {
	if span, exists := tracer.SpanFromContext(ctx.Request.Context()); exists {
		traceID := span.Context().TraceID()
		spanID := span.Context().SpanID()
		return append(fields,
			zap.Uint64("dd.trace_id", traceID),
			zap.Uint64("dd.span_id", spanID),
		)
	}
	return fields
}

func generateRequestID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (lm *LogManager) ResponseLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		writer := &responseWriter{
			ResponseWriter: c.Writer,
			statusCode:     200,
		}
		c.Writer = writer

		c.Next()

		duration := time.Since(start)
		loggerFromCtx := FromContext(c.Request.Context())

		loggerFromCtx.Info("API Response",
			"status", writer.statusCode,
			"response", string(writer.body),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"duration", duration.String(),
			"request_id", c.GetString(RequestIDKey),
			"client_ip", c.ClientIP(),
		)
	}
}

func (lm *LogManager) RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		var bodyBytes []byte
		if c.Request.Body != nil {
			// Lire le body
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// Restaurer le body pour les prochains middlewares
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Tenter de formater le JSON pour le logging
		var prettyJSON bytes.Buffer
		if len(bodyBytes) > 0 {
			if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err != nil {
				// Si ce n'est pas du JSON valide, utiliser le body brut
				prettyJSON = *bytes.NewBuffer(bodyBytes)
			}
		}

		loggerFromCtx := FromContext(c.Request.Context())
		loggerFromCtx.Info("API Request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"payload", prettyJSON.String(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"request_id", c.GetString(RequestIDKey),
		)

		c.Next()
	}
}

func (lm *LogManager) SetupHttp(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Set("logger", lm.logger)
		c.Next()
	})

	r.Use(lm.httpRequestIDMiddleware())

	r.Use(lm.RequestLoggerMiddleware())

	r.Use(lm.ResponseLoggerMiddleware())

	r.Use(ginzap.RecoveryWithZap(lm.logger, true))

	lm.logger.Info("HTTP logging initialized", zap.String("env", lm.env))
}

func (lm *LogManager) httpRequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

		fields := []zap.Field{zap.String("request_id", requestID)}
		fields = lm.addTraceInfo(c, fields)

		requestLogger := lm.logger.With(fields...)
		ctx := WithContext(c.Request.Context(), requestLogger)
		c.Request = c.Request.WithContext(ctx)
		c.Set("logger", requestLogger)

		c.Next()
	}
}
