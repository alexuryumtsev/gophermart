package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/alexuryumtsev/gophermart/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// Тестовый обработчик
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что контекст содержит ID пользователя
		userID, ok := GetUserID(r.Context())
		if !ok {
			t.Error("Failed to get user ID from context")
		}

		// Пишем ID пользователя в ответ
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("userID: " + strconv.Itoa(userID)))
	})

	// Создаем middleware
	middleware := AuthMiddleware(nextHandler)

	// Тест с токеном в заголовке Authorization
	t.Run("Token in Authorization header", func(t *testing.T) {
		// Генерируем валидный токен
		token, err := auth.GenerateToken(123)
		assert.NoError(t, err)

		// Создаем запрос с токеном в заголовке
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Создаем рекордер для записи ответа
		rec := httptest.NewRecorder()

		// Вызываем middleware
		middleware.ServeHTTP(rec, req)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Тест с токеном в cookie
	t.Run("Token in cookie", func(t *testing.T) {
		// Генерируем валидный токен
		token, err := auth.GenerateToken(456)
		assert.NoError(t, err)

		// Создаем запрос с токеном в cookie
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "auth_token",
			Value: token,
		})

		// Создаем рекордер для записи ответа
		rec := httptest.NewRecorder()

		// Вызываем middleware
		middleware.ServeHTTP(rec, req)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Тест без токена
	t.Run("No token", func(t *testing.T) {
		// Создаем запрос без токена
		req := httptest.NewRequest("GET", "/", nil)

		// Создаем рекордер для записи ответа
		rec := httptest.NewRecorder()

		// Вызываем middleware
		middleware.ServeHTTP(rec, req)

		// Должны получить 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	// Тест с невалидным токеном
	t.Run("Invalid token", func(t *testing.T) {
		// Создаем запрос с невалидным токеном
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")

		// Создаем рекордер для записи ответа
		rec := httptest.NewRecorder()

		// Вызываем middleware
		middleware.ServeHTTP(rec, req)

		// Должны получить 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestGetUserID(t *testing.T) {
	// Тест получения ID из контекста
	t.Run("Get user ID from context", func(t *testing.T) {
		expectedID := 789
		ctx := context.WithValue(context.Background(), UserIDKey, expectedID)

		userID, ok := GetUserID(ctx)
		assert.True(t, ok)
		assert.Equal(t, expectedID, userID)
	})

	// Тест с отсутствующим ID в контексте
	t.Run("No user ID in context", func(t *testing.T) {
		ctx := context.Background()

		userID, ok := GetUserID(ctx)
		assert.False(t, ok)
		assert.Equal(t, 0, userID)
	})

	// Тест с неверным типом в контексте
	t.Run("Wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, "not_an_int")

		userID, ok := GetUserID(ctx)
		assert.False(t, ok)
		assert.Equal(t, 0, userID)
	})
}
