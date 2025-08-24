package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/circuit-breaker/internal/service"
)

func TestNewPaymentService(t *testing.T) {
	t.Run("valid parameters", func(t *testing.T) {
		svc, err := service.NewPaymentService(100*time.Millisecond, 0.1)
		require.NoError(t, err)
		require.NotNil(t, svc)
		require.True(t, svc.IsHealthy())
	})

	t.Run("negative delay", func(t *testing.T) {
		svc, err := service.NewPaymentService(-time.Second, 0.1)
		require.Error(t, err)
		require.Nil(t, svc)
		require.Contains(t, err.Error(), "delay must be greater than or equal to 0")
	})

	t.Run("invalid failure rate - negative", func(t *testing.T) {
		svc, err := service.NewPaymentService(100*time.Millisecond, -0.1)
		require.Error(t, err)
		require.Nil(t, svc)
		require.Contains(t, err.Error(), "failure rate must be between 0 and 1")
	})

	t.Run("invalid failure rate - too high", func(t *testing.T) {
		svc, err := service.NewPaymentService(100*time.Millisecond, 1.1)
		require.Error(t, err)
		require.Nil(t, svc)
		require.Contains(t, err.Error(), "failure rate must be between 0 and 1")
	})

	t.Run("zero delay and failure rate", func(t *testing.T) {
		svc, err := service.NewPaymentService(0, 0)
		require.NoError(t, err)
		require.NotNil(t, svc)
	})
}

func TestProcessPayment(t *testing.T) {
	t.Run("successful payment", func(t *testing.T) {
		svc, err := service.NewPaymentService(10*time.Millisecond, 0.0) // No failures
		require.NoError(t, err)

		ctx := context.Background()
		request := service.PaymentRequest{
			ID:        "test-payment-1",
			Amount:    100.50,
			Currency:  "USD",
			MerchantID: "merchant-123",
			CardToken:  "tok_test123",
		}

		response, err := svc.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, request.ID, response.ID)
		require.NotEmpty(t, response.TransactionID)
		require.Equal(t, "completed", response.Status)
		require.Equal(t, request.Amount, response.Amount)
		require.Equal(t, request.Currency, response.Currency)
		require.False(t, response.ProcessedAt.IsZero())
		require.Greater(t, response.ProcessingTime, time.Duration(0))
	})

	t.Run("unhealthy service", func(t *testing.T) {
		svc, err := service.NewPaymentService(10*time.Millisecond, 0.0)
		require.NoError(t, err)
		
		svc.SetHealthy(false)
		require.False(t, svc.IsHealthy())

		ctx := context.Background()
		request := service.PaymentRequest{
			ID:        "test-payment-1",
			Amount:    100.50,
			Currency:  "USD",
			MerchantID: "merchant-123",
			CardToken:  "tok_test123",
		}

		response, err := svc.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Contains(t, err.Error(), "payment service unavailable")
		require.Equal(t, service.PaymentResponse{}, response)
	})

	t.Run("context cancellation", func(t *testing.T) {
		svc, err := service.NewPaymentService(100*time.Millisecond, 0.0)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		request := service.PaymentRequest{
			ID:        "test-payment-1",
			Amount:    100.50,
			Currency:  "USD",
			MerchantID: "merchant-123",
			CardToken:  "tok_test123",
		}

		// Cancel immediately
		cancel()

		response, err := svc.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, context.Canceled, err)
		require.Equal(t, service.PaymentResponse{}, response)
	})

	t.Run("context timeout", func(t *testing.T) {
		svc, err := service.NewPaymentService(200*time.Millisecond, 0.0)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		request := service.PaymentRequest{
			ID:        "test-payment-1",
			Amount:    100.50,
			Currency:  "USD",
			MerchantID: "merchant-123",
			CardToken:  "tok_test123",
		}

		response, err := svc.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, context.DeadlineExceeded, err)
		require.Equal(t, service.PaymentResponse{}, response)
	})
}

func TestValidateRequest(t *testing.T) {
	svc, err := service.NewPaymentService(10*time.Millisecond, 0.0)
	require.NoError(t, err)

	ctx := context.Background()

	validRequest := service.PaymentRequest{
		ID:        "test-payment-1",
		Amount:    100.50,
		Currency:  "USD",
		MerchantID: "merchant-123",
		CardToken:  "tok_test123",
	}

	// Test valid request
	response, err := svc.ProcessPayment(ctx, validRequest)
	require.NoError(t, err)
	require.NotEmpty(t, response.ID)

	testCases := []struct {
		name     string
		request  service.PaymentRequest
		errorMsg string
	}{
		{
			name: "missing ID",
			request: service.PaymentRequest{
				Amount:     100.50,
				Currency:   "USD",
				MerchantID: "merchant-123",
				CardToken:  "tok_test123",
			},
			errorMsg: "payment ID is required",
		},
		{
			name: "zero amount",
			request: service.PaymentRequest{
				ID:        "test-payment-1",
				Amount:    0,
				Currency:  "USD",
				MerchantID: "merchant-123",
				CardToken:  "tok_test123",
			},
			errorMsg: "payment amount must be greater than 0",
		},
		{
			name: "negative amount",
			request: service.PaymentRequest{
				ID:        "test-payment-1",
				Amount:    -10.50,
				Currency:  "USD",
				MerchantID: "merchant-123",
				CardToken:  "tok_test123",
			},
			errorMsg: "payment amount must be greater than 0",
		},
		{
			name: "missing currency",
			request: service.PaymentRequest{
				ID:        "test-payment-1",
				Amount:    100.50,
				MerchantID: "merchant-123",
				CardToken:  "tok_test123",
			},
			errorMsg: "currency is required",
		},
		{
			name: "missing merchant ID",
			request: service.PaymentRequest{
				ID:       "test-payment-1",
				Amount:   100.50,
				Currency: "USD",
				CardToken: "tok_test123",
			},
			errorMsg: "merchant ID is required",
		},
		{
			name: "missing card token",
			request: service.PaymentRequest{
				ID:        "test-payment-1",
				Amount:    100.50,
				Currency:  "USD",
				MerchantID: "merchant-123",
			},
			errorMsg: "card token is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := svc.ProcessPayment(ctx, tc.request)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errorMsg)
			require.Equal(t, service.PaymentResponse{}, response)
		})
	}
}

func TestSetFailureRate(t *testing.T) {
	svc, err := service.NewPaymentService(10*time.Millisecond, 0.1)
	require.NoError(t, err)

	t.Run("valid failure rates", func(t *testing.T) {
		testCases := []float64{0.0, 0.5, 1.0}
		for _, rate := range testCases {
			err := svc.SetFailureRate(rate)
			require.NoError(t, err)
		}
	})

	t.Run("invalid failure rates", func(t *testing.T) {
		testCases := []float64{-0.1, 1.1, 2.0}
		for _, rate := range testCases {
			err := svc.SetFailureRate(rate)
			require.Error(t, err)
			require.Contains(t, err.Error(), "failure rate must be between 0 and 1")
		}
	})
}

func TestHealthToggle(t *testing.T) {
	svc, err := service.NewPaymentService(10*time.Millisecond, 0.0)
	require.NoError(t, err)

	// Initially healthy
	require.True(t, svc.IsHealthy())

	// Set unhealthy
	svc.SetHealthy(false)
	require.False(t, svc.IsHealthy())

	// Set healthy again
	svc.SetHealthy(true)
	require.True(t, svc.IsHealthy())
}