package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"otel/utils"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riandyrn/otelchi"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var db *pgxpool.Pool

const (
	addr        = ":8080"
	serviceName = "back-svc"
)

func main() {
	ctx := context.Background()

	// トレーサーとトレーサープロバイダ初期化
	tracer, tp, err := utils.NewTracer(serviceName)
	if err != nil {
		log.Fatalf("Tracer init failed: %v", err)
	}
	defer func() {
		_ = tp.Shutdown(ctx)
	}()

	// DB接続
	db, err = initDB(ctx, tp)
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}
	defer db.Close()

	// ルーター設定
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(otelchi.Middleware(serviceName, otelchi.WithChiRoutes(r)))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		name := generateName(ctx, tracer)

		var result int
		err := db.QueryRow(ctx, "SELECT 1").Scan(&result)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Name: %s, DB result: %d", name, result)
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

func generateName(ctx context.Context, tracer trace.Tracer) string {
	_, span := tracer.Start(ctx, "generateName")
	defer span.End()

	rndNum := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000)
	return fmt.Sprintf("user_%d\n", rndNum)
}
