package logging

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger est une interface qui définit les méthodes de logging communes
type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
	Debug(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Fatal(msg string, fields ...any)
}

// loggerKey est la clé utilisée pour stocker le logger dans le context
type loggerKey struct{}

// responseWriter est un wrapper pour gin.ResponseWriter qui capture le status code et la réponse
type responseWriter struct {
	gin.ResponseWriter
	body        []byte
	statusCode  int
	writeStatus bool
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = b
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.writeStatus = true
	w.ResponseWriter.WriteHeader(code)
}

// FromContext récupère le logger du context
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return &zapLogger{logger: logger}
	}
	return &zapLogger{logger: zap.L()} // Logger par défaut si aucun n'est trouvé dans le context
}

// WithContext ajoute un logger au context
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// zapLogger est l'implémentation de Logger utilisant zap
type zapLogger struct {
	logger *zap.Logger
}

// convertToZapFields convertit les arguments variadic en zap.Field
func convertToZapFields(args ...any) []zap.Field {
	// Si un seul argument est passé et c'est une error, la traiter directement
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			return []zap.Field{zap.Error(err)}
		}
	}

	fields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				continue
			}

			// Traitement spécial pour les erreurs
			if err, ok := args[i+1].(error); ok {
				fields = append(fields, zap.Error(err))
				continue
			}

			fields = append(fields, zap.Any(key, args[i+1]))
		}
	}
	return fields
}

func (l *zapLogger) Info(msg string, fields ...any) {
	l.logger.Info(msg, convertToZapFields(fields...)...)
}

func (l *zapLogger) Error(msg string, fields ...any) {
	l.logger.Error(msg, convertToZapFields(fields...)...)
}

func (l *zapLogger) Fatal(msg string, fields ...any) {
	l.logger.Fatal(msg, convertToZapFields(fields...)...)
}

func (l *zapLogger) Debug(msg string, fields ...any) {
	l.logger.Debug(msg, convertToZapFields(fields...)...)
}

func (l *zapLogger) Warn(msg string, fields ...any) {
	l.logger.Warn(msg, convertToZapFields(fields...)...)
}

func InitLogger(env string) *zap.Logger {
	config := zap.NewProductionConfig()

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	// Remplace le logger global
	zap.ReplaceGlobals(logger)
	return logger
}
