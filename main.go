package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"otel/handler"
	middlewares "otel/middlewares"
	"otel/utils"

	"github.com/exaring/otelpgx"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riandyrn/otelchi"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var db *pgxpool.Pool

const (
	addr        = ":8080"
	serviceName = "main-api"
)

func main() {
	ctx := context.Background()

	// ===== Logger初期化 =====
	logger, err := utils.NewLogger(serviceName)
	if err != nil {
		panic("Logger init failed: " + err.Error())
	}
	logger.Info("Logger initialized", slog.String("service", serviceName))

	// ===== Tracer初期化（otel SDK）=====
	tracer, tp, err := utils.NewTracer(serviceName)
	if err != nil {
		logger.Error("Tracer init failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error("Tracer shutdown failed", slog.String("error", err.Error()))
		}
	}()
	logger.Info("Tracer initialized")

	// ===== DB接続（otelpgx付き）=====
	db, err = initDB(ctx, tp)
	if err != nil {
		logger.Error("DB init failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("DB connected")

	// ===== Metrics初期化 =====
	metrics := utils.NewMetrics()
	logger.Info("Metrics initialized")

	// ===== HTTPルーターとミドルウェア設定 =====
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(otelchi.Middleware(serviceName, otelchi.WithChiRoutes(r)))
	r.Use(middlewares.TraceIDMiddleware(tracer))
	r.Use(middlewares.MetricsMiddleware(metrics))

	// /metrics エンドポイントなどのハンドラ登録
	handler.RegisterMetricsRoute(r)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var result int
		err := db.QueryRow(ctx, "SELECT 1").Scan(&result)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		slog.InfoContext(ctx, "health_check success")
	})

	log.Printf("Starting server on %s...\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func initDB(ctx context.Context, tp *sdktrace.TracerProvider) (*pgxpool.Pool, error) {
	connStr := getConnStr()
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithTracerProvider(tp))
	return pgxpool.NewWithConfig(ctx, cfg)
}

func getConnStr() string {
	user := getEnv("DB_USER", "admin")
	pass := getEnv("DB_PASS", "admin")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	dbname := getEnv("DB_NAME", "postgres")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, dbname)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
