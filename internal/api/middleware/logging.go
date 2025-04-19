package middleware

import (
	"log"
	"net/http"
	"time"
)

// RequestLogger - middleware для логирования HTTP запросов
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := newLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %s %d %s",
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
			r.Proto,
			lrw.statusCode,
			duration,
		)
	})
}

// loggingResponseWriter - обертка над http.ResponseWriter для отслеживания кода ответа
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
