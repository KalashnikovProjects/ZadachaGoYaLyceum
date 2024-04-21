package db_connect

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
)

type SQLQueryExec interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type SQLTXQueryExec interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func OpenDb(ctx context.Context, connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Создание таблиц, если они не существуют
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS operations_times (
			id SERIAL PRIMARY KEY,
			plus INTEGER DEFAULT 10 CHECK(plus >= 0 AND plus <= 100),
			minus INTEGER DEFAULT 10 CHECK(minus >= 0 AND minus <= 100),
			division INTEGER DEFAULT 10 CHECK(division >= 0 AND division <= 100),
			multiplication INTEGER DEFAULT 10 CHECK(multiplication >= 0 AND multiplication <= 100)
		);

		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			operations_time_id INTEGER NOT NULL 
		);

		CREATE TABLE IF NOT EXISTS expressions (
			id SERIAL PRIMARY KEY,
			need_to_do TEXT NOT NULL,
			status TEXT NOT NULL,
			result FLOAT8,
			start_time INTEGER,
			end_time INTEGER,
			user_id INTEGER NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS operations (
			id SERIAL PRIMARY KEY,
			znak TEXT NOT NULL,
			left_is_ready INTEGER,
			left_data FLOAT8,
			right_is_ready INTEGER,
			right_data FLOAT8,
			father_id INTEGER,
			son_side INTEGER NOT NULL,
			expression_id INTEGER NOT NULL
		);

		CREATE TABLE IF NOT EXISTS agents (
			id SERIAL PRIMARY KEY,
			status TEXT NOT NULL,
			status_text TEXT,
			ping_time INTEGER NOT NULL
		);

		
	`)
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
