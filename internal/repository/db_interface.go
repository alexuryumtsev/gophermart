package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// интерфейс для pgxpool.Pool для возможности мока
type PgxPool interface {
	// Основные методы для выполнения запросов
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row

	// Транзакции
	Begin(ctx context.Context) (pgx.Tx, error)

	// Закрытие соединения
	Close()
}
