package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"pvz-service/internal/logger" // Обновите импорт согласно вашему проекту

	"github.com/google/uuid"
)

// RequestIDKey для хранения ID запроса в контексте
type RequestIDKey struct{}

// LoggingMiddleware логирует информацию о HTTP запросах с использованием структурированного логгера
func LoggingMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Генерируем уникальный ID для запроса
			requestID := uuid.New().String()

			// Создаем логгер с контекстом запроса
			requestLog := log.With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Добавляем логгер и ID запроса в контекст
			ctx := logger.WithLogger(r.Context(), requestLog)
			ctx = context.WithValue(ctx, RequestIDKey{}, requestID)

			// Логируем начало запроса
			requestLog.Info("входящий запрос")

			// Создаем обертку для отслеживания статус-кода
			lrw := newLoggingResponseWriter(w)

			// Добавляем заголовок с ID запроса для отслеживания
			lrw.Header().Set("X-Request-ID", requestID)

			// Передаем управление следующему обработчику с обновленным контекстом
			next.ServeHTTP(lrw, r.WithContext(ctx))

			// Логируем результат запроса
			duration := time.Since(start)
			requestLog.Info("запрос обработан",
				"status", lrw.statusCode,
				"duration", duration.String(),
				"duration_ms", float64(duration.Microseconds())/1000.0,
			)
		})
	}
}

// loggingResponseWriter обертка над http.ResponseWriter для отслеживания кода ответа
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.written += n
	return n, err
}

// Добавляем методы для получения метрик
func (lrw *loggingResponseWriter) Status() int {
	return lrw.statusCode
}

func (lrw *loggingResponseWriter) Size() int {
	return lrw.written
}
