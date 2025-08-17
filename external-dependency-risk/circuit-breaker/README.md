# Circuit Breaker Pattern

The Circuit Breaker pattern prevents cascading failures by monitoring calls to external services and "opening" the circuit when failures exceed a threshold, temporarily blocking further calls.

## Overview

This implementation demonstrates a circuit breaker with two states and context support:
- **Closed**: Normal operation, calls pass through
- **Open**: Circuit is open, calls are blocked after threshold reached
- **Context Support**: Handles timeouts and cancellation gracefully

## Key Components

- `CircuitBreaker`: Core implementation with failure tracking
- `State`: Enum for circuit breaker states (Closed/Open)
- `Call()`: Basic function execution with circuit breaker protection
- `CallWithContext()`: Context-aware function execution
- Configurable failure threshold
- Automatic state transitions

## Context Handling

The circuit breaker intelligently handles context cancellation:
- **Timeout/Cancellation**: Does not count as service failure
- **Actual Failures**: Service errors increment failure counter
- **Early Exit**: Checks context before making calls

## Usage

```bash
# Run the example
make run

# Run tests
make test

# Build binary
make build
```

## API Examples

### Basic Usage
```go
cb := circuitbreaker.New(3)
err := cb.Call(func() error {
    return callExternalService()
})
```

### With Context
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := cb.CallWithContext(ctx, func(ctx context.Context) error {
    return callExternalServiceWithContext(ctx)
})
```

## Example Output

```
Circuit Breaker Example
======================
Threshold: 3 failures

Demo 1: Basic Circuit Breaker
-----------------------------
Call 1 → service down (failures: 1, state: CLOSED)
Call 2 → service down (failures: 2, state: CLOSED)
Call 3 → service down (failures: 3, state: OPEN)
Call 4 → circuit open – skipping call (failures: 3, state: OPEN)
Call 5 → circuit open – skipping call (failures: 3, state: OPEN)

Demo 2: Context with Timeout
----------------------------
Slow service call → context deadline exceeded (failures: 0, state: CLOSED)

Demo 3: Context Cancellation Handling
-------------------------------------
Timeout call 1 → context deadline exceeded (failures: 0, state: CLOSED)
Timeout call 2 → context deadline exceeded (failures: 0, state: CLOSED)
Timeout call 3 → context deadline exceeded (failures: 0, state: CLOSED)

Demo 4: Real Failures After Timeouts
------------------------------------
Failing call 1 → actual service failure (failures: 1, state: CLOSED)
Failing call 2 → actual service failure (failures: 2, state: CLOSED)
Failing call 3 → actual service failure (failures: 3, state: OPEN)
Failing call 4 → circuit open – skipping call (failures: 3, state: OPEN)
```

## Pattern Benefits

- **Fail Fast**: Quickly detect and respond to service failures
- **Resource Protection**: Prevent wasted resources on failing calls
- **System Stability**: Avoid cascading failures across services
- **Recovery Support**: Allow time for failing services to recover
- **Context Awareness**: Proper handling of timeouts and cancellation
- **Smart Failure Detection**: Distinguishes between service failures and client-side cancellation

## Real-World Considerations

This implementation could be extended with:
- **Half-Open State**: Test service recovery periodically
- **Timeout Support**: Automatic circuit reset after time period
- **Metrics**: Detailed monitoring and alerting
- **Fallback Mechanisms**: Alternative responses when circuit is open
- **Concurrent Safety**: Thread-safe operations for production use
