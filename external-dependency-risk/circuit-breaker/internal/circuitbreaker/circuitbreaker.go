package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/circuit-breaker/internal/service"
)

// State represents the circuit breaker state
type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	switch s {
	case Closed:
		return "Closed"
	case Open:
		return "Open"
	case HalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

var (
	ErrCircuitOpen     = errors.New("circuit is open – skipping call")
	ErrCircuitHalfOpen = errors.New("circuit is half-open – too many requests")
)

// PaymentProcessor defines the interface for payment processing operations
type PaymentProcessor interface {
	ProcessPayment(ctx context.Context, request service.PaymentRequest) (service.PaymentResponse, error)
}

// circuitBreaker wraps a payment service with circuit breaker functionality
type circuitBreaker struct {
	service PaymentProcessor
	lock    sync.RWMutex
	clock   clockwork.Clock

	// Configuration
	failureThreshold int           // Number of failures to trigger opening
	successThreshold int           // Number of consecutive successful requests before closing the circuit
	cooldown         time.Duration // Time to wait before allowing retry
	maxRequests      int           // Max requests in half-open state

	// State
	state     State
	failures  int
	lastFail  time.Time
	requests  int // Current request count in half-open state
	successes int // Current consecutive successful requests
}

// Option is a functional option for configuring the circuit breaker
type Option func(*circuitBreaker) error

// WithClock sets a custom clock for the circuit breaker
func WithClock(clock clockwork.Clock) Option {
	return func(cb *circuitBreaker) error {
		if clock == nil {
			return errors.New("clock is nil")
		}
		cb.clock = clock
		return nil
	}
}

// New creates a new circuit breaker
func New(service PaymentProcessor, failureThreshold int, cooldown time.Duration, maxRequests, successThreshold int, opts ...Option) (*circuitBreaker, error) {
	switch {
	case service == nil:
		return nil, errors.New("service is nil")
	case failureThreshold <= 0:
		return nil, errors.New("failureThreshold must be greater than 0")
	case cooldown <= 0:
		return nil, errors.New("cooldown must be greater than 0")
	case maxRequests <= 0:
		return nil, errors.New("maxRequests must be greater than 0")
	case successThreshold <= 0:
		return nil, errors.New("successThreshold must be greater than 0")
	}

	cb := &circuitBreaker{
		service:          service,
		failureThreshold: failureThreshold,
		cooldown:         cooldown,
		maxRequests:      maxRequests,
		state:            Closed,
		successThreshold: successThreshold,
		clock:            clockwork.NewRealClock(), // Default to real clock
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(cb); err != nil {
			return nil, err
		}
	}

	return cb, nil
}

// Call executes a function through the circuit breaker
func (cb *circuitBreaker) call(fn func() error) error {
	cb.lock.Lock()
	defer cb.lock.Unlock()

	now := cb.clock.Now()

	if cb.state == Open {
		if now.Sub(cb.lastFail) > cb.cooldown {
			// If cooldown period has passed, transition to HalfOpen
			cb.state = HalfOpen
			cb.requests = 0
		} else {
			return ErrCircuitOpen
		}
	}

	if cb.state == HalfOpen && cb.requests >= cb.maxRequests {
		return ErrCircuitHalfOpen
	}

	cb.requests++
	err := fn() // call the function
	if err != nil {
		cb.successes = 0
		cb.failures++
		cb.lastFail = now
		if cb.failures >= cb.failureThreshold {
			cb.state = Open
		}
		return err
	}

	// Success → reset
	cb.successes++
	cb.failures = 0
	if cb.successes >= cb.successThreshold {
		cb.state = Closed
	}
	cb.requests = 0
	return nil
}

// ProcessPayment processes a payment request through the circuit breaker
func (cb *circuitBreaker) ProcessPayment(ctx context.Context, request service.PaymentRequest) (service.PaymentResponse, error) {
	var response service.PaymentResponse

	err := cb.call(func() error {
		var err error
		response, err = cb.service.ProcessPayment(ctx, request)
		return err
	})
	if err != nil {
		return service.PaymentResponse{}, err
	}

	return response, nil
}

// State returns the current state of the circuit breaker
func (cb *circuitBreaker) State() State {
	cb.lock.RLock()
	defer cb.lock.RUnlock()
	return cb.state
}

// Failures returns the current failure count
func (cb *circuitBreaker) Failures() int {
	cb.lock.RLock()
	defer cb.lock.RUnlock()
	return cb.failures
}
