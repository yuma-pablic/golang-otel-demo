package utils

import (
	"context"
	"log/slog"
	"os"

	log "otel/log"

	"go.opentelemetry.io/otel/trace"
)

type LoggerProvider struct {
	base *slog.Logger
}

func (p *LoggerProvider) WithTraceContext(ctx context.Context) *slog.Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return p.base
	}
	return p.base.With(
		slog.String("trace_id", sc.TraceID().String()),
		slog.String("span_id", sc.SpanID().String()),
	)
}

func NewLoggerProvider(serviceName string) (*LoggerProvider, error) {
	// 共通属性（service_name）
	serviceAttr := slog.Attr{
		Key:   "service_name",
		Value: slog.StringValue(serviceName),
	}

	stdoutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	stdoutHandlerWithAttrs := stdoutHandler.WithAttrs([]slog.Attr{serviceAttr})

	// Trace対応 + stdoutのみの MultiHandler
	traceAwareHandler := &log.TraceHandler{
		Handler: stdoutHandlerWithAttrs,
	}

	logger := slog.New(traceAwareHandler)
	slog.SetDefault(logger)

	return &LoggerProvider{base: logger}, nil
}
