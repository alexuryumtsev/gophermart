package app

import (
	"context"
	"log"
	"time"

	"github.com/alexuryumtsev/gophermart/internal/accrual"
	"github.com/alexuryumtsev/gophermart/internal/model"
	"github.com/alexuryumtsev/gophermart/internal/repository"
)

type AccrualProcessor struct {
	orderRepo     *repository.OrderRepository
	balanceRepo   *repository.BalanceRepository
	accrualClient *accrual.AccrualClient
	interval      time.Duration
	done          chan struct{}
}

func NewAccrualProcessor(
	orderRepo *repository.OrderRepository,
	balanceRepo *repository.BalanceRepository,
	accrualClient *accrual.AccrualClient,
	interval time.Duration,
) *AccrualProcessor {
	return &AccrualProcessor{
		orderRepo:     orderRepo,
		balanceRepo:   balanceRepo,
		accrualClient: accrualClient,
		interval:      interval,
		done:          make(chan struct{}),
	}
}

func (p *AccrualProcessor) Start() {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.processOrders()
			case <-p.done:
				return
			}
		}
	}()
}

func (p *AccrualProcessor) Stop() {
	close(p.done)
}

func (p *AccrualProcessor) processOrders() {
	ctx := context.Background()

	// Получаем заказы для обработки
	orders, err := p.orderRepo.GetOrdersForProcessing(ctx)
	if err != nil {
		log.Printf("Error getting orders for processing: %v", err)
		return
	}

	for _, order := range orders {
		// Делаем запрос к системе начислений
		accrualResp, err := p.accrualClient.GetOrderAccrual(order.Number)
		if err != nil {
			log.Printf("Error getting accrual for order %s: %v", order.Number, err)
			// Если достигнут лимит запросов, прерываем обработку
			if err.Error() == "rate limit exceeded" {
				break
			}
			continue
		}

		if accrualResp == nil {
			continue
		}

		// Обновляем статус заказа
		newStatus := p.accrualClient.MapStatusToOrderStatus(accrualResp.Status)
		err = p.orderRepo.UpdateStatus(ctx, order.Number, newStatus, accrualResp.Accrual)
		if err != nil {
			log.Printf("Error updating order status: %v", err)
			continue
		}

		// Если заказ обработан и есть начисление, добавляем транзакцию в баланс
		if newStatus == model.OrderStatusProcessed && accrualResp.Accrual != nil && *accrualResp.Accrual > 0 {
			_, err = p.balanceRepo.AddTransaction(
				ctx,
				order.UserID,
				order.Number,
				*accrualResp.Accrual,
				model.OperationTypeAccrual,
			)
			if err != nil {
				log.Printf("Error adding accrual transaction: %v", err)
			}
		}
	}
}
