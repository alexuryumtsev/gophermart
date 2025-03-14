// repository/order.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexuryumtsev/gophermart/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, number string, userID int) (*model.Order, error) {
	// Проверяем, существует ли заказ в системе
	existingOrder, err := r.GetByNumber(ctx, number)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			return existingOrder, nil // Заказ уже загружен этим пользователем
		}
		return nil, fmt.Errorf("order already exists for another user")
	}

	// Создаем заказ
	order := &model.Order{
		Number:     number,
		UserID:     userID,
		Status:     model.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	query := `
        INSERT INTO orders (number, user_id, status, uploaded_at)
        VALUES ($1, $2, $3, $4)
        RETURNING number, user_id, status, accrual, uploaded_at
    `

	err = r.db.QueryRow(
		ctx,
		query,
		order.Number,
		order.UserID,
		order.Status,
		order.UploadedAt,
	).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) GetByNumber(ctx context.Context, number string) (*model.Order, error) {
	order := &model.Order{}
	query := `SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1`

	err := r.db.QueryRow(ctx, query, number).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	query := `
        SELECT number, user_id, status, accrual, uploaded_at
        FROM orders
        WHERE user_id = $1
        ORDER BY uploaded_at DESC
    `

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, number string, status model.OrderStatus, accrual *float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`
	_, err := r.db.Exec(ctx, query, status, accrual, number)
	return err
}

func (r *OrderRepository) GetOrdersForProcessing(ctx context.Context) ([]model.Order, error) {
	query := `
        SELECT number, user_id, status, accrual, uploaded_at
        FROM orders
        WHERE status IN ($1, $2)
        ORDER BY uploaded_at ASC
    `

	rows, err := r.db.Query(ctx, query, model.OrderStatusNew, model.OrderStatusProcessing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
