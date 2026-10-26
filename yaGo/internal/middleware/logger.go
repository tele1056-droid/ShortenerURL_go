package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Логирующее middleware, которое оборачивает http.ResponseWriter,
// чтобы перехватить код статуса и размер ответа.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// LoggerMiddleware логирует информацию о запросе и ответе.
func LoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Оборачиваем ResponseWriter, чтобы перехватить код и размер
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // по умолчанию, если WriteHeader не вызван
			}

			// Передаем управление дальше по цепочке
			next.ServeHTTP(rw, r)

			// Логируем после завершения обработки
			logger.Info("request handled",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Duration("duration", time.Since(start)),
				zap.Int("status", rw.statusCode),
				zap.Int("size", rw.size),
			)
		})
	}
}
