package middlewares

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const TraceIDKey = contextKey("traceId")

func TraceIDMiddleware(tracer trace.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := GenerateTraceID(r.Context(), tracer, r.Method, r.URL.Path)
			ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GenerateTraceID(ctx context.Context, tracer trace.Tracer, method, path string) string {
	_, span := tracer.Start(ctx, fmt.Sprintf("%s %s", method, path))
	defer span.End()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("user_%d", rnd.Intn(100000))
}
