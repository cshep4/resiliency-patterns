package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// OrderRequest represents an order processing request
type OrderRequest struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Items    []Item  `json:"items"`
}

// Item represents an order item
type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// OrderResponse represents an order processing response
type OrderResponse struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"order_id"`
	Status      string    `json:"status"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	ProcessedAt time.Time `json:"processed_at"`
}

// orderService simulates an external order processing service
type orderService struct {
	failureRate float64
	delay       time.Duration
}

// NewOrderService creates a new order service
func NewOrderService(delay time.Duration, failureRate float64) (*orderService, error) {
	if delay < 0 {
		return nil, errors.New("delay must be greater than or equal to 0")
	}
	if failureRate < 0 || failureRate > 1 {
		return nil, errors.New("failure rate must be between 0 and 1")
	}

	return &orderService{
		failureRate: failureRate,
		delay:       delay,
	}, nil
}

// ProcessOrder processes an order request
func (s *orderService) ProcessOrder(ctx context.Context, request OrderRequest) (OrderResponse, error) {
	// Simulate network delay
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return OrderResponse{}, ctx.Err()
	}

	// Simulate failures
	if rand.Float64() < s.failureRate {
		return OrderResponse{}, fmt.Errorf("order processing failed: service unavailable for order %s", request.ID)
	}

	// Create successful response
	response := OrderResponse{
		ID:          request.ID,
		OrderID:     uuid.New().String(),
		Status:      "completed",
		Amount:      request.Amount,
		Currency:    request.Currency,
		ProcessedAt: time.Now(),
	}

	return response, nil
}

// SetFailureRate updates the failure rate
func (s *orderService) SetFailureRate(rate float64) error {
	if rate < 0 || rate > 1 {
		return errors.New("failure rate must be between 0 and 1")
	}
	s.failureRate = rate
	return nil
}
