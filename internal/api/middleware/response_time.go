package middleware

import (
	"log"
	"net/http"
	"time"
)

// ResponseTimeMiddleware - промежуточное ПО для измерения времени обработки запросов
func ResponseTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("[%s] %s %s - Обработано за %v",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			duration)
	})
}
