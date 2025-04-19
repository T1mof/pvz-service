package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// ResponseWriter - обертка для http.ResponseWriter для доступа к коду статуса
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// MetricsMiddleware создает middleware для сбора метрик по HTTP-запросам
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := NewResponseWriter(w)
		start := time.Now()

		next.ServeHTTP(ww, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(ww.statusCode)

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, statusCode).Observe(duration)
	})
}
