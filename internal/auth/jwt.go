package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	tokenTTL = 24 * time.Hour
)

type JWTManager struct {
	secretKey []byte
}

// NewJWTManager создает новый экземпляр JWTManager
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
	}
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken создает новый JWT токен для пользователя
func (m *JWTManager) GenerateToken(userID int) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ParseToken проверяет и извлекает данные из JWT токена
func (m *JWTManager) ParseToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return m.secretKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	expTime := claims.ExpiresAt.Time

	if expTime.Before(time.Now()) {
		return 0, errors.New("token expired")
	}

	return claims.UserID, nil
}
