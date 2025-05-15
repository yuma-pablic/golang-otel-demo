package middlewares

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"otel/ctxx"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func TraceIDMiddleware(tracer trace.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := GenerateTraceID(r.Context(), tracer, r.Method, r.URL.Path)
			ctx := ctxx.SetTraceID(r.Context(), traceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GenerateTraceID(ctx context.Context, tracer trace.Tracer, method, path string) string {
	_, span := tracer.Start(ctx, fmt.Sprintf("%s %s", method, path))
	defer span.End()

	trace_id := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("user_%d", trace_id.Intn(100000))
}
