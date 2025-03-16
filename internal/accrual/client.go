// internal/accrual/client.go
package accrual

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alexuryumtsev/gophermart/internal/model"
)

// настройки для ретраев
const (
	maxRetries     = 3   // Максимальное количество повторных попыток
	defaultBackoff = 60  // Время ожидания по умолчанию в секундах
	maxBackoff     = 300 // Максимальное время ожидания в секундах
)

type AccrualClient struct {
	baseURL    string
	httpClient *http.Client
}

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetOrderAccrual делает запрос к системе расчета начислений для получения информации по заказу
// с автоматическими повторными попытками в случае ошибок превышения лимита запросов
func (c *AccrualClient) GetOrderAccrual(orderNumber string) (*AccrualResponse, error) {
	var lastErr error
	retryCount := 0

	// Цикл для повторных попыток
	for retryCount <= maxRetries {
		// Если это не первая попытка, выводим информацию о повторе
		if retryCount > 0 {
			log.Printf("Retry %d/%d for order %s", retryCount, maxRetries, orderNumber)
		}

		// Делаем запрос к API
		response, err, shouldRetry, waitTime := c.makeRequest(orderNumber)

		// Если нет ошибки или не нужно повторять запрос, возвращаем результат
		if err == nil || !shouldRetry {
			return response, err
		}

		// Сохраняем последнюю ошибку для возможного возврата
		lastErr = err

		// Если достигнуто максимальное количество повторов, завершаем
		if retryCount >= maxRetries {
			break
		}

		// Увеличиваем счетчик повторов
		retryCount++

		// Ждем перед следующей попыткой
		log.Printf("Waiting %d seconds before next retry for order %s", waitTime, orderNumber)
		time.Sleep(time.Duration(waitTime) * time.Second)
	}

	// Если все попытки неудачны, возвращаем последнюю ошибку
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// makeRequest выполняет запрос к системе начислений
// Возвращает ответ, ошибку, флаг необходимости повтора и время ожидания
func (c *AccrualClient) makeRequest(orderNumber string) (*AccrualResponse, bool, int, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Сетевые ошибки обычно временные, повторяем запрос
		return nil, true, 5, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, false, 0, fmt.Errorf("failed to decode response: %w", err)
		}
		return &accrualResp, false, 0, nil

	case http.StatusNoContent:
		return nil, false, 0, nil

	case http.StatusTooManyRequests:
		// Извлекаем время ожидания из заголовка
		retryAfter := resp.Header.Get("Retry-After")
		waitTime := defaultBackoff

		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
				waitTime = seconds

				// Ограничиваем максимальное время ожидания
				if waitTime > maxBackoff {
					waitTime = maxBackoff
				}
			}
		}

		return nil, true, waitTime, fmt.Errorf("rate limit exceeded, retry after %d seconds", waitTime)

	case http.StatusInternalServerError:
		// Внутренняя ошибка сервера, возможно временная, повторяем запрос
		return nil, true, 10, fmt.Errorf("server error: %d", resp.StatusCode)

	default:
		// Другие ошибки не повторяем
		return nil, false, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// MapStatusToOrderStatus конвертирует статус из системы начислений в статус нашей системы
func (c *AccrualClient) MapStatusToOrderStatus(status string) model.OrderStatus {
	switch status {
	case "REGISTERED":
		return model.OrderStatusProcessing
	case "PROCESSING":
		return model.OrderStatusProcessing
	case "INVALID":
		return model.OrderStatusInvalid
	case "PROCESSED":
		return model.OrderStatusProcessed
	default:
		return model.OrderStatusNew
	}
}
