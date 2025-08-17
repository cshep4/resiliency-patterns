package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/circuit-breaker/internal/circuitbreaker"
)

func main() {
	cb := circuitbreaker.New(3)

	fmt.Println("Circuit Breaker Example")
	fmt.Println("======================")
	fmt.Printf("Threshold: %d failures\n\n", 3)

	// Demo 1: Basic circuit breaker behavior
	fmt.Println("Demo 1: Basic Circuit Breaker")
	fmt.Println("-----------------------------")
	
	failingService := func() error { 
		return errors.New("service down") 
	}

	for i := 1; i <= 5; i++ {
		err := cb.Call(failingService)
		fmt.Printf("Call %d → %v (failures: %d, state: %v)\n", 
			i, err, cb.GetFailures(), getStateName(cb.GetState()))
	}

	fmt.Println()

	// Demo 2: Context with timeout
	fmt.Println("Demo 2: Context with Timeout")
	fmt.Println("----------------------------")
	
	// Reset circuit breaker for demo
	cb = circuitbreaker.New(3)
	
	slowService := func(ctx context.Context) error {
		select {
		case <-time.After(2 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := cb.CallWithContext(ctx, slowService)
	fmt.Printf("Slow service call → %v (failures: %d, state: %v)\n", 
		err, cb.GetFailures(), getStateName(cb.GetState()))

	// Demo 3: Context cancellation doesn't count as failure
	fmt.Println()
	fmt.Println("Demo 3: Context Cancellation Handling")
	fmt.Println("-------------------------------------")

	for i := 1; i <= 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		err := cb.CallWithContext(ctx, slowService)
		cancel()
		fmt.Printf("Timeout call %d → %v (failures: %d, state: %v)\n", 
			i, err, cb.GetFailures(), getStateName(cb.GetState()))
	}

	// Demo 4: Actual failures after timeouts
	fmt.Println()
	fmt.Println("Demo 4: Real Failures After Timeouts")
	fmt.Println("------------------------------------")

	actualFailingService := func(ctx context.Context) error {
		return errors.New("actual service failure")
	}

	for i := 1; i <= 4; i++ {
		ctx := context.Background()
		err := cb.CallWithContext(ctx, actualFailingService)
		fmt.Printf("Failing call %d → %v (failures: %d, state: %v)\n", 
			i, err, cb.GetFailures(), getStateName(cb.GetState()))
	}
}

func getStateName(state circuitbreaker.State) string {
	switch state {
	case circuitbreaker.Closed:
		return "CLOSED"
	case circuitbreaker.Open:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}
