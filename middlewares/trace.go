package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"otel/ctxx"

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

	return span.SpanContext().TraceID().String()
}
