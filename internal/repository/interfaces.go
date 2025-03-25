package repository

import (
	"context"

	"github.com/alexuryumtsev/gophermart/internal/model"
)

// UserRepository определяет операции для работы с пользователями
type UserRepository interface {
	// Create создает нового пользователя
	Create(ctx context.Context, login, password string) (*model.User, error)

	// GetByLogin находит пользователя по логину
	GetByLogin(ctx context.Context, login string) (*model.User, error)

	// VerifyPassword проверяет правильность пароля для пользователя
	VerifyPassword(user *model.User, password string) bool
}

// OrderRepository определяет операции для работы с заказами
type OrderRepository interface {
	// Create создает новый заказ
	Create(ctx context.Context, number string, userID int) (*model.Order, error)

	// GetByNumber находит заказ по номеру
	GetByNumber(ctx context.Context, number string) (*model.Order, error)

	// GetByUserID возвращает все заказы пользователя
	GetByUserID(ctx context.Context, userID int) ([]model.Order, error)

	// UpdateStatus обновляет статус заказа и информацию о начислении
	UpdateStatus(ctx context.Context, number string, status model.OrderStatus, accrual *float64) error

	// GetOrdersForProcessing получает заказы, требующие обработки
	GetOrdersForProcessing(ctx context.Context) ([]model.Order, error)
}

// BalanceRepository определяет операции для работы с балансом
type BalanceRepository interface {
	// GetBalance получает текущий баланс пользователя
	GetBalance(ctx context.Context, userID int) (*model.Balance, error)

	// AddTransaction добавляет транзакцию в истории баланса
	AddTransaction(
		ctx context.Context,
		userID int,
		orderNumber string,
		amount float64,
		operationType model.OperationType,
	) (*model.BalanceTransaction, error)

	// GetWithdrawals получает историю списаний пользователя
	GetWithdrawals(ctx context.Context, userID int) ([]model.BalanceTransaction, error)
}

// AccrualClient определяет интерфейс для взаимодействия с системой начислений
type AccrualClient interface {
	// GetOrderAccrual получает информацию о начислении для заказа
	GetOrderAccrual(orderNumber string) (*AccrualResponse, error)

	// MapStatusToOrderStatus конвертирует статус из системы начислений в статус заказа
	MapStatusToOrderStatus(status string) model.OrderStatus
}

// AccrualResponse представляет ответ от системы начислений
type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

// Repository объединяет все репозитории для удобства внедрения зависимостей
type Repository struct {
	Users    UserRepository
	Orders   OrderRepository
	Balances BalanceRepository
}

// NewRepository создает экземпляр репозитория с конкретными реализациями
func NewRepository(db PgxPool) *Repository {
	return &Repository{
		Users:    NewUserRepository(db),
		Orders:   NewOrderRepository(db),
		Balances: NewBalanceRepository(db),
	}
}
