package main

import (
	"errors"
	"log"
	"time"
)

const (
	Closed = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	failures    int
	threshold   int
	state       int
	lastFail    time.Time
	cooldown    time.Duration
	requests    int
	maxRequests int
}

var (
	ErrCircuitOpen     = errors.New("circuit is open – skipping call")
	ErrCircuitHalfOpen = errors.New("circuit is half-open – too many requests")
)

func (cb *CircuitBreaker) Call(fn func() error) error {
	now := time.Now()

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
		cb.failures++
		cb.lastFail = now
		if cb.failures >= cb.threshold {
			cb.state = Open
		}
		return err
	}

	// Success → reset
	cb.failures = 0
	cb.state = Closed
	return nil
}

func main() {
	cb := &CircuitBreaker{
		threshold:   3,
		cooldown:    2 * time.Second,
		maxRequests: 2,
	}

	failing := func() error { return errors.New("service down") }
	success := func() error { return nil }

	// Initial state is Closed
	log.Println("Success:", cb.Call(success))

	for i := 1; i <= 3; i++ {
		log.Println("Fail", i, "→", cb.Call(failing))
	}

	log.Println("State should now be Open:", cb.state)

	time.Sleep(3 * time.Second) // allow cooldown

	log.Println("Cooldown period passed, trying Half-Open state...")

	log.Println("Request 1:", cb.Call(failing))
	log.Println("Request 2:", cb.Call(success))
}
