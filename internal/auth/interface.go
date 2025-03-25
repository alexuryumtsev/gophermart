package auth

type JWTManagerInterface interface {
	// GenerateToken создает новый JWT токен для пользователя
	GenerateToken(userID int) (string, error)

	// ParseToken проверяет и извлекает данные из JWT токена
	ParseToken(tokenString string) (int, error)
}
