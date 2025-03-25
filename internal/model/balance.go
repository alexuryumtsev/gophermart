package model

import "time"

type OperationType string

const (
	OperationTypeAccrual    OperationType = "ACCRUAL"
	OperationTypeWithdrawal OperationType = "WITHDRAWAL"
)

type BalanceTransaction struct {
	ID            int           `json:"id" db:"id"`
	UserID        int           `json:"user_id" db:"user_id"`
	OrderNumber   string        `json:"order" db:"order_number"`
	Amount        float64       `json:"sum" db:"amount"`
	OperationType OperationType `json:"-" db:"operation_type"`
	ProcessedAt   time.Time     `json:"processed_at" db:"processed_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
