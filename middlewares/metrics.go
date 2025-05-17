package middlewares

import (
	"context"
	"math/rand"
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
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			prosessing(ctx, tracer)

			duration := time.Since(startTime)
			histogram.Record(
				ctx,
				float64(duration.Seconds()),
			)

		})
	}
}

func prosessing(ctx context.Context, tracer trace.Tracer) {
	ctx, span := tracer.Start(ctx, "processing...")
	defer span.End()

	if rand.Float64() < 1.0/100.0 {
		funcAbnormal(ctx, tracer)
	} else {
		funcNormal(ctx, tracer)
	}
}

func funcNormal(ctx context.Context, tracer trace.Tracer) {
	_, span := tracer.Start(ctx, "funcNormal")
	defer span.End()
	time.Sleep(10 * time.Millisecond)
}

func funcAbnormal(ctx context.Context, tracer trace.Tracer) {
	_, span := tracer.Start(ctx, "funcAbNormal(Oh...taking a lot of time...)")
	defer span.End()
	time.Sleep(3 * time.Second)
}
