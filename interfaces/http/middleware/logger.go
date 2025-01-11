package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	buffer      *bytes.Buffer
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		buffer:         bytes.NewBuffer([]byte{}),
	}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	rw.buffer.Write(b)
	return rw.ResponseWriter.Write(b)
}

func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}

		value := values[0]
		key = strings.ToLower(key)

		if key == "authorization" && strings.HasPrefix(value, "Bearer ") {
			parts := strings.Split(value, ".")
			if len(parts) == 3 {
				value = parts[0] + "." + parts[1] + ".REDACTED"
			}
		}

		sanitized[key] = value
	}
	return sanitized
}

func sanitizeBody(body []byte) (map[string]interface{}, error) {
	if len(body) == 0 {
		return nil, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	sensitiveFields := []string{
		"password",
		"password_hash",
		"token",
		"access_token",
		"refresh_token",
		"secret",
		"api_key",
	}

	var redact func(v interface{}) interface{}
	redact = func(v interface{}) interface{} {
		switch t := v.(type) {
		case map[string]interface{}:
			result := make(map[string]interface{})
			for k, val := range t {
				isSensitive := false
				for _, sf := range sensitiveFields {
					if strings.Contains(strings.ToLower(k), sf) {
						isSensitive = true
						break
					}
				}

				if isSensitive {
					result[k] = "REDACTED"
				} else {
					result[k] = redact(val)
				}
			}
			return result
		case []interface{}:
			result := make([]interface{}, len(t))
			for i, val := range t {
				result[i] = redact(val)
			}
			return result
		default:
			return v
		}
	}

	return redact(data).(map[string]interface{}), nil
}

func LoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			var reqBody []byte
			if r.Body != nil {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			}

			requestID := middleware.GetReqID(r.Context())

			sanitizedReqBody, _ := sanitizeBody(reqBody)
			logger.Info("API Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("request_id", requestID),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Any("headers", sanitizeHeaders(r.Header)),
				zap.Any("body", sanitizedReqBody),
			)

			next.ServeHTTP(wrapped, r)

			sanitizedResBody, _ := sanitizeBody(wrapped.buffer.Bytes())
			logger.Info("API Response",
				zap.String("request_id", requestID),
				zap.Int("status", wrapped.status),
				zap.Duration("duration", time.Since(start)),
				zap.Any("headers", sanitizeHeaders(wrapped.Header())),
				zap.Any("body", sanitizedResBody),
			)
		})
	}
}
