package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool() *pgxpool.Pool {
	connStr := getConnStr()
	ctx := context.Background()

	db, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	return db
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
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	db := NewPool()
	defer db.Close()

	r := chi.NewRouter()

	// /healthz エンドポイントで SELECT 1 を実行
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		var result int
		err := db.QueryRow(context.Background(), "SELECT 1").Scan(&result)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "DB result: %d", result)
	})

	// サーバー起動
	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}
