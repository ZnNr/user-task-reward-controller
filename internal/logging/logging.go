package logging

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

// LoggingMiddleware создает middleware для логирования HTTP запросов
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем ResponseWriter, который может отслеживать статус ответа
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			// Логируем входящий запрос
			logRequest(logger, r)

			// Передаем запрос дальше
			next.ServeHTTP(wrappedWriter, r)

			// Логируем результат запроса
			logResponse(logger, r, wrappedWriter, time.Since(start))
		})
	}
}

// logRequest логирует входящий HTTP запрос
func logRequest(logger *zap.Logger, r *http.Request) {
	logger.Info("Incoming request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	)
}

// logResponse логирует результат исходящего HTTP ответа
func logResponse(logger *zap.Logger, r *http.Request, rw *responseWriter, duration time.Duration) {
	logger.Info("Request completed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", rw.status),
		zap.Duration("duration", duration),
	)
}

// responseWriter оборачивает http.ResponseWriter для отслеживания статуса ответа
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader обрабатывает статус ответа
func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return // предотвратить дублирование
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

// Write отправляет данные и обновляет статус, если хедер еще не был записан
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK) // устанавливаем статус, если он еще не был установлен
	}
	return rw.ResponseWriter.Write(b)
}
