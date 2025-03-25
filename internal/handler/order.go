package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexuryumtsev/gophermart/internal/middleware"
	"github.com/alexuryumtsev/gophermart/internal/repository"
	"github.com/alexuryumtsev/gophermart/internal/service"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// UploadOrder обрабатывает запрос на загрузку номера заказа
func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Читаем номер заказа из тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "Invalid order number format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Используем сервис для создания заказа
	order, err := h.orderService.CreateOrder(ctx, orderNumber, userID)

	if order != nil && errors.Is(err, repository.ErrOrderIsExist) {
		http.Error(w, "Order already uploaded by another user", http.StatusConflict)
		return
	}

	if order != nil && errors.Is(err, repository.ErrOrderUserIsExist) {
		w.WriteHeader(http.StatusOK)
		return
	}

	if order == nil && errors.Is(err, service.ErrInvalidOrderNumberFormat) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if order == nil && err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Новый заказ был принят в обработку
	w.WriteHeader(http.StatusAccepted)
}

// GetOrders обрабатывает запрос на получение списка заказов пользователя
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Используем сервис для получения заказов пользователя
	orders, err := h.orderService.GetUserOrders(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type OrderResponse struct {
		Number     string   `json:"number"`
		Status     string   `json:"status"`
		Accrual    *float64 `json:"accrual,omitempty"`
		UploadedAt string   `json:"uploaded_at"`
	}

	var response []OrderResponse
	for _, order := range orders {
		resp := OrderResponse{
			Number:     order.Number,
			Status:     string(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}
		response = append(response, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
