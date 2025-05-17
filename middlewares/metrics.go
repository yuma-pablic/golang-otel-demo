package middlewares

import (
	"net/http"
	utils "otel/utils"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func MetricsMiddleware(tracer trace.Tracer, metrics *utils.Metrics, histogram metric.Float64Histogram) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "http.server.request")
			defer span.End()
			start := time.Now()

			next.ServeHTTP(w, r)
			path := r.URL.Path
			metrics.Requests.WithLabelValues(path).Inc()
			duration := time.Since(start).Seconds()

			metrics.Duration.WithLabelValues(path).Observe(duration)
			histogram.Record(
				ctx,
				float64(duration),
			)

		})
	}
}
