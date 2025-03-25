package service

import (
	"github.com/alexuryumtsev/gophermart/internal/auth"
)

// AuthServiceImpl реализует интерфейс AuthService
type AuthServiceImpl struct {
	jwtManager *auth.JWTManager
}

// NewAuthService создает экземпляр AuthService
func NewAuthService(jwtManager *auth.JWTManager) AuthService {
	return &AuthServiceImpl{
		jwtManager: jwtManager,
	}
}

// GenerateToken создает JWT токен для пользователя
func (s *AuthServiceImpl) GenerateToken(userID int) (string, error) {
	return s.jwtManager.GenerateToken(userID)
}

// ParseToken проверяет и извлекает данные из JWT токена
func (s *AuthServiceImpl) ParseToken(tokenString string) (int, error) {
	return s.jwtManager.ParseToken(tokenString)
}
