package model

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var schema = `
CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    balance FLOAT
);
`

func CreateDB() (*sqlx.DB, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	dsn := fmt.Sprintf("host=%s user=postgres password=11111111 dbname=banking sslmode=disable", host)
	db, err := sqlx.Connect("postgres", dsn)
	db.MustExec(schema)
	return db, err
}

func AddDBContext(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "db", db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
