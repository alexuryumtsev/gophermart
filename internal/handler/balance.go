// handler/balance.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/alexuryumtsev/gophermart/internal/middleware"
	"github.com/alexuryumtsev/gophermart/internal/model"
	"github.com/alexuryumtsev/gophermart/internal/repository"
)

type BalanceHandler struct {
	balanceRepo *repository.BalanceRepository
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func NewBalanceHandler(balanceRepo *repository.BalanceRepository) *BalanceHandler {
	return &BalanceHandler{
		balanceRepo: balanceRepo,
	}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	balance, err := h.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

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

	// Проверяем валидность номера заказа
	if _, err := strconv.ParseInt(req.Order, 10, 64); err != nil {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Проверяем валидность по алгоритму Луна
	if !validateLuhn(req.Order) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()

	// Проверяем достаточно ли средств
	balance, err := h.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if balance.Current < req.Sum {
		http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		return
	}

	// Создаем транзакцию списания
	_, err = h.balanceRepo.AddTransaction(
		ctx,
		userID,
		req.Order,
		req.Sum,
		model.OperationTypeWithdrawal,
	)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	withdrawals, err := h.balanceRepo.GetWithdrawals(ctx, userID)
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
