package repository

import (
	"context"

	"github.com/alexuryumtsev/gophermart/internal/model"
)

type BalanceRepositoryImpl struct {
	db PgxPool
}

// Переименовываем конструктор
func NewBalanceRepository(db PgxPool) BalanceRepository {
	return &BalanceRepositoryImpl{db: db}
}

func (r *BalanceRepositoryImpl) GetBalance(ctx context.Context, userID int) (*model.Balance, error) {
	balance := &model.Balance{}

	// Получаем сумму начислений
	accrualQuery := getQueryBalance()
	var accrual float64
	err := r.db.QueryRow(ctx, accrualQuery, userID, model.OperationTypeAccrual).Scan(&accrual)
	if err != nil {
		return nil, err
	}

	// Получаем сумму списаний
	withdrawalQuery := getQueryBalance()
	var withdrawal float64
	err = r.db.QueryRow(ctx, withdrawalQuery, userID, model.OperationTypeWithdrawal).Scan(&withdrawal)
	if err != nil {
		return nil, err
	}

	balance.Current = accrual - withdrawal
	balance.Withdrawn = withdrawal

	return balance, nil
}

func getQueryBalance() string {
	return `
		SELECT COALESCE(SUM(amount), 0)
		FROM balance_transactions
		WHERE user_id = $1 AND operation_type = $2
	`
}

func (r *BalanceRepositoryImpl) AddTransaction(
	ctx context.Context,
	userID int,
	orderNumber string,
	amount float64,
	operationType model.OperationType,
) (*model.BalanceTransaction, error) {
	tx := &model.BalanceTransaction{}
	query := `
        INSERT INTO balance_transactions (user_id, order_number, amount, operation_type)
        VALUES ($1, $2, $3, $4)
        RETURNING id, user_id, order_number, amount, operation_type, processed_at
    `
	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		orderNumber,
		amount,
		operationType,
	).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.OrderNumber,
		&tx.Amount,
		&tx.OperationType,
		&tx.ProcessedAt,
	)

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (r *BalanceRepositoryImpl) GetWithdrawals(ctx context.Context, userID int) ([]model.BalanceTransaction, error) {
	query := `
        SELECT id, user_id, order_number, amount, operation_type, processed_at
        FROM balance_transactions
        WHERE user_id = $1 AND operation_type = $2
        ORDER BY processed_at DESC
    `

	rows, err := r.db.Query(ctx, query, userID, model.OperationTypeWithdrawal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.BalanceTransaction
	for rows.Next() {
		var tx model.BalanceTransaction
		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.OrderNumber,
			&tx.Amount,
			&tx.OperationType,
			&tx.ProcessedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
