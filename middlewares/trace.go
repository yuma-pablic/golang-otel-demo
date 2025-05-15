package middlewares

import (
	"fmt"
	"net/http"
	"otel/ctxx"

	"go.opentelemetry.io/otel/trace"
)

func TraceIDMiddleware(tracer trace.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			defer span.End()

			traceID := span.SpanContext().TraceID().String()
			ctx = ctxx.SetTraceID(ctx, traceID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
