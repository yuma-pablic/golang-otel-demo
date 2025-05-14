package middleware

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

func AddTraceID(tracer trace.Tracer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := GenerateTraceID(r.Context(), tracer, r.Method, r.URL.Path)
		ctx := context.WithValue(r.Context(), TraceIDKey, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GenerateTraceID(ctx context.Context, tracer trace.Tracer, method, path string) string {
	_, span := tracer.Start(ctx, fmt.Sprintf("%s %s", method, path))
	defer span.End()

	rndNum := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000)
	return fmt.Sprintf("user_%d", rndNum)
}
