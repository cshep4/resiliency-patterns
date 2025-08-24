package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// PaymentProcessor defines the interface for payment processing operations
type PaymentProcessor interface {
	ProcessPayment(ctx context.Context, request PaymentRequest) (PaymentResponse, error)
}

// HealthToggler defines interface for toggling service health
type HealthToggler interface {
	SetHealthy(healthy bool)
	IsHealthy() bool
}

// ControllablePaymentService combines payment processing with health control
type ControllablePaymentService interface {
	HealthToggler
	ProcessPayment(ctx context.Context, request PaymentRequest) (PaymentResponse, error)
	SetFailureRate(rate float64) error
}

// PaymentRequest represents a payment processing request
type PaymentRequest struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	MerchantID string `json:"merchant_id"`
	CardToken  string `json:"card_token"`
}

// PaymentResponse represents a payment processing response
type PaymentResponse struct {
	ID              string    `json:"id"`
	TransactionID   string    `json:"transaction_id"`
	Status          string    `json:"status"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	ProcessedAt     time.Time `json:"processed_at"`
	ProcessingTime  time.Duration `json:"processing_time"`
}

// paymentService simulates an external payment processing service
type paymentService struct {
	delay       time.Duration
	failureRate float64 // 0.0 to 1.0 probability of failure
	isHealthy   bool
}

// NewPaymentService creates a new payment service
func NewPaymentService(delay time.Duration, failureRate float64) (*paymentService, error) {
	if delay < 0 {
		return nil, errors.New("delay must be greater than or equal to 0")
	}
	if failureRate < 0 || failureRate > 1 {
		return nil, errors.New("failure rate must be between 0 and 1")
	}

	return &paymentService{
		delay:       delay,
		failureRate: failureRate,
		isHealthy:   true,
	}, nil
}

// ProcessPayment processes a payment request
func (s *paymentService) ProcessPayment(ctx context.Context, request PaymentRequest) (PaymentResponse, error) {
	start := time.Now()

	// Simulate network delay
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return PaymentResponse{}, ctx.Err()
	}

	// Simulate random failures based on failure rate
	if !s.isHealthy || rand.Float64() < s.failureRate {
		return PaymentResponse{}, fmt.Errorf("payment service unavailable for request %s", request.ID)
	}

	// Validate request
	if err := s.validateRequest(request); err != nil {
		return PaymentResponse{}, fmt.Errorf("invalid payment request: %w", err)
	}

	// Create successful response
	response := PaymentResponse{
		ID:             request.ID,
		TransactionID:  uuid.New().String(),
		Status:         "completed",
		Amount:         request.Amount,
		Currency:       request.Currency,
		ProcessedAt:    time.Now(),
		ProcessingTime: time.Since(start),
	}

	return response, nil
}

// validateRequest validates the payment request
func (s *paymentService) validateRequest(request PaymentRequest) error {
	if request.ID == "" {
		return errors.New("payment ID is required")
	}
	if request.Amount <= 0 {
		return errors.New("payment amount must be greater than 0")
	}
	if request.Currency == "" {
		return errors.New("currency is required")
	}
	if request.MerchantID == "" {
		return errors.New("merchant ID is required")
	}
	if request.CardToken == "" {
		return errors.New("card token is required")
	}
	return nil
}

// SetHealthy sets the health status of the service
func (s *paymentService) SetHealthy(healthy bool) {
	s.isHealthy = healthy
}

// IsHealthy returns the current health status
func (s *paymentService) IsHealthy() bool {
	return s.isHealthy
}

// SetFailureRate updates the failure rate
func (s *paymentService) SetFailureRate(rate float64) error {
	if rate < 0 || rate > 1 {
		return errors.New("failure rate must be between 0 and 1")
	}
	s.failureRate = rate
	return nil
}