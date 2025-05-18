package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	middlewares "otel/middlewares"
	"otel/utils"

	"github.com/exaring/otelpgx"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var db *pgxpool.Pool

const (
	addr        = ":8080"
	serviceName = "main-api"
)

var (
	meter = otel.Meter("main-api")

	histogram metric.Float64Histogram
)

func main() {
	ctx := context.Background()

	// ===== Logger初期化 =====
	lp, err := utils.NewLoggerProvider("main-api")
	if err != nil {
		panic(err)
	}
	logger := lp.WithTraceContext(ctx)

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

	// ===== DB接続（otelpgx付き）=====
	db, err = initDB(ctx, tp)
	if err != nil {
		logger.Error("DB init failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// ===== Metrics初期化 =====
	mp, err := utils.NewMetrics()
	if err != nil {
		log.Fatalf("failed to initialize histogram: %v", err)
	}
	defer func() { _ = mp.Shutdown(context.Background()) }()

	// ===== Histogram（OTel）初期化 =====
	meter = otel.Meter(serviceName)
	histogram, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("A histogram of the HTTP request durations in seconds."),
		metric.WithUnit("s"),
	)
	if err != nil {
		log.Fatalf("failed to initialize histogram: %v", err)
	}

	// ===== HTTPルーターとミドルウェア設定 =====
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(otelchi.Middleware(serviceName, otelchi.WithChiRoutes(r)))
	r.Use(middlewares.TraceIDMiddleware(tracer))
	r.Use(middlewares.MetricsMiddleware(tracer, histogram))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var result int
		err := db.QueryRow(ctx, "SELECT 1").Scan(&result)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		logger.InfoContext(ctx, "health_check success")
	})

	logger.Info("Starting server on %s...")
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
