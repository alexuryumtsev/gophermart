package router

import (
	"github.com/alexuryumtsev/gophermart/internal/handler"
	customMiddleware "github.com/alexuryumtsev/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter создает и настраивает новый роутер с необходимыми маршрутами и middleware
func NewRouter(
	userHandler *handler.UserHandler,
	orderHandler *handler.OrderHandler,
	balanceHandler *handler.BalanceHandler,
	authMiddleware *customMiddleware.AuthMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// Маршруты без аутентификации
	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)

	// Маршруты, требующие аутентификации
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware) // Обновленный middleware

		r.Post("/api/user/orders", orderHandler.UploadOrder)
		r.Get("/api/user/orders", orderHandler.GetOrders)
		r.Get("/api/user/balance", balanceHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", balanceHandler.Withdraw)
		r.Get("/api/user/withdrawals", balanceHandler.GetWithdrawals)
	})

	return r
}
