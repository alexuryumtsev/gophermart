// internal/service/interfaces.go
package service

import (
	"context"

	"github.com/alexuryumtsev/gophermart/internal/auth"
	"github.com/alexuryumtsev/gophermart/internal/model"
	"github.com/alexuryumtsev/gophermart/internal/repository"
)

// AuthService определяет бизнес-логику для аутентификации
type AuthService interface {
	// GenerateToken создает JWT токен для пользователя
	GenerateToken(userID int) (string, error)

	// ParseToken проверяет и извлекает данные из JWT токена
	ParseToken(tokenString string) (int, error)
}

// UserService определяет бизнес-логику для работы с пользователями
type UserService interface {
	// Register регистрирует нового пользователя
	Register(ctx context.Context, login, password string) (*model.User, error)

	// Authenticate проверяет учетные данные пользователя
	Authenticate(ctx context.Context, login, password string) (*model.User, error)

	// GenerateToken создает JWT токен для пользователя
	GenerateToken(userID int) (string, error)
}

// OrderService определяет бизнес-логику для работы с заказами
type OrderService interface {
	// CreateOrder создает новый заказ
	CreateOrder(ctx context.Context, number string, userID int) (*model.Order, error)

	// GetUserOrders получает все заказы пользователя
	GetUserOrders(ctx context.Context, userID int) ([]model.Order, error)

	// ValidateOrderNumber проверяет номер заказа на соответствие алгоритму Луна
	ValidateOrderNumber(number string) bool
}

// BalanceService определяет бизнес-логику для работы с балансом
type BalanceService interface {
	// GetUserBalance получает текущий баланс пользователя
	GetUserBalance(ctx context.Context, userID int) (*model.Balance, error)

	// WithdrawFunds списывает средства с баланса пользователя
	WithdrawFunds(ctx context.Context, userID int, orderNumber string, amount float64) error

	// GetUserWithdrawals получает историю списаний пользователя
	GetUserWithdrawals(ctx context.Context, userID int) ([]model.BalanceTransaction, error)
}

// Service объединяет все сервисы для удобства внедрения зависимостей
type Service struct {
	Users    UserService
	Orders   OrderService
	Balances BalanceService
	Auth     AuthService
}

// NewService создает экземпляр сервиса с конкретными реализациями
func NewService(repos *repository.Repository, accrualClient repository.AccrualClient, jwtManager *auth.JWTManager) *Service {
	return &Service{
		Users:    NewUserService(repos.Users, NewAuthService(jwtManager)),
		Orders:   NewOrderService(repos.Orders),
		Balances: NewBalanceService(repos.Balances),
		Auth:     NewAuthService(jwtManager), // Инициализируем AuthService
	}
}
