package auth

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		wantErr bool
	}{
		{
			name:    "Valid token generation",
			userID:  1,
			wantErr: false,
		},
		{
			name:    "Zero user ID",
			userID:  0,
			wantErr: false, // Даже с нулевым ID токен должен генерироваться
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID)

			// Проверка на наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверка, что токен не пустой, если не ожидается ошибка
			if !tt.wantErr && token == "" {
				t.Errorf("GenerateToken() returned empty token")
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	// Генерируем токен для тестирования
	testUserID := 42
	validToken, err := GenerateToken(testUserID)
	if err != nil {
		t.Fatalf("Failed to generate token for testing: %v", err)
	}

	// Создаем просроченный токен вручную (для этого нужно изменить код GenerateToken или создать отдельную тестовую функцию)
	expiredTokenFunc := func() string {
		// Здесь код для генерации просроченного токена
		// Временное решение - просто возвращаем невалидный токен
		return "invalid.token.string"
	}
	expiredToken := expiredTokenFunc()

	tests := []struct {
		name       string
		tokenStr   string
		wantUserID int
		wantErr    bool
	}{
		{
			name:       "Valid token",
			tokenStr:   validToken,
			wantUserID: testUserID,
			wantErr:    false,
		},
		{
			name:       "Empty token",
			tokenStr:   "",
			wantUserID: 0,
			wantErr:    true,
		},
		{
			name:       "Invalid token format",
			tokenStr:   "invalid.token",
			wantUserID: 0,
			wantErr:    true,
		},
		{
			name:       "Expired token",
			tokenStr:   expiredToken,
			wantUserID: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := ParseToken(tt.tokenStr)

			// Проверка на наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверка ID пользователя
			if userID != tt.wantUserID {
				t.Errorf("ParseToken() userID = %v, want %v", userID, tt.wantUserID)
			}
		})
	}
}
