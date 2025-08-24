# Retry Pattern with Timeouts

The Retry pattern handles transient failures by automatically retrying failed operations with configurable backoff strategies and per-attempt timeouts, improving system resilience against temporary service unavailability.

## Overview

This implementation demonstrates both automatic retries and timeout handling with:
- **Automatic retries**: Configurable maximum attempts with intelligent backoff
- **Per-attempt timeouts**: Individual timeout for each retry attempt
- **Exponential backoff**: Progressive delay increase between retry attempts
- **Timeout enforcement**: Prevents hanging on slow or unresponsive services
- **Context support**: Proper context cancellation and timeout handling
- **Clock injection**: Testable time operations using clockwork

## Key Components

- [`retryClient`](internal/retry/retry.go): Core retry mechanism with exponential backoff and timeouts
- [`orderService`](internal/service/order.go): Mock order service with configurable delays and failure rates  
- [`OrderProcessor`](internal/retry/retry.go): Interface for order processing operations

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │───▶│  Retry Client   │───▶│  Order Service  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌───────────────────────┐
                    │   Retry Strategy      │
                    │ • Timeout per attempt │
                    │ • Exponential backoff │
                    │ • Max attempts        │
                    │ • Context handling    │
                    └───────────────────────┘
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

### Creating a Retry Client with Timeouts
```go
// Create order service with delay and failure rate
orderService, err := service.NewOrderService(100*time.Millisecond, 0.7)
if err != nil {
    log.Fatalf("Failed to create order service: %v", err)
}

// Create retry client with timeout and exponential backoff
retryClient, err := retry.New(
    orderService,
    5,                    // max attempts
    2*time.Second,        // timeout per attempt (prevents hanging)
    100*time.Millisecond, // initial backoff interval
    1*time.Second,        // max backoff interval
    2.0,                  // backoff multiplier
)
if err != nil {
    log.Fatalf("Failed to create retry client: %v", err)
}
```

### Using the Retry Client
```go
ctx := context.Background()
request := service.OrderRequest{
    ID:       "order-001",
    UserID:   "user-123",
    Amount:   99.99,
    Currency: "USD",
    Items: []service.Item{
        {ProductID: "prod-1", Quantity: 2, Price: 29.99},
    },
}

// Process order with automatic retries and timeouts
response, err := retryClient.ProcessOrder(ctx, request)
if err != nil {
    if err == retry.ErrMaxAttemptsExceeded {
        log.Println("Order failed after maximum retry attempts")
    } else if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Order failed due to timeout")
    } else {
        log.Printf("Order failed: %v", err)
    }
    return
}

fmt.Printf("Order successful: %s\n", response.OrderID)
```

### Testing with Custom Clock
```go
// For testing, inject a fake clock
fakeClock := clockwork.NewFakeClock()
retryClient, err := retry.New(
    orderService,
    3,                    // max attempts
    1*time.Second,        // timeout per attempt
    200*time.Millisecond, // initial interval
    800*time.Millisecond, // max interval
    2.0,                  // multiplier
    retry.WithClock(fakeClock),
)

// Control time progression in tests
fakeClock.Advance(200 * time.Millisecond)
```

## Example Output

```
🔄 Retry Pattern with Timeouts Demonstration
============================================

✅ Successful Retry Demo
------------------------
✅ Order processed successfully!
   📦 Order ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890
   💰 Amount: $99.99 USD
   ⏱️  Total time: 347ms (including retries and backoff)

⏰ Exponential Backoff Demo
---------------------------
🔍 Demonstrating backoff delays (service will fail initially):
   Expected delays: 200ms, 400ms, 800ms (capped)
✅ Order eventually succeeded!
   📦 Order ID: b2c3d4e5-f6g7-8901-bcde-f23456789012
   ⏱️  Total time: 2.1s

🚫 Max Attempts Exceeded Demo
-----------------------------
🔍 Attempting order with 3 max attempts (service is down)...
❌ Order failed: Maximum attempts exceeded
   ⏱️  Total time: 1.05s (after 3 attempts with timeouts)

🎉 Retry pattern demonstration complete!
```

## Pattern Benefits

- **Transient failure handling**: Automatically recovers from temporary service issues
- **Timeout protection**: Prevents indefinite blocking on slow or unresponsive services  
- **Intelligent backoff**: Reduces load on struggling services while allowing recovery
- **Resource protection**: Limits total time spent on failing operations
- **Context awareness**: Respects cancellation and deadline requirements
- **Configurable behavior**: Adjustable retry attempts, timeouts, and backoff parameters