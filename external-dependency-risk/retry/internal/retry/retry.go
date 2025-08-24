package retry

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/retry/internal/service"
)

var ErrMaxAttemptsExceeded = errors.New("max attempts exceeded")

// OrderProcessor defines the interface for order processing operations
type OrderProcessor interface {
	ProcessOrder(ctx context.Context, request service.OrderRequest) (service.OrderResponse, error)
}

// retryClient wraps an order service with retry functionality
type retryClient struct {
	service         OrderProcessor
	maxAttempts     int
	timeout         time.Duration
	initialInterval time.Duration
	maxInterval     time.Duration
	multiplier      float64
	clock           clockwork.Clock
}

// Option is a functional option for configuring the retry client
type Option func(*retryClient) error

// WithClock sets a custom clock for the retry client
func WithClock(clock clockwork.Clock) Option {
	return func(r *retryClient) error {
		if clock == nil {
			return errors.New("clock is nil")
		}
		r.clock = clock
		return nil
	}
}

// New creates a new retry client
func New(service OrderProcessor, maxAttempts int, timeout, initialInterval, maxInterval time.Duration, multiplier float64, opts ...Option) (*retryClient, error) {
	switch {
	case service == nil:
		return nil, errors.New("service is nil")
	case maxAttempts <= 0:
		return nil, errors.New("maxAttempts must be greater than 0")
	case timeout <= 0:
		return nil, errors.New("timeout must be greater than 0")
	case initialInterval <= 0:
		return nil, errors.New("initialInterval must be greater than 0")
	case maxInterval <= 0:
		return nil, errors.New("maxInterval must be greater than 0")
	case multiplier <= 0:
		return nil, errors.New("multiplier must be greater than 0")
	}

	r := &retryClient{
		service:         service,
		maxAttempts:     maxAttempts,
		timeout:         timeout,
		initialInterval: initialInterval,
		maxInterval:     maxInterval,
		multiplier:      multiplier,
		clock:           clockwork.NewRealClock(),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// ProcessOrder processes an order request with retry logic and exponential backoff
func (r *retryClient) ProcessOrder(ctx context.Context, req service.OrderRequest) (service.OrderResponse, error) {
	for i := 0; i < r.maxAttempts; i++ {
		// Create timeout context for this attempt
		ctx, cancel := context.WithTimeout(ctx, r.timeout)

		// Try the operation
		resp, err := r.service.ProcessOrder(ctx, req)
		cancel()

		if err == nil {
			return resp, nil
		}

		// Don't wait after the last attempt
		if i < r.maxAttempts-1 {
			<-r.clock.After(r.backoffDelay(i))
		}
	}

	return service.OrderResponse{}, ErrMaxAttemptsExceeded
}

// backoffDelay calculates the exponential backoff delay
func (r *retryClient) backoffDelay(attempt int) time.Duration {
	delay := float64(r.initialInterval) * math.Pow(r.multiplier, float64(attempt))
	if time.Duration(delay) > r.maxInterval {
		return r.maxInterval
	}
	return time.Duration(delay)
}
