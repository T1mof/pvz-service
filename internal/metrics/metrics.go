package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Технические метрики
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP запросов",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Время выполнения HTTP запросов в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// Бизнес-метрики
	pvzCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Общее количество созданных ПВЗ",
		},
	)

	receptionsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "receptions_created_total",
			Help: "Общее количество созданных приёмок заказов",
		},
	)

	productsAddedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Общее количество добавленных товаров",
		},
	)
)

// InitMetrics инициализирует метрики (при необходимости)
func InitMetrics() {

}

// IncrementPVZCreated увеличивает счетчик созданных ПВЗ
func IncrementPVZCreated() {
	pvzCreatedTotal.Inc()
}

// IncrementReceptionCreated увеличивает счетчик созданных приемок
func IncrementReceptionCreated() {
	receptionsCreatedTotal.Inc()
}

// IncrementProductAdded увеличивает счетчик добавленных товаров
func IncrementProductAdded() {
	productsAddedTotal.Inc()
}

// PrometheusMiddleware измеряет HTTP-запросы
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := newWrappedResponseWriter(w)

		next.ServeHTTP(ww, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(ww.status)

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, statusCode).Observe(duration)
	})
}

// wrappedResponseWriter - обертка для http.ResponseWriter для получения статус-кода
type wrappedResponseWriter struct {
	http.ResponseWriter
	status int
}

func newWrappedResponseWriter(w http.ResponseWriter) *wrappedResponseWriter {
	return &wrappedResponseWriter{w, http.StatusOK}
}

func (ww *wrappedResponseWriter) WriteHeader(code int) {
	ww.status = code
	ww.ResponseWriter.WriteHeader(code)
}
