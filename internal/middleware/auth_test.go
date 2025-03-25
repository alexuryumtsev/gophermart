package middleware

import (
	"errors"
	"testing"

	"github.com/alexuryumtsev/gophermart/internal/service"
	"github.com/alexuryumtsev/gophermart/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_GenerateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок для JWTManager
	mockJWTManager := mocks.NewMockJWTManagerInterface(ctrl)

	// Создаем сервис аутентификации с моком
	authService := service.AuthService(mockJWTManager)

	t.Run("Success", func(t *testing.T) {
		userID := 123
		expectedToken := "valid.token.here"

		// Настраиваем ожидаемый вызов GenerateToken
		mockJWTManager.EXPECT().
			GenerateToken(userID).
			Return(expectedToken, nil)

		// Вызываем тестируемый метод
		token, err := authService.GenerateToken(userID)

		// Проверяем результаты
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
	})

	t.Run("Error", func(t *testing.T) {
		userID := 123
		expectedError := errors.New("token generation failed")

		// Настраиваем ожидаемый вызов GenerateToken с ошибкой
		mockJWTManager.EXPECT().
			GenerateToken(userID).
			Return("", expectedError)

		// Вызываем тестируемый метод
		token, err := authService.GenerateToken(userID)

		// Проверяем результаты
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})
}

func TestAuthService_ParseToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок для JWTManager
	mockJWTManager := mocks.NewMockJWTManagerInterface(ctrl)

	// Создаем сервис аутентификации с моком
	authService := service.AuthService(mockJWTManager)

	t.Run("Valid token", func(t *testing.T) {
		tokenString := "valid.token.here"
		expectedUserID := 123

		// Настраиваем ожидаемый вызов ParseToken
		mockJWTManager.EXPECT().
			ParseToken(tokenString).
			Return(expectedUserID, nil)

		// Вызываем тестируемый метод
		userID, err := authService.ParseToken(tokenString)

		// Проверяем результаты
		assert.NoError(t, err)
		assert.Equal(t, expectedUserID, userID)
	})

	t.Run("Invalid token", func(t *testing.T) {
		tokenString := "invalid.token.here"
		expectedError := errors.New("invalid token")

		// Настраиваем ожидаемый вызов ParseToken с ошибкой
		mockJWTManager.EXPECT().
			ParseToken(tokenString).
			Return(0, expectedError)

		// Вызываем тестируемый метод
		userID, err := authService.ParseToken(tokenString)

		// Проверяем результаты
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, 0, userID)
	})

	t.Run("Expired token", func(t *testing.T) {
		tokenString := "expired.token.here"
		expectedError := errors.New("token expired")

		// Настраиваем ожидаемый вызов ParseToken с ошибкой истечения срока
		mockJWTManager.EXPECT().
			ParseToken(tokenString).
			Return(0, expectedError)

		// Вызываем тестируемый метод
		userID, err := authService.ParseToken(tokenString)

		// Проверяем результаты
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, 0, userID)
	})

	t.Run("Empty token", func(t *testing.T) {
		tokenString := ""
		expectedError := errors.New("token contains an invalid number of segments")

		// Настраиваем ожидаемый вызов ParseToken с ошибкой пустого токена
		mockJWTManager.EXPECT().
			ParseToken(tokenString).
			Return(0, expectedError)

		// Вызываем тестируемый метод
		userID, err := authService.ParseToken(tokenString)

		// Проверяем результаты
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, 0, userID)
	})
}
