// internal/service/user.go
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexuryumtsev/gophermart/internal/model"
	"github.com/alexuryumtsev/gophermart/internal/repository"
)

// UserServiceImpl реализует интерфейс UserService
type UserServiceImpl struct {
	userRepo repository.UserRepository
	auth     AuthService // Добавляем AuthService
}

// NewUserService создает экземпляр UserService
func NewUserService(userRepo repository.UserRepository, auth AuthService) UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
		auth:     auth,
	}
}

// Register регистрирует нового пользователя
func (s *UserServiceImpl) Register(ctx context.Context, login, password string) (*model.User, error) {
	// Проверяем, существует ли пользователь
	existingUser, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return nil, errors.New("login already taken")
	}

	// Создаем пользователя
	user, err := s.userRepo.Create(ctx, login, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Authenticate проверяет учетные данные пользователя
func (s *UserServiceImpl) Authenticate(ctx context.Context, login, password string) (*model.User, error) {
	// Получаем пользователя по логину
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil || !s.userRepo.VerifyPassword(user, password) {
		return nil, errors.New("invalid login/password pair")
	}

	return user, nil
}

// GenerateToken создает JWT токен для пользователя
func (s *UserServiceImpl) GenerateToken(userID int) (string, error) {
	return s.auth.GenerateToken(userID)
}
