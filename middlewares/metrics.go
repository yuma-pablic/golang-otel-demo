package middlewares

import (
	"net/http"
	utils "otel/utils"
)

func MetricsMiddleware(metrics *utils.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.Requests.WithLabelValues(r.URL.Path).Inc()

			next.ServeHTTP(w, r)
		})
	}
}
