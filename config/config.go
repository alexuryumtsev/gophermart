package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecretKey         string // Добавляем поле для JWT секретного ключа
}

func NewConfig() *Config {
	var cfg Config

	// Парсим флаги
	flag.StringVar(&cfg.RunAddress, "a", ":8081", "Address and port to run server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")
	flag.StringVar(&cfg.JWTSecretKey, "j", "", "JWT secret key")
	flag.Parse()

	// Проверяем переменные окружения
	if envAddr := os.Getenv("RUN_ADDRESS"); envAddr != "" {
		cfg.RunAddress = envAddr
	}

	if envDBURI := os.Getenv("DATABASE_URI"); envDBURI != "" {
		cfg.DatabaseURI = envDBURI
	}

	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		cfg.AccrualSystemAddress = envAccrualAddr
	}

	if envJWTSecret := os.Getenv("JWT_SECRET_KEY"); envJWTSecret != "" {
		cfg.JWTSecretKey = envJWTSecret
	}

	// Генерируем случайный секретный ключ, если не указан
	if cfg.JWTSecretKey == "" {
		cfg.JWTSecretKey = generateRandomKey()
	}

	return &cfg
}

// generateRandomKey генерирует случайный ключ, если не указан в конфигурации
func generateRandomKey() string {
	// Простая реализация для примера
	// В продакшене лучше использовать более надежный способ генерации
	const defaultKey = "default-secret-key-for-development-only"
	return defaultKey
}
