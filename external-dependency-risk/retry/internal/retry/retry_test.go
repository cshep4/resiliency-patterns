package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/retry/internal/mocks"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/retry/internal/retry"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/retry/internal/service"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("valid configuration", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		r, err := retry.New(mockService, 3, time.Second, 100*time.Millisecond, time.Second, 2.0)
		require.NoError(t, err)
		require.NotNil(t, r)
	})

	t.Run("nil service", func(t *testing.T) {
		r, err := retry.New(nil, 3, time.Second, 100*time.Millisecond, time.Second, 2.0)
		require.Error(t, err)
		require.Nil(t, r)
		require.Contains(t, err.Error(), "service is nil")
	})

	t.Run("invalid max attempts", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		r, err := retry.New(mockService, 0, time.Second, 100*time.Millisecond, time.Second, 2.0)
		require.Error(t, err)
		require.Nil(t, r)
		require.Contains(t, err.Error(), "maxAttempts must be greater than 0")
	})

	t.Run("invalid timeout", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		r, err := retry.New(mockService, 3, 0, 100*time.Millisecond, time.Second, 2.0)
		require.Error(t, err)
		require.Nil(t, r)
		require.Contains(t, err.Error(), "timeout must be greater than 0")
	})

	t.Run("invalid initial interval", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		r, err := retry.New(mockService, 3, time.Second, 0, time.Second, 2.0)
		require.Error(t, err)
		require.Nil(t, r)
		require.Contains(t, err.Error(), "initialInterval must be greater than 0")
	})

	t.Run("invalid max interval", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		r, err := retry.New(mockService, 3, time.Second, 100*time.Millisecond, 0, 2.0)
		require.Error(t, err)
		require.Nil(t, r)
		require.Contains(t, err.Error(), "maxInterval must be greater than 0")
	})

	t.Run("invalid multiplier", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		r, err := retry.New(mockService, 3, time.Second, 100*time.Millisecond, time.Second, 0)
		require.Error(t, err)
		require.Nil(t, r)
		require.Contains(t, err.Error(), "multiplier must be greater than 0")
	})

	t.Run("with custom clock", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		fakeClock := clockwork.NewFakeClock()
		r, err := retry.New(mockService, 3, time.Second, 100*time.Millisecond, time.Second, 2.0, retry.WithClock(fakeClock))
		require.NoError(t, err)
		require.NotNil(t, r)
	})

	t.Run("with nil clock", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		r, err := retry.New(mockService, 3, time.Second, 100*time.Millisecond, time.Second, 2.0, retry.WithClock(nil))
		require.Error(t, err)
		require.Nil(t, r)
		require.Contains(t, err.Error(), "clock is nil")
	})
}

func TestProcessOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOrder := service.OrderResponse{
		ID:          "order-1",
		OrderID:     "ord-123",
		Status:      "completed",
		Amount:      99.99,
		Currency:    "USD",
		ProcessedAt: time.Now(),
	}

	t.Run("success on first attempt", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		r, err := retry.New(mockService, 3, time.Second, 100*time.Millisecond, time.Second, 2.0)
		require.NoError(t, err)

		ctx := context.Background()
		request := service.OrderRequest{ID: "order-1", Amount: 99.99}

		mockService.EXPECT().
			ProcessOrder(gomock.Any(), request).
			Return(expectedOrder, nil).
			Times(1)

		order, err := r.ProcessOrder(ctx, request)
		require.NoError(t, err)
		require.Equal(t, expectedOrder, order)
	})

	t.Run("success after retries", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		fakeClock := clockwork.NewFakeClock()
		r, err := retry.New(mockService, 3, time.Second, 100*time.Millisecond, time.Second, 2.0, retry.WithClock(fakeClock))
		require.NoError(t, err)

		ctx := context.Background()
		request := service.OrderRequest{ID: "order-1", Amount: 99.99}

		serviceErr := errors.New("service unavailable")
		mockService.EXPECT().
			ProcessOrder(gomock.Any(), request).
			Return(service.OrderResponse{}, serviceErr).
			Times(2)

		mockService.EXPECT().
			ProcessOrder(gomock.Any(), request).
			Return(expectedOrder, nil).
			Times(1)

		// Start the operation in a goroutine
		resultChan := make(chan struct {
			order service.OrderResponse
			err   error
		})

		go func() {
			order, err := r.ProcessOrder(ctx, request)
			resultChan <- struct {
				order service.OrderResponse
				err   error
			}{order, err}
		}()

		// Advance time to simulate backoff delays
		fakeClock.BlockUntilContext(ctx, 1) // Wait for first retry delay
		fakeClock.Advance(100 * time.Millisecond)
		fakeClock.BlockUntilContext(ctx, 1) // Wait for second retry delay
		fakeClock.Advance(200 * time.Millisecond)

		result := <-resultChan
		require.NoError(t, result.err)
		require.Equal(t, expectedOrder, result.order)
	})

	t.Run("max attempts exceeded", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		fakeClock := clockwork.NewFakeClock()
		r, err := retry.New(mockService, 2, time.Second, 100*time.Millisecond, time.Second, 2.0, retry.WithClock(fakeClock))
		require.NoError(t, err)

		ctx := context.Background()
		request := service.OrderRequest{ID: "order-1", Amount: 99.99}

		serviceErr := errors.New("service unavailable")
		mockService.EXPECT().
			ProcessOrder(gomock.Any(), request).
			Return(service.OrderResponse{}, serviceErr).
			Times(2)

		// Start the operation in a goroutine
		resultChan := make(chan struct {
			order service.OrderResponse
			err   error
		})

		go func() {
			order, err := r.ProcessOrder(ctx, request)
			resultChan <- struct {
				order service.OrderResponse
				err   error
			}{order, err}
		}()

		// Advance time to simulate backoff delay
		fakeClock.BlockUntilContext(ctx, 1)
		fakeClock.Advance(100 * time.Millisecond)

		result := <-resultChan
		require.Error(t, result.err)
		require.Equal(t, service.OrderResponse{}, result.order)
		require.Equal(t, retry.ErrMaxAttemptsExceeded, result.err)
	})

	t.Run("success after context cancellation (timeout)", func(t *testing.T) {
		mockService := mocks.NewMockOrderProcessor(ctrl)
		fakeClock := clockwork.NewFakeClock()
		r, err := retry.New(mockService, 3, time.Second, 100*time.Millisecond, time.Second, 2.0, retry.WithClock(fakeClock))
		require.NoError(t, err)

		ctx := context.Background()
		request := service.OrderRequest{ID: "order-1", Amount: 99.99}

		mockService.EXPECT().
			ProcessOrder(gomock.Any(), request).
			Return(service.OrderResponse{}, context.DeadlineExceeded).
			Times(2)

		mockService.EXPECT().
			ProcessOrder(gomock.Any(), request).
			Return(expectedOrder, nil).
			Times(1)

		// Start the operation in a goroutine
		resultChan := make(chan struct {
			order service.OrderResponse
			err   error
		})

		go func() {
			order, err := r.ProcessOrder(ctx, request)
			resultChan <- struct {
				order service.OrderResponse
				err   error
			}{order, err}
		}()

		// Advance time to simulate backoff delays
		fakeClock.BlockUntilContext(ctx, 1) // Wait for first retry delay
		fakeClock.Advance(100 * time.Millisecond)
		fakeClock.BlockUntilContext(ctx, 1) // Wait for second retry delay
		fakeClock.Advance(200 * time.Millisecond)

		result := <-resultChan
		require.NoError(t, result.err)
		require.Equal(t, expectedOrder, result.order)
	})
}
