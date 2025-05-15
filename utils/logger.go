package utils

import (
	"context"
	"log/slog"
	"os"
	"otel/ctxx"
	log "otel/log"
	"path/filepath"
)

func NewLogger(ctx context.Context, serviceName string) (*slog.Logger, error) {
	logDir := "logs"
	logPath := filepath.Join(logDir, "app.log")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	traceID := ctxx.GetTraceID(ctx)
	serviceAttr := []slog.Attr{
		{
			Key:   "service_name",
			Value: slog.StringValue(serviceName),
		},
		{
			Key:   "trace_id",
			Value: slog.StringValue(traceID),
		},
	}

	stdoutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	stdoutHandlerWithAttrs := stdoutHandler.WithAttrs(serviceAttr)

	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo})
	fileHandlerWithAttrs := fileHandler.WithAttrs(serviceAttr)

	baseHandler := log.NewMultiHandler(stdoutHandlerWithAttrs, fileHandlerWithAttrs)
	traceAwareHandler := &log.TraceHandler{Handler: baseHandler}

	logger := slog.New(traceAwareHandler)
	slog.SetDefault(logger)

	return logger, nil
}
