package utils

import (
	"log/slog"
	"os"
	log "otel/log"
	"path/filepath"
)

func NewLogger(serviceName string) (*slog.Logger, error) {
	logDir := "logs"
	logPath := filepath.Join(logDir, "app.log")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	serviceAttr := slog.String("service", serviceName)

	stdoutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	stdoutHandlerWithAttrs := stdoutHandler.WithAttrs([]slog.Attr{serviceAttr})

	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo})
	fileHandlerWithAttrs := fileHandler.WithAttrs([]slog.Attr{serviceAttr})

	baseHandler := log.NewMultiHandler(stdoutHandlerWithAttrs, fileHandlerWithAttrs)
	traceAwareHandler := &log.TraceHandler{Handler: baseHandler}

	logger := slog.New(traceAwareHandler)
	slog.SetDefault(logger)

	return logger, nil
}
