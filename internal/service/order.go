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
	ErrInvalidOrderNumberFormat = errors.New("invalid order number format")
)

// OrderServiceImpl реализует интерфейс OrderService
type OrderServiceImpl struct {
	orderRepo repository.OrderRepository
}

// NewOrderService создает экземпляр OrderService
func NewOrderService(orderRepo repository.OrderRepository) OrderService {
	return &OrderServiceImpl{
		orderRepo: orderRepo,
	}
}

// CreateOrder создает новый заказ
func (s *OrderServiceImpl) CreateOrder(ctx context.Context, number string, userID int) (*model.Order, error) {
	// Валидация номера заказа
	if !s.ValidateOrderNumber(number) {
		return nil, ErrInvalidOrderNumberFormat
	}

	// Создаем заказ
	order, err := s.orderRepo.Create(ctx, number, userID)
	if err != nil {
		return order, err
	}

	return order, nil
}

// GetUserOrders получает все заказы пользователя
func (s *OrderServiceImpl) GetUserOrders(ctx context.Context, userID int) ([]model.Order, error) {
	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	return orders, nil
}

// ValidateOrderNumber проверяет номер заказа на соответствие алгоритму Луна
func (s *OrderServiceImpl) ValidateOrderNumber(number string) bool {
	return validation.ValidateOrderNumber(number)
}
