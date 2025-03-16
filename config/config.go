package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func NewConfig() *Config {
	var cfg Config

	// Парсим флаги
	flag.StringVar(&cfg.RunAddress, "a", ":8081", "Address and port to run server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")
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

	return &cfg
}
