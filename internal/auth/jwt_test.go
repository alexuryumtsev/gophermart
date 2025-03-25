package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	// Создаем экземпляр JWTManager с тестовым секретным ключом
	secretKey := "test-secret-key"
	manager := NewJWTManager(secretKey)

	tests := []struct {
		name    string
		userID  int
		wantErr bool
	}{
		{
			name:    "Valid user ID",
			userID:  123,
			wantErr: false,
		},
		{
			name:    "Zero user ID",
			userID:  0,
			wantErr: false, // Даже для нулевого ID токен должен генерироваться
		},
		{
			name:    "Negative user ID",
			userID:  -1,
			wantErr: false, // Для тестирования граничного случая
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Генерируем токен
			token, err := manager.GenerateToken(tt.userID)

			// Проверяем наличие ошибки
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Проверяем, что токен не пустой
			assert.NotEmpty(t, token)

			// Дополнительно проверяем структуру токена
			parsedToken, err := jwt.ParseWithClaims(
				token,
				&Claims{},
				func(token *jwt.Token) (interface{}, error) {
					return []byte(secretKey), nil
				},
			)
			require.NoError(t, err)

			// Проверяем, что токен валидный
			assert.True(t, parsedToken.Valid)

			// Проверяем содержимое claims
			claims, ok := parsedToken.Claims.(*Claims)
			assert.True(t, ok)
			assert.Equal(t, tt.userID, claims.UserID)

			// Проверяем время жизни токена
			expiresAt := claims.ExpiresAt
			assert.Greater(t, expiresAt.Unix(), time.Now().Unix())
			assert.LessOrEqual(t, expiresAt.Unix(), time.Now().Add(25*time.Hour).Unix())
		})
	}
}

func TestJWTManager_ParseToken(t *testing.T) {
	// Создаем экземпляр JWTManager с тестовым секретным ключом
	secretKey := "test-secret-key"
	manager := NewJWTManager(secretKey)

	// Генерируем валидный токен для тестирования
	validUserID := 123
	validToken, err := manager.GenerateToken(validUserID)
	require.NoError(t, err)

	// Создаем просроченный токен
	expiredToken := createExpiredToken(t, secretKey, 456)

	// Создаем токен с неправильной подписью
	invalidSignatureToken := createTokenWithInvalidSignature(t, "wrong-secret-key", 789)

	tests := []struct {
		name       string
		token      string
		wantUserID int
		wantErr    bool
	}{
		{
			name:       "Valid token",
			token:      validToken,
			wantUserID: validUserID,
			wantErr:    false,
		},
		{
			name:       "Empty token",
			token:      "",
			wantUserID: 0,
			wantErr:    true,
		},
		{
			name:       "Invalid token format",
			token:      "invalid.token.format",
			wantUserID: 0,
			wantErr:    true,
		},
		{
			name:       "Expired token",
			token:      expiredToken,
			wantUserID: 0,
			wantErr:    true,
		},
		{
			name:       "Invalid signature",
			token:      invalidSignatureToken,
			wantUserID: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Парсим токен
			userID, err := manager.ParseToken(tt.token)

			// Проверяем наличие ошибки
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Проверяем извлеченный ID пользователя
			assert.Equal(t, tt.wantUserID, userID)
		})
	}
}

// Вспомогательная функция для создания просроченного токена
func createExpiredToken(t *testing.T, secretKey string, userID int) string {
	t.Helper()

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Токен просрочен
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	return tokenString
}

// Вспомогательная функция для создания токена с неправильной подписью
func createTokenWithInvalidSignature(t *testing.T, wrongSecretKey string, userID int) string {
	t.Helper()

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(wrongSecretKey))
	require.NoError(t, err)

	return tokenString
}

// Тест для проверки обработки разных алгоритмов подписи
func TestJWTManager_DifferentAlgorithms(t *testing.T) {
	// Создаем экземпляр JWTManager с тестовым секретным ключом
	secretKey := "test-secret-key"
	manager := NewJWTManager(secretKey)

	// Создаем токен с другим алгоритмом (например, HS512)
	claims := &Claims{
		UserID: 999,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	// Проверяем, что такой токен не принимается (из-за WithValidMethods в ParseToken)
	userID, err := manager.ParseToken(tokenString)
	assert.Error(t, err)
	assert.Zero(t, userID)
}

// Тест для проверки безопасности от атаки None алгоритма
func TestJWTManager_NoneAlgorithmAttack(t *testing.T) {
	// Создаем экземпляр JWTManager с тестовым секретным ключом
	secretKey := "test-secret-key"
	manager := NewJWTManager(secretKey)

	// Создаем токен с алгоритмом "none" (это имитация потенциальной атаки)
	// В реальной атаке злоумышленник модифицирует заголовок токена,
	// но мы просто проверяем, что наша реализация защищена
	claims := &Claims{
		UserID: 777,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	// Создаем токен без подписи
	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := unsignedToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Проверяем, что токен с алгоритмом "none" отвергается
	userID, err := manager.ParseToken(tokenString)
	assert.Error(t, err)
	assert.Zero(t, userID)
}
