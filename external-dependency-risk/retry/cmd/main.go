package main

import (
	"context"
	"log"
	"time"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/retry/internal/retry"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/retry/internal/service"
)

func main() {
	log.Println("🔄 Retry Pattern Demonstration")
	log.Println("==============================")

	log.Println()

	// Demonstrate successful retry after failures
	demonstrateSuccessfulRetry()

	log.Println()

	// Demonstrate exponential backoff
	demonstrateBackoffStrategy()

	log.Println()

	// Demonstrate max attempts exceeded
	demonstrateMaxAttemptsExceeded()

	log.Println()
	log.Println("🎉 Retry pattern demonstration complete!")
}

func demonstrateSuccessfulRetry() {
	log.Println("✅ Successful Retry Demo")
	log.Println("------------------------")

	// Create order service with 70% failure rate and 100ms delay
	orderService, err := service.NewOrderService(100*time.Millisecond, 0.01)
	if err != nil {
		log.Fatalf("Failed to create order service: %v", err)
	}

	// Create retry client with 5 attempts
	retryClient, err := retry.New(
		orderService,
		5,                    // max attempts
		2*time.Second,        // timeout per attempt
		100*time.Millisecond, // initial interval
		1*time.Second,        // max interval
		2.0,                  // multiplier
	)
	if err != nil {
		log.Fatalf("Failed to create retry client: %v", err)
	}

	ctx := context.Background()
	request := service.OrderRequest{
		ID:       "order-001",
		UserID:   "user-123",
		Amount:   99.99,
		Currency: "USD",
		Items: []service.Item{
			{ProductID: "prod-1", Quantity: 2, Price: 29.99},
			{ProductID: "prod-2", Quantity: 1, Price: 39.99},
		},
	}

	start := time.Now()

	response, err := retryClient.ProcessOrder(ctx, request)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Order failed after retries: %v\n", err)
		return
	}

	log.Printf("✅ Order processed successfully!\n")
	log.Printf("   📦 Order ID: %s\n", response.OrderID)
	log.Printf("   💰 Amount: $%.2f %s\n", response.Amount, response.Currency)
	log.Printf("   ⏱️  Total time: %v (including retries)\n", duration)
}

func demonstrateBackoffStrategy() {
	log.Println("⏰ Exponential Backoff Demo")
	log.Println("---------------------------")

	// Create order service that always fails initially
	orderService, err := service.NewOrderService(50*time.Millisecond, 0.9)
	if err != nil {
		log.Fatalf("Failed to create order service: %v", err)
	}

	// Create retry client with clear backoff progression
	retryClient, err := retry.New(
		orderService,
		100,                  // max attempts
		1*time.Second,        // timeout per attempt
		200*time.Millisecond, // initial interval
		800*time.Millisecond, // max interval
		2.0,                  // multiplier
	)
	if err != nil {
		log.Fatalf("Failed to create retry client: %v", err)
	}

	ctx := context.Background()
	request := service.OrderRequest{
		ID:       "order-002",
		UserID:   "user-456",
		Amount:   149.99,
		Currency: "USD",
		Items: []service.Item{
			{ProductID: "prod-3", Quantity: 1, Price: 149.99},
		},
	}

	log.Println("🔍 Demonstrating backoff delays (service will fail initially):")
	log.Println("   Expected delays: 200ms, 400ms, 800ms (capped)")

	start := time.Now()
	response, err := retryClient.ProcessOrder(ctx, request)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Order failed: %v\n", err)
		return
	}

	log.Printf("✅ Order eventually succeeded!\n")
	log.Printf("   📦 Order ID: %s\n", response.OrderID)
	log.Printf("   ⏱️  Total time: %v\n", duration)
}

func demonstrateMaxAttemptsExceeded() {
	log.Println("🚫 Max Attempts Exceeded Demo")
	log.Println("-----------------------------")

	// Create order service that always fails
	orderService, err := service.NewOrderService(100*time.Millisecond, 1)
	if err != nil {
		log.Fatalf("Failed to create order service: %v", err)
	}

	// Create retry client with limited attempts
	retryClient, err := retry.New(
		orderService,
		3,                    // max attempts
		1*time.Second,        // timeout per attempt
		150*time.Millisecond, // initial interval
		600*time.Millisecond, // max interval
		2.0,                  // multiplier
	)
	if err != nil {
		log.Fatalf("Failed to create retry client: %v", err)
	}

	ctx := context.Background()
	request := service.OrderRequest{
		ID:       "order-003",
		UserID:   "user-789",
		Amount:   199.99,
		Currency: "USD",
		Items: []service.Item{
			{ProductID: "prod-4", Quantity: 3, Price: 66.66},
		},
	}

	start := time.Now()

	response, err := retryClient.ProcessOrder(ctx, request)
	duration := time.Since(start)

	if err != nil {
		if err == retry.ErrMaxAttemptsExceeded {
			log.Printf("❌ Order failed: Maximum attempts exceeded\n")
		} else {
			log.Printf("❌ Order failed: %v\n", err)
		}
		log.Printf("   ⏱️  Total time: %v\n", duration)
		return
	}

	// This shouldn't happen in this demo
	log.Printf("✅ Order succeeded: %s\n", response.OrderID)
}
