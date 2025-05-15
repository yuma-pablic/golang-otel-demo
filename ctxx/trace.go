package ctxx

import "context"

type contextKey string

const TraceIDKey = contextKey("traceId")

func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(TraceIDKey).(string); ok {
		return v
	}
	return ""
}
