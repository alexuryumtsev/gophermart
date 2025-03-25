// internal/service/balance.go
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexuryumtsev/gophermart/internal/model"
	"github.com/alexuryumtsev/gophermart/internal/repository"
	"github.com/alexuryumtsev/gophermart/internal/validation"
)

var (
	ErrInvalidOrderFormat = errors.New("invalid order number format")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)

// BalanceServiceImpl реализует интерфейс BalanceService
type BalanceServiceImpl struct {
	balanceRepo repository.BalanceRepository
}

// NewBalanceService создает экземпляр BalanceService
func NewBalanceService(balanceRepo repository.BalanceRepository) BalanceService {
	return &BalanceServiceImpl{
		balanceRepo: balanceRepo,
	}
}

// GetUserBalance получает текущий баланс пользователя
func (s *BalanceServiceImpl) GetUserBalance(ctx context.Context, userID int) (*model.Balance, error) {
	balance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}

// WithdrawFunds списывает средства с баланса пользователя
func (s *BalanceServiceImpl) WithdrawFunds(ctx context.Context, userID int, orderNumber string, amount float64) error {
	// Валидация номера заказа
	if !validation.ValidateOrderNumber(orderNumber) {
		return ErrInvalidOrderFormat
	}

	// Получаем текущий баланс
	balance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user balance: %w", err)
	}

	// Проверяем достаточно ли средств
	if balance.Current < amount {
		return ErrInsufficientFunds
	}

	// Создаем транзакцию списания
	_, err = s.balanceRepo.AddTransaction(
		ctx,
		userID,
		orderNumber,
		amount,
		model.OperationTypeWithdrawal,
	)
	if err != nil {
		return fmt.Errorf("failed to add withdrawal transaction: %w", err)
	}

	return nil
}

// GetUserWithdrawals получает историю списаний пользователя
func (s *BalanceServiceImpl) GetUserWithdrawals(ctx context.Context, userID int) ([]model.BalanceTransaction, error) {
	withdrawals, err := s.balanceRepo.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user withdrawals: %w", err)
	}

	return withdrawals, nil
}
