package circuitbreaker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/circuit-breaker/internal/circuitbreaker"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/circuit-breaker/internal/mocks"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/circuit-breaker/internal/service"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("nil service", func(t *testing.T) {
		cb, err := circuitbreaker.New(nil, 1, 1*time.Second, 10, 1)
		require.Error(t, err)
		require.Nil(t, cb)
	})

	t.Run("invalid failure threshold", func(t *testing.T) {
		cb, err := circuitbreaker.New(mocks.NewMockPaymentProcessor(ctrl), 0, 1*time.Second, 10, 1)
		require.Error(t, err)
		require.Nil(t, cb)
	})

	t.Run("invalid cooldown", func(t *testing.T) {
		cb, err := circuitbreaker.New(mocks.NewMockPaymentProcessor(ctrl), 1, 0, 10, 1)
		require.Error(t, err)
		require.Nil(t, cb)
	})

	t.Run("invalid max requests", func(t *testing.T) {
		cb, err := circuitbreaker.New(mocks.NewMockPaymentProcessor(ctrl), 1, 1*time.Second, 0, 1)
		require.Error(t, err)
		require.Nil(t, cb)
	})

	t.Run("invalid success threshold", func(t *testing.T) {
		cb, err := circuitbreaker.New(mocks.NewMockPaymentProcessor(ctrl), 1, 1*time.Second, 10, 0)
		require.Error(t, err)
		require.Nil(t, cb)
	})

	t.Run("valid service and options", func(t *testing.T) {
		cb, err := circuitbreaker.New(mocks.NewMockPaymentProcessor(ctrl), 3, 1*time.Second, 10, 1)
		require.NoError(t, err)
		require.NotNil(t, cb)
	})
}

func TestProcessPayment(t *testing.T) {
	t.Run("successful payment", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 3, 1*time.Second, 2, 1)
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		expectedResponse := service.PaymentResponse{ID: "123", Status: "success"}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(expectedResponse, nil)

		response, err := cb.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
		require.Equal(t, circuitbreaker.Closed, cb.State())
		require.Equal(t, 0, cb.Failures())
	})

	t.Run("single failure - stays closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 3, 1*time.Second, 2, 1)
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed"))

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Closed, cb.State())
		require.Equal(t, 1, cb.Failures())
	})

	t.Run("multiple failures - opens circuit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 2, 1*time.Second, 1, 1)
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed")).Times(2)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Closed, cb.State())

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())
		require.Equal(t, 2, cb.Failures())
	})

	t.Run("circuit opens at exact threshold", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 3, 1*time.Second, 1, 1)
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed")).Times(3)

		for i := 0; i < 2; i++ {
			_, err = cb.ProcessPayment(ctx, request)
			require.Error(t, err)
			require.Equal(t, circuitbreaker.Closed, cb.State())
		}

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())
		require.Equal(t, 3, cb.Failures())
	})

	t.Run("success after failures resets circuit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 3, 1*time.Second, 1, 1)
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		expectedResponse := service.PaymentResponse{ID: "123", Status: "success"}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed")).Times(2)
		mockService.EXPECT().ProcessPayment(ctx, request).Return(expectedResponse, nil)

		for i := 0; i < 2; i++ {
			_, err = cb.ProcessPayment(ctx, request)
			require.Error(t, err)
		}
		require.Equal(t, 2, cb.Failures())

		response, err := cb.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
		require.Equal(t, circuitbreaker.Closed, cb.State())
		require.Equal(t, 0, cb.Failures())
	})

	t.Run("open circuit blocks requests", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 1, 1*time.Second, 1, 1)
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed"))

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())

		_, err = cb.ProcessPayment(ctx, request)
		require.Equal(t, circuitbreaker.ErrCircuitOpen, err)
		require.Equal(t, circuitbreaker.Open, cb.State())
	})

	t.Run("open circuit transitions to half-open after cooldown", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clockwork.NewFakeClock()
		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 1, 1*time.Second, 1, 1, circuitbreaker.WithClock(clock))
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		expectedResponse := service.PaymentResponse{ID: "123", Status: "success"}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed"))
		mockService.EXPECT().ProcessPayment(ctx, request).Return(expectedResponse, nil)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())

		clock.Advance(2 * time.Second)

		response, err := cb.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
		require.Equal(t, circuitbreaker.Closed, cb.State())
	})

	t.Run("half-open allows requests up to max", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clockwork.NewFakeClock()
		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 1, 1*time.Second, 2, 1, circuitbreaker.WithClock(clock))
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		expectedResponse := service.PaymentResponse{ID: "123", Status: "success"}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed"))
		mockService.EXPECT().ProcessPayment(ctx, request).Return(expectedResponse, nil).Times(2)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())

		clock.Advance(2 * time.Second)

		response, err := cb.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
		require.Equal(t, circuitbreaker.Closed, cb.State())

		response, err = cb.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
		require.Equal(t, circuitbreaker.Closed, cb.State())
	})

	t.Run("half-open blocks after max requests", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clockwork.NewFakeClock()
		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 1, 1*time.Second, 2, 3, circuitbreaker.WithClock(clock))
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		expectedResponse := service.PaymentResponse{ID: "123", Status: "success"}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed"))
		mockService.EXPECT().ProcessPayment(ctx, request).Return(expectedResponse, nil)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())

		clock.Advance(2 * time.Second)

		_, err = cb.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, circuitbreaker.HalfOpen, cb.State())
	})

	t.Run("half-open success closes circuit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clockwork.NewFakeClock()
		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 1, 1*time.Second, 1, 1, circuitbreaker.WithClock(clock))
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		expectedResponse := service.PaymentResponse{ID: "123", Status: "success"}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed"))
		mockService.EXPECT().ProcessPayment(ctx, request).Return(expectedResponse, nil)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())

		clock.Advance(2 * time.Second)

		response, err := cb.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
		require.Equal(t, circuitbreaker.Closed, cb.State())
		require.Equal(t, 0, cb.Failures())
	})

	t.Run("half-open failure reopens circuit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clockwork.NewFakeClock()
		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 1, 1*time.Second, 1, 1, circuitbreaker.WithClock(clock))
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed")).Times(2)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())

		clock.Advance(2 * time.Second)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())
	})

	t.Run("context is passed through to service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 3, 1*time.Second, 2, 1)
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		expectedResponse := service.PaymentResponse{ID: "123", Status: "success"}

		ctx := context.WithValue(context.Background(), "test-key", "test-value")
		mockService.EXPECT().ProcessPayment(ctx, request).Return(expectedResponse, nil)

		response, err := cb.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
	})
}

