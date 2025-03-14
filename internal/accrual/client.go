// accrual/client.go
package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexuryumtsev/gophermart/internal/model"
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

func (c *AccrualClient) GetOrderAccrual(orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, err
		}
		return &accrualResp, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		// Можно добавить логику ретраев здесь
		return nil, fmt.Errorf("rate limit exceeded")
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

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
