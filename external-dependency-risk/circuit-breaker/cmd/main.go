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

	log.Println()

	// Demonstrate normal operation
	demonstrateNormalOperation()

	log.Println()

	// Demonstrate circuit breaker opening
	demonstrateCircuitOpening()

	log.Println()

	log.Println("🎉 Circuit breaker demonstration complete!")
}

func demonstrateNormalOperation() {
	log.Println("✅ Normal Operation Demo")
	log.Println("------------------------")

	paymentService, err := service.NewPaymentService(0.0)
	if err != nil {
		log.Fatalf("Failed to create payment service: %v", err)
	}

	cb, err := circuitbreaker.New(paymentService, 3, 2*time.Second, 2, 1)
	if err != nil {
		log.Fatalf("Failed to create circuit breaker: %v", err)
	}

	ctx := context.Background()

	request := service.PaymentRequest{
		ID:         "payment-001",
		Amount:     99.99,
		Currency:   "USD",
		MerchantID: "merchant-abc",
		CardToken:  "tok_1234567890",
	}

	log.Printf("🔍 Circuit state: %s, Failures: %d\n", cb.State(), cb.Failures())

	response, err := cb.ProcessPayment(ctx, request)

	if err != nil {
		log.Printf("❌ Payment failed: %v\n", err)
		return
	}

	log.Printf("✅ Payment processed successfully!\n")
	log.Printf("   💳 Transaction ID: %s\n", response.TransactionID)
	log.Printf("   💰 Amount: $%.2f %s\n", response.Amount, response.Currency)
	log.Printf("🔍 Circuit state: %s, Failures: %d\n", cb.State(), cb.Failures())
}

func demonstrateCircuitOpening() {
	log.Println("🚨 Circuit Opening Demo")
	log.Println("-----------------------")

	// Create payment service with no initial failure rate
	paymentService, err := service.NewPaymentService(0.0)
	if err != nil {
		log.Fatalf("Failed to create payment service: %v", err)
	}

	// Create circuit breaker with custom configuration
	cb, err := circuitbreaker.New(
		paymentService,
		3,             // failure threshold
		2*time.Second, // timeout
		2,             // max requests in half-open
		2,             // success threshold
	)
	if err != nil {
		log.Fatalf("Failed to create circuit breaker: %v", err)
	}

	ctx := context.Background()

	// Make service unhealthy to trigger failures
	log.Println("💥 Simulating service failures...")
	paymentService.SetHealthy(false)

	request := service.PaymentRequest{
		ID:         "payment-002",
		Amount:     149.99,
		Currency:   "USD",
		MerchantID: "merchant-xyz",
		CardToken:  "tok_9876543210",
	}

	// Trigger failures to open the circuit
	for i := 1; i <= 4; i++ {
		log.Printf("🔍 Attempt %d - Circuit state: %s, Failures: %d\n", i, cb.State(), cb.Failures())

		_, err := cb.ProcessPayment(ctx, request)

		if err != nil {
			if err == circuitbreaker.ErrCircuitOpen {
				log.Printf("🔴 Circuit is OPEN - Request blocked immediately\n")
			} else {
				log.Printf("❌ Payment failed: %v\n", err)
			}
		}

		if cb.State().String() == "Open" && i == 3 {
			log.Printf("🔴 Circuit opened after %d failures!\n", cb.Failures())
		}
	}

	log.Printf("🔍 Final state - Circuit: %s, Failures: %d\n", cb.State(), cb.Failures())

	log.Println()
	log.Println("🔄 Circuit Recovery Demo")
	log.Println("------------------------")

	log.Println("⏳ Waiting for circuit breaker timeout...")
	time.Sleep(3 * time.Second) // Wait longer than the 2-second timeout

	request = service.PaymentRequest{
		ID:         "payment-003",
		Amount:     199.99,
		Currency:   "USD",
		MerchantID: "merchant-recovery",
		CardToken:  "tok_recovery123",
	}

	// First request should transition to half-open but still fail
	log.Printf("🔍 After timeout - Circuit state: %s\n", cb.State())

	// Restore service health and make successful request
	log.Println("🩹 Restoring service health...")
	paymentService.SetHealthy(true)

	log.Println("🔄 Attempting request (should transition to half-open)...")

	_, err = cb.ProcessPayment(ctx, request)
	if err != nil {
		log.Printf("❌ Request failed (circuit half-open): %v\n", err)
	}
	log.Printf("🔍 Circuit state: %s\n", cb.State())

	log.Println("🔄 Making successful request to close circuit...")
	response, err := cb.ProcessPayment(ctx, request)
	if err != nil {
		log.Printf("❌ Recovery attempt failed: %v\n", err)
		return
	}

	log.Printf("✅ Circuit recovered! Payment processed successfully!\n")
	log.Printf("   💳 Transaction ID: %s\n", response.TransactionID)
	log.Printf("   💰 Amount: $%.2f %s\n", response.Amount, response.Currency)
	log.Printf("🔍 Final circuit state: %s, Failures: %d\n", cb.State(), cb.Failures())

	// Test that circuit is fully operational
	log.Println("🧪 Testing circuit is fully operational...")
	for i := 1; i <= 3; i++ {
		testRequest := service.PaymentRequest{
			ID:         fmt.Sprintf("payment-test-%d", i),
			Amount:     50.00,
			Currency:   "USD",
			MerchantID: "merchant-test",
			CardToken:  "tok_test",
		}

		_, err = cb.ProcessPayment(ctx, testRequest)
		if err != nil {
			log.Printf("❌ Test payment %d failed: %v\n", i, err)
		} else {
			log.Printf("✅ Test payment %d successful\n", i)
		}
	}
}
