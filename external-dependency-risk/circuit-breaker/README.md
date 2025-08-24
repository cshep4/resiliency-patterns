# Circuit Breaker Pattern

The Circuit Breaker pattern prevents cascading failures in distributed systems by monitoring service calls and "opening the circuit" when failure rates exceed a threshold, providing fast-fail behavior and automatic recovery.

## Overview

This implementation provides a payment service circuit breaker with:
- **Thread-safe operations**: Concurrent access with proper locking
- **State management**: Closed, Open, and Half-Open states with proper transitions
- **Configurable thresholds**: Failure count, timeout duration, and max requests
- **Clock injection**: Testable time operations using clockwork
- **Comprehensive monitoring**: State inspection and failure tracking

## Key Components

- [`circuitBreaker`](internal/circuitbreaker/circuitbreaker.go): Core thread-safe circuit breaker implementation
- [`paymentService`](internal/service/payment.go): Mock payment service with configurable failure rates
- [`State`](internal/circuitbreaker/circuitbreaker.go): Circuit breaker states (Closed, Open, HalfOpen)

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │───▶│ Circuit Breaker │───▶│ Payment Service │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  State Machine  │
                       │ Closed/Open/    │
                       │   Half-Open     │
                       └─────────────────┘
```

## Usage

```bash
# Run the example
make run

# Generate mocks
make mocks

# Run tests
make test

# Build binary
make build
```

## API Examples

### Creating a Circuit Breaker
```go
// Create a payment service with a failure rate
paymentService, err := service.NewPaymentService(0.1)
if err != nil {
    log.Fatalf("Failed to create payment service: %v", err)
}

// Create circuit breaker with custom configuration
circuitBreaker, err := circuitbreaker.New(
    paymentService,
    3,                // failure threshold
    5*time.Second,    // timeout
    2,                // max requests in half-open
    1,                // success threshold
)
if err != nil {
    log.Fatalf("Failed to create circuit breaker: %v", err)
}
```

### Using the Circuit Breaker
```go
ctx := context.Background()
request := service.PaymentRequest{
    ID:        "payment-001",
    Amount:    99.99,
    Currency:  "USD",
    MerchantID: "merchant-abc",
    CardToken:  "tok_1234567890",
}

// Process payment through circuit breaker
response, err := circuitBreaker.ProcessPayment(ctx, request)
if err != nil {
    if err == circuitbreaker.ErrCircuitOpen {
        log.Println("Circuit is open - failing fast")
    } else {
        log.Printf("Payment failed: %v", err)
    }
    return
}

fmt.Printf("Payment successful: %s\n", response.TransactionID)
```

### Testing with Custom Clock
```go
// For testing, inject a fake clock
fakeClock := clockwork.NewFakeClock()
cb, err := circuitbreaker.New(
    paymentService,
    3,                // failure threshold
    30*time.Second,   // timeout
    2,                // max requests in half-open
    1,                // success threshold
    circuitbreaker.WithClock(fakeClock),
)

// Advance time to test timeout behavior
fakeClock.Advance(31 * time.Second)
```

## Example Output

```
🔌 Circuit Breaker Demonstration
================================

✅ Normal Operation Demo
------------------------
🔍 Circuit state: Closed, Failures: 0
✅ Payment processed successfully!
   💳 Transaction ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890
   💰 Amount: $99.99 USD
🔍 Circuit state: Closed, Failures: 0

🚨 Circuit Opening Demo
-----------------------
💥 Simulating service failures...
🔍 Attempt 1 - Circuit state: Closed, Failures: 0
❌ Payment failed: payment processing failed: payment service unavailable for request payment-002
🔍 Attempt 2 - Circuit state: Closed, Failures: 1
❌ Payment failed: payment processing failed: payment service unavailable for request payment-002
🔍 Attempt 3 - Circuit state: Closed, Failures: 2
❌ Payment failed: payment processing failed: payment service unavailable for request payment-002
🔴 Circuit opened after 3 failures!
🔍 Attempt 4 - Circuit state: Open, Failures: 3
🔴 Circuit is OPEN - Request blocked immediately
🔍 Final state - Circuit: Open, Failures: 3

🔄 Circuit Recovery Demo
------------------------
⏳ Waiting for circuit breaker timeout...
🔍 After timeout - Circuit state: Open
🩹 Restoring service health...
🔄 Attempting request (should transition to half-open)...
🔍 Circuit state: HalfOpen
🔄 Making successful request to close circuit...
✅ Circuit recovered! Payment processed successfully!
   💳 Transaction ID: b2c3d4e5-f6g7-8901-bcde-f23456789012
   💰 Amount: $199.99 USD
🔍 Final circuit state: Closed, Failures: 0
🧪 Testing circuit is fully operational...
✅ Test payment 1 successful
✅ Test payment 2 successful
✅ Test payment 3 successful

🎉 Circuit breaker demonstration complete!
```