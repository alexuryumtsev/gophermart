package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexuryumtsev/gophermart/internal/middleware"
	"github.com/alexuryumtsev/gophermart/internal/service"
)

type BalanceHandler struct {
	balanceService service.BalanceService
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func NewBalanceHandler(balanceService service.BalanceService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

// GetBalance обрабатывает запрос на получение текущего баланса пользователя
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Используем сервис для получения баланса пользователя
	balance, err := h.balanceService.GetUserBalance(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// Withdraw обрабатывает запрос на списание средств с баланса
func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Используем сервис для списания средств (с внутренней валидацией)
	err := h.balanceService.WithdrawFunds(ctx, userID, req.Order, req.Sum)
	if err != nil {
		switch err {
		case service.ErrInvalidOrderFormat:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case service.ErrInsufficientFunds:
			http.Error(w, err.Error(), http.StatusPaymentRequired)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals обрабатывает запрос на получение истории списаний пользователя
func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Используем сервис для получения истории списаний
	withdrawals, err := h.balanceService.GetUserWithdrawals(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type WithdrawalResponse struct {
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}

	var response []WithdrawalResponse
	for _, tx := range withdrawals {
		resp := WithdrawalResponse{
			Order:       tx.OrderNumber,
			Sum:         tx.Amount,
			ProcessedAt: tx.ProcessedAt.Format(time.RFC3339),
		}
		response = append(response, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
