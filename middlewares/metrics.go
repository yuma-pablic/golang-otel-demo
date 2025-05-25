package middlewares

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func MetricsMiddleware(tracer trace.Tracer, histogram metric.Float64Histogram) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			_, span := tracer.Start(ctx, "main handler")
			defer span.End()

			startTime := time.Now()

			duration := time.Since(startTime)
			histogram.Record(
				ctx,
				float64(duration.Seconds()),
			)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
