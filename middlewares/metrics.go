package middlewares

import (
	"net/http"
	utils "otel/utils"
	"time"
)

func MetricsMiddleware(metrics *utils.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)
			path := r.URL.Path
			metrics.Requests.WithLabelValues(path).Inc()
			duration := time.Since(start).Seconds()

			metrics.Duration.WithLabelValues(path).Observe(duration)

		})
	}
}
