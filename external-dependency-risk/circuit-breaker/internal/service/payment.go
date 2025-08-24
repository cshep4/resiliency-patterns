package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// PaymentRequest represents a payment processing request
type PaymentRequest struct {
	ID         string  `json:"id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	MerchantID string  `json:"merchant_id"`
	CardToken  string  `json:"card_token"`
}

// PaymentResponse represents a payment processing response
type PaymentResponse struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	ProcessedAt   time.Time `json:"processed_at"`
}

// paymentService simulates an external payment processing service
type paymentService struct {
	failureRate float64
	isHealthy   bool
}

// NewPaymentService creates a new payment service
func NewPaymentService(failureRate float64) (*paymentService, error) {
	if failureRate < 0 || failureRate > 1 {
		return nil, errors.New("failure rate must be between 0 and 1")
	}

	return &paymentService{
		failureRate: failureRate,
		isHealthy:   true,
	}, nil
}

// ProcessPayment processes a payment request
func (s *paymentService) ProcessPayment(ctx context.Context, request PaymentRequest) (PaymentResponse, error) {
	// Check health and simulate failures
	if !s.isHealthy || rand.Float64() < s.failureRate {
		return PaymentResponse{}, fmt.Errorf("payment processing failed: payment service unavailable for request %s", request.ID)
	}

	// Create successful response
	response := PaymentResponse{
		ID:            request.ID,
		TransactionID: uuid.New().String(),
		Status:        "completed",
		Amount:        request.Amount,
		Currency:      request.Currency,
		ProcessedAt:   time.Now(),
	}

	return response, nil
}

// SetHealthy sets the health status of the service
func (s *paymentService) SetHealthy(healthy bool) {
	s.isHealthy = healthy
}
