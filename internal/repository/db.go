// repository/db.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Функция для создания пула соединений
func NewDB(dsn string) (*pgxpool.Pool, error) {
	ctx := context.Background()
	return pgxpool.New(ctx, dsn)
}

// Инициализация схемы базы данных
func InitDB(db *pgxpool.Pool) error {
	ctx := context.Background()

	// Создаем таблицы
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        login VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS orders (
        number VARCHAR(255) PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id),
        status VARCHAR(20) NOT NULL,
        accrual NUMERIC(10, 2),
        uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS balance_transactions (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users(id),
        order_number VARCHAR(255) NOT NULL,
        amount NUMERIC(10, 2) NOT NULL,
        operation_type VARCHAR(20) NOT NULL,
        processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    `

	_, err := db.Exec(ctx, schema)
	return err
}
