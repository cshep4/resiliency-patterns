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

- `circuitBreaker`: Core thread-safe circuit breaker implementation
- `PaymentService`: Interface for payment processing operations
- `paymentService`: Controllable payment service with configurable failure rates
- `State`: Circuit breaker states (Closed, Open, HalfOpen)
- Clock abstraction for testable time operations

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
// Create a payment service with delay and failure rate
paymentService, err := service.NewPaymentService(200*time.Millisecond, 0.1)
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

### Monitoring Circuit State
```go
// Check current state
fmt.Printf("Circuit state: %s\n", circuitBreaker.State())
fmt.Printf("Failure count: %d\n", circuitBreaker.Failures())

// State predicates
if circuitBreaker.IsOpen() {
    fmt.Println("Circuit is open - requests will fail fast")
} else if circuitBreaker.IsHalfOpen() {
    fmt.Println("Circuit is half-open - limited requests allowed")
} else {
    fmt.Println("Circuit is closed - normal operation")
}
```

### Testing with Custom Clock
```go
// For testing, inject a fake clock
fakeClock := clockwork.NewFakeClock()
cb, err := circuitbreaker.New(
    paymentService, 
    circuitbreaker.WithClock(fakeClock),
    circuitbreaker.WithTimeout(30*time.Second),
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
   ⏱️  Processing time: 203.4ms
🔍 Circuit state: Closed, Failures: 0

🚨 Circuit Opening Demo
-----------------------
💥 Simulating service failures...
🔍 Attempt 1 - Circuit state: Closed, Failures: 0
❌ Payment failed: payment processing failed: payment service unavailable for request payment-002 (took 201.2ms)
🔍 Attempt 2 - Circuit state: Closed, Failures: 1
❌ Payment failed: payment processing failed: payment service unavailable for request payment-002 (took 203.8ms)
🔍 Attempt 3 - Circuit state: Closed, Failures: 2
❌ Payment failed: payment processing failed: payment service unavailable for request payment-002 (took 202.1ms)
🔴 Circuit opened after 3 failures!
🔍 Attempt 4 - Circuit state: Open, Failures: 3
🔴 Circuit is OPEN - Request blocked immediately (took 45.2µs)
🔍 Final state - Circuit: Open, Failures: 3

🔄 Circuit Recovery Demo
------------------------
⏳ Waiting for circuit breaker timeout...
🔍 After timeout - Circuit state: Open
🔄 Attempting request (should transition to half-open)...
❌ Request failed (circuit half-open): payment processing failed: payment service unavailable for request payment-003
🔍 Circuit state: HalfOpen
🩹 Restoring service health...
🔄 Making successful request to close circuit...
✅ Circuit recovered! Payment processed successfully!
   💳 Transaction ID: b2c3d4e5-f6g7-8901-bcde-f23456789012
   💰 Amount: $199.99 USD
   ⏱️  Processing time: 201.7ms
🔍 Final circuit state: Closed, Failures: 0
🧪 Testing circuit is fully operational...
✅ Test payment 1 successful
✅ Test payment 2 successful
✅ Test payment 3 successful

🎉 Circuit breaker demonstration complete!
```

## Circuit Breaker States

### Closed State
- **Normal Operation**: All requests pass through to the service
- **Failure Tracking**: Counts consecutive failures
- **State Transition**: Opens when failure threshold is reached

### Open State
- **Fast Fail**: Requests fail immediately without calling service
- **Timeout**: Waits for configured timeout period
- **State Transition**: Transitions to Half-Open after timeout

### Half-Open State
- **Limited Requests**: Allows configured number of requests to test service
- **Recovery Testing**: Monitors success/failure of test requests
- **State Transition**: 
  - Closes on successful request
  - Opens immediately on any failure

## Configuration Options

```go
circuitbreaker.New(service,
    // Number of consecutive failures to trigger opening (default: 5)
    circuitbreaker.WithFailureThreshold(3),
    
    // Time to wait before transitioning from Open to Half-Open (default: 30s)
    circuitbreaker.WithTimeout(10*time.Second),
    
    // Max requests allowed in Half-Open state (default: 3)
    circuitbreaker.WithMaxRequests(2),
    
    // Custom clock for testing (default: real clock)
    circuitbreaker.WithClock(fakeClock),
)
```

## Pattern Benefits

- **Cascading Failure Prevention**: Stops failures from propagating through the system
- **Fast Fail**: Provides immediate failure response when service is down
- **Automatic Recovery**: Self-healing capability when service recovers
- **Resource Protection**: Prevents resource exhaustion from repeated failed calls
- **Monitoring**: Provides visibility into service health and failure patterns
- **Graceful Degradation**: Allows applications to handle service outages gracefully

## Real-World Considerations

### Production Enhancements
- **Metrics Collection**: Track open/close events, failure rates, and response times
- **Alerting**: Notify operations when circuits open frequently
- **Fallback Strategies**: Implement cached responses or alternative services
- **Circuit Breaker Registry**: Centralized management of multiple circuit breakers
- **Configuration Management**: Runtime configuration updates without restarts

### Integration Patterns
- **API Gateway Integration**: Implement circuit breakers at the gateway level
- **Service Mesh**: Use with Istio/Envoy for automatic circuit breaking
- **Database Connections**: Protect database connection pools
- **External APIs**: Wrap third-party service calls

### Monitoring and Observability
- **State Changes**: Log all state transitions with timestamps
- **Failure Classification**: Distinguish between different types of failures
- **Recovery Time**: Track how long services take to recover
- **Success Rate**: Monitor success rates in Half-Open state

## Testing

The implementation includes comprehensive test coverage with:

### Test Features
- **Mock Dependencies**: Uses `gomock` for service mocking
- **Time Control**: Injects fake clocks for deterministic timeout testing
- **State Transitions**: Tests all state transition scenarios
- **Error Scenarios**: Tests various failure conditions
- **Concurrency Safety**: Race detection enabled
- **Configuration Validation**: Tests all configuration options

### Key Test Cases
- Circuit breaker creation with valid/invalid parameters
- Normal operation in Closed state
- State transitions (Closed → Open → Half-Open → Closed)
- Fast-fail behavior in Open state
- Limited request handling in Half-Open state
- Automatic recovery scenarios
- Context cancellation and timeout handling
- Concurrent access safety

### Running Tests
```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run with race detection
make test-race

# Run with coverage
make test-cover
```

The test suite demonstrates best practices for testing stateful components with time dependencies and concurrent access patterns.