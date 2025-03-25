package accrual

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexuryumtsev/gophermart/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestAccrualClient_GetOrderAccrual(t *testing.T) {
	// Тест успешного получения информации о начислении
	t.Run("Successful accrual info", func(t *testing.T) {
		// Создаем тестовый сервер, который будет имитировать систему начислений
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем путь запроса
			if r.URL.Path != "/api/orders/12345" {
				t.Errorf("Expected path /api/orders/12345, got %s", r.URL.Path)
			}

			// Возвращаем успешный ответ
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"order":"12345","status":"PROCESSED","accrual":500}`))
		}))
		defer server.Close()

		// Создаем клиент с адресом тестового сервера
		client := NewAccrualClient(server.URL)

		// Вызываем метод
		resp, err := client.GetOrderAccrual("12345")

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "12345", resp.Order)
		assert.Equal(t, "PROCESSED", resp.Status)
		assert.NotNil(t, resp.Accrual)
		assert.Equal(t, 500.0, *resp.Accrual)
	})

	// Тест отсутствия заказа в системе
	t.Run("Order not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		resp, err := client.GetOrderAccrual("nonexistent")

		assert.NoError(t, err)
		assert.Nil(t, resp)
	})

	// Тест превышения лимита запросов
	t.Run("Rate limit exceeded", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("No more than N requests per minute allowed"))
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		resp, err := client.GetOrderAccrual("12345")

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})

	// Тест ошибки сервера
	t.Run("Server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewAccrualClient(server.URL)
		resp, err := client.GetOrderAccrual("12345")

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "max retries exceeded: server error: 500")
	})
}

func TestAccrualClient_MapStatusToOrderStatus(t *testing.T) {
	client := NewAccrualClient("http://example.com")

	// Тестируем преобразование всех возможных статусов
	testCases := []struct {
		name     string
		status   string
		expected model.OrderStatus
	}{
		{"REGISTERED status", "REGISTERED", model.OrderStatusProcessing},
		{"PROCESSING status", "PROCESSING", model.OrderStatusProcessing},
		{"INVALID status", "INVALID", model.OrderStatusInvalid},
		{"PROCESSED status", "PROCESSED", model.OrderStatusProcessed},
		{"Unknown status", "UNKNOWN", model.OrderStatusNew},
		{"Empty status", "", model.OrderStatusNew},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.MapStatusToOrderStatus(tc.status)
			assert.Equal(t, tc.expected, result)
		})
	}
}