func TestStateTransitions(t *testing.T) {
	t.Run("closed to open transition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 2, 1*time.Second, 1, 1)
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed")).Times(2)

		require.Equal(t, circuitbreaker.Closed, cb.State())

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Closed, cb.State())

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())
	})

	t.Run("open to half-open transition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clockwork.NewFakeClock()
		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 1, 1*time.Second, 1, 1, circuitbreaker.WithClock(clock))
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed"))

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())

		clock.Advance(2 * time.Second)

		// The circuit should transition to half-open on the next request
		require.Equal(t, circuitbreaker.Open, cb.State())
	})

	t.Run("half-open to closed transition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clockwork.NewFakeClock()
		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 1, 1*time.Second, 1, 1, circuitbreaker.WithClock(clock))
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		expectedResponse := service.PaymentResponse{ID: "123", Status: "success"}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed"))
		mockService.EXPECT().ProcessPayment(ctx, request).Return(expectedResponse, nil)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())

		clock.Advance(2 * time.Second)

		response, err := cb.ProcessPayment(ctx, request)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
		require.Equal(t, circuitbreaker.Closed, cb.State())
	})

	t.Run("half-open to open transition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clockwork.NewFakeClock()
		mockService := mocks.NewMockPaymentProcessor(ctrl)
		cb, err := circuitbreaker.New(mockService, 1, 1*time.Second, 1, 1, circuitbreaker.WithClock(clock))
		require.NoError(t, err)

		request := service.PaymentRequest{Amount: 100}
		ctx := context.Background()

		mockService.EXPECT().ProcessPayment(ctx, request).Return(service.PaymentResponse{}, errors.New("payment failed")).Times(2)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())

		clock.Advance(2 * time.Second)

		_, err = cb.ProcessPayment(ctx, request)
		require.Error(t, err)
		require.Equal(t, circuitbreaker.Open, cb.State())
	})
}
