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

	// 共通属性（service_name）を定義
	serviceAttr := slog.Attr{
		Key:   "service_name",
		Value: slog.StringValue(serviceName),
	}

	// stdout用ハンドラ（Text形式）に属性追加
	stdoutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	stdoutHandlerWithAttrs := stdoutHandler.WithAttrs([]slog.Attr{serviceAttr})

	// ファイル用ハンドラ（JSON形式）に属性追加
	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo})
	fileHandlerWithAttrs := fileHandler.WithAttrs([]slog.Attr{serviceAttr})

	// 両方のハンドラを組み合わせてトレース対応に包む
	baseHandler := log.NewMultiHandler(stdoutHandlerWithAttrs, fileHandlerWithAttrs)
	traceAwareHandler := &log.TraceHandler{Handler: baseHandler}

	logger := slog.New(traceAwareHandler)
	slog.SetDefault(logger)

	return logger, nil
}
