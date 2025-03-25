package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alexuryumtsev/gophermart/config"
	"github.com/alexuryumtsev/gophermart/internal/accrual"
	"github.com/alexuryumtsev/gophermart/internal/app"
	"github.com/alexuryumtsev/gophermart/internal/auth"
	"github.com/alexuryumtsev/gophermart/internal/handler"
	"github.com/alexuryumtsev/gophermart/internal/middleware"
	"github.com/alexuryumtsev/gophermart/internal/repository"
	"github.com/alexuryumtsev/gophermart/internal/router"
	"github.com/alexuryumtsev/gophermart/internal/service"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.NewConfig()

	// Проверяем, что все обязательные параметры указаны
	if cfg.DatabaseURI == "" {
		log.Fatal("Database URI is required")
	}

	if cfg.AccrualSystemAddress == "" {
		log.Fatal("Accrual system address is required")
	}

	// Подключаемся к базе данных
	db, err := repository.NewDB(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализируем схему базы данных
	if err := repository.InitDB(db); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Создаем JWTManager с секретным ключом из конфигурации
	jwtManager := auth.NewJWTManager(cfg.JWTSecretKey)

	// Инициализируем репозитории
	repos := repository.NewRepository(db)

	// Инициализируем клиент для системы начислений
	accrualClient := accrual.NewAccrualClient(cfg.AccrualSystemAddress)

	// Инициализируем сервисы
	authService := service.NewAuthService(jwtManager)
	services := service.NewService(repos, accrualClient, jwtManager)

	// Инициализируем middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Инициализируем обработчики HTTP
	userHandler := handler.NewUserHandler(services.Users)
	orderHandler := handler.NewOrderHandler(services.Orders)
	balanceHandler := handler.NewBalanceHandler(services.Balances)

	// Инициализируем сервис обработки заказов
	processor := app.NewAccrualProcessor(
		repos.Orders,
		repos.Balances,
		accrualClient,
		5*time.Second,
	)
	processor.Start()
	defer processor.Stop()

	// Инициализируем роутер
	r := router.NewRouter(userHandler, orderHandler, balanceHandler, authMiddleware)

	// Запускаем сервер
	log.Printf("Starting server on %s", cfg.RunAddress)
	if err := http.ListenAndServe(cfg.RunAddress, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
