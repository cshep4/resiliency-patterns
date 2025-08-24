package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/circuit-breaker/internal/circuitbreaker"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/circuit-breaker/internal/service"
)

func main() {
	log.Println("🔌 Circuit Breaker Demonstration")
	log.Println("================================")

	// Create a payment service with some delay and failure rate
	paymentService, err := service.NewPaymentService(200*time.Millisecond, 0.0)
	if err != nil {
		log.Fatalf("Failed to create payment service: %v", err)
	}

	// Create circuit breaker with custom configuration
	circuitBreaker, err := circuitbreaker.New(
		paymentService,
		circuitbreaker.WithFailureThreshold(3),
		circuitbreaker.WithTimeout(5*time.Second),
		circuitbreaker.WithMaxRequests(2),
	)
	if err != nil {
		log.Fatalf("Failed to create circuit breaker: %v", err)
	}

	ctx := context.Background()

	log.Println()

	// Demonstrate normal operation
	demonstrateNormalOperation(ctx, circuitBreaker, paymentService)

	log.Println()

	// Demonstrate circuit breaker opening
	demonstrateCircuitOpening(ctx, circuitBreaker, paymentService)

	log.Println()

	// Demonstrate circuit breaker recovery
	demonstrateCircuitRecovery(ctx, circuitBreaker, paymentService)

	log.Println()
	log.Println("🎉 Circuit breaker demonstration complete!")
}

func demonstrateNormalOperation(ctx context.Context, cb circuitbreaker.CircuitBreaker, svc service.ControllablePaymentService) {
	log.Println("✅ Normal Operation Demo")
	log.Println("------------------------")

	request := service.PaymentRequest{
		ID:        "payment-001",
		Amount:    99.99,
		Currency:  "USD",
		MerchantID: "merchant-abc",
		CardToken:  "tok_1234567890",
	}

	log.Printf("🔍 Circuit state: %s, Failures: %d\n", cb.State(), cb.Failures())

	start := time.Now()
	response, err := cb.ProcessPayment(ctx, request)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Payment failed: %v\n", err)
	} else {
		log.Printf("✅ Payment processed successfully!\n")
		log.Printf("   💳 Transaction ID: %s\n", response.TransactionID)
		log.Printf("   💰 Amount: $%.2f %s\n", response.Amount, response.Currency)
		log.Printf("   ⏱️  Processing time: %v\n", duration)
	}

	log.Printf("🔍 Circuit state: %s, Failures: %d\n", cb.State(), cb.Failures())
}

func demonstrateCircuitOpening(ctx context.Context, cb circuitbreaker.CircuitBreaker, svc service.ControllablePaymentService) {
	log.Println("🚨 Circuit Opening Demo")
	log.Println("-----------------------")

	// Make service unhealthy to trigger failures
	log.Println("💥 Simulating service failures...")
	svc.SetHealthy(false)

	request := service.PaymentRequest{
		ID:        "payment-002",
		Amount:    149.99,
		Currency:  "USD", 
		MerchantID: "merchant-xyz",
		CardToken:  "tok_9876543210",
	}

	// Trigger failures to open the circuit
	for i := 1; i <= 4; i++ {
		log.Printf("🔍 Attempt %d - Circuit state: %s, Failures: %d\n", i, cb.State(), cb.Failures())
		
		start := time.Now()
		_, err := cb.ProcessPayment(ctx, request)
		duration := time.Since(start)

		if err != nil {
			if err == circuitbreaker.ErrCircuitOpen {
				log.Printf("🔴 Circuit is OPEN - Request blocked immediately (took %v)\n", duration)
			} else {
				log.Printf("❌ Payment failed: %v (took %v)\n", err, duration)
			}
		}

		if cb.IsOpen() && i < 4 {
			log.Printf("🔴 Circuit opened after %d failures!\n", cb.Failures())
		}
	}

	log.Printf("🔍 Final state - Circuit: %s, Failures: %d\n", cb.State(), cb.Failures())
}

func demonstrateCircuitRecovery(ctx context.Context, cb circuitbreaker.CircuitBreaker, svc service.ControllablePaymentService) {
	log.Println("🔄 Circuit Recovery Demo")
	log.Println("------------------------")

	log.Println("⏳ Waiting for circuit breaker timeout...")
	time.Sleep(6 * time.Second) // Wait longer than the 5-second timeout

	request := service.PaymentRequest{
		ID:        "payment-003",
		Amount:    199.99,
		Currency:  "USD",
		MerchantID: "merchant-recovery",
		CardToken:  "tok_recovery123",
	}

	// First request should transition to half-open
	log.Printf("🔍 After timeout - Circuit state: %s\n", cb.State())
	log.Println("🔄 Attempting request (should transition to half-open)...")
	
	_, err := cb.ProcessPayment(ctx, request)
	if err != nil {
		log.Printf("❌ Request failed (circuit half-open): %v\n", err)
	}
	log.Printf("🔍 Circuit state: %s\n", cb.State())

	// Restore service health and make successful request
	log.Println("🩹 Restoring service health...")
	svc.SetHealthy(true)

	log.Println("🔄 Making successful request to close circuit...")
	start := time.Now()
	response, err := cb.ProcessPayment(ctx, request)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Recovery attempt failed: %v\n", err)
	} else {
		log.Printf("✅ Circuit recovered! Payment processed successfully!\n")
		log.Printf("   💳 Transaction ID: %s\n", response.TransactionID)
		log.Printf("   💰 Amount: $%.2f %s\n", response.Amount, response.Currency)
		log.Printf("   ⏱️  Processing time: %v\n", duration)
	}

	log.Printf("🔍 Final circuit state: %s, Failures: %d\n", cb.State(), cb.Failures())

	// Demonstrate that circuit is fully operational
	log.Println("🧪 Testing circuit is fully operational...")
	for i := 1; i <= 3; i++ {
		testRequest := service.PaymentRequest{
			ID:        fmt.Sprintf("payment-test-%d", i),
			Amount:    50.00,
			Currency:  "USD",
			MerchantID: "merchant-test",
			CardToken:  "tok_test",
		}

		_, err := cb.ProcessPayment(ctx, testRequest)
		if err != nil {
			log.Printf("❌ Test payment %d failed: %v\n", i, err)
		} else {
			log.Printf("✅ Test payment %d successful\n", i)
		}
	}
}