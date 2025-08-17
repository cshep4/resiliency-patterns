package circuitbreaker

import (
	"context"
	"errors"
)

var ErrCircuitOpen = errors.New("circuit open – skipping call")

// State represents the circuit breaker state
type State int

const (
	Closed State = iota
	Open
)

// CircuitBreaker implements a basic circuit breaker pattern
type CircuitBreaker struct {
	failures  int
	threshold int
	state     State
}

// New creates a new circuit breaker with the specified failure threshold
func New(threshold int) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		state:     Closed,
	}
}

// Call executes the provided function with circuit breaker protection
func (cb *CircuitBreaker) Call(ctx context.Context, fn func(ctx context.Context) error) error {
	if cb.state == Open {
		return ErrCircuitOpen
	}

	err := fn(ctx)
	if err != nil {
		cb.failures++
		if cb.failures >= cb.threshold {
			cb.state = Open
		}
		return err
	}

	cb.failures = 0 // reset on success
	return nil
}

// CallWithContext executes the provided function with circuit breaker protection and context support
func (cb *CircuitBreaker) CallWithContext(ctx context.Context, funcCall func(context.Context) error) error {
	if cb.state == Open {
		return ErrCircuitOpen
	}

	// Check if context is already cancelled before making the call
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := fn(ctx)
	if err != nil {
		// Don't count context cancellation as a failure
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		cb.failures++
		if cb.failures >= cb.threshold {
			cb.state = Open
		}
		return err
	}

	cb.failures = 0 // reset on success
	return nil
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	return cb.state
}

// Failures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	return cb.failures
}
