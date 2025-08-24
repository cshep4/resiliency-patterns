# Data Caching Pattern

The Data Caching pattern reduces dependency on external services by storing frequently accessed data in memory, improving performance and reducing load on downstream systems.

## Overview

This implementation provides a user service caching solution with:
- **Thread-safe operations**: Concurrent read/write access with proper locking
- **TTL support**: Automatic expiration of cached entries
- **Service wrapper**: Easy integration with existing user services
- **Clock injection**: Testable time operations using clockwork
- **Error handling**: Proper error propagation and wrapping

## Key Components

- `cache`: Core thread-safe cache with TTL support for user data
- `UserService`: Interface for user operations
- `entry`: Cache entry with expiration tracking
- `service.userService`: Mock user service with configurable delay
- Clock abstraction for testable time operations

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │───▶│   User Cache    │───▶│  User Service   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Cache Store   │
                       │  (In-Memory)    │
                       │   TTL-based     │
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

### Creating a Cache
```go
// Create a user service with 500ms delay to simulate slow external dependency
userService, err := service.NewUserService(500 * time.Millisecond)
if err != nil {
    log.Fatalf("Failed to create user service: %v", err)
}

// Create cache with 30 second TTL
userCache, err := cache.New(userService, 30*time.Second)
if err != nil {
    log.Fatalf("Failed to create cache: %v", err)
}
```

### Using the Cache
```go
ctx := context.Background()

// First call - cache miss, calls underlying service
user, err := userCache.GetUser(ctx, "1")
if err != nil {
    log.Printf("Error: %v", err)
    return
}
fmt.Printf("User: %s (%s)\n", user.Name, user.Email)

// Second call - cache hit, returns immediately
user, err = userCache.GetUser(ctx, "1")
// This call is much faster!
```

### Testing with Custom Clock
```go
// For testing, inject a fake clock
fakeClock := clockwork.NewFakeClock()
userCache, err := cache.New(userService, 10*time.Minute, cache.WithClock(fakeClock))

// Advance time to test expiration
fakeClock.Advance(11 * time.Minute)
```

## Example Output

```
🚀 Cache Demonstration
======================

📊 Cache Miss vs Cache Hit Demo
--------------------------------
🔍 First call (cache miss) for user 1...
✅ Retrieved user: Alice Johnson (alice@example.com) in 503.2ms
🔍 Second call (cache hit) for user 1...
⚡ Retrieved user: Alice Johnson (alice@example.com) in 12.7µs (from cache!)

🏎️  Performance Comparison
---------------------------
🔥 Warming up cache...
⏱️  Fetching 4 users from cache...
   📋 2: Bob Smith
   📋 3: Charlie Brown
   📋 4: Diana Wilson
   📋 5: Eve Wilson
🎯 Total time: 127.3µs (avg: 31.8µs per user)
💡 Without cache, this would take ~2s (500ms per user)

⏰ TTL Expiration Demo
----------------------
🔍 Initial call for user 1...
✅ Got Alice Johnson in 502.1ms
🔍 Immediate second call (should be cached)...
⚡ Got Alice Johnson in 8.4µs (cached)
⏳ Waiting for TTL to expire (2 seconds)...
🔍 Call after TTL expiration...
🔄 Got Alice Johnson in 501.8ms (cache expired, fetched fresh)

🎉 Cache demonstration complete!
```

## Pattern Benefits

- **Performance**: Dramatically reduce response times for repeated requests
- **Resilience**: Continue serving cached data when external services are unavailable
- **Load Reduction**: Decrease load on downstream services and databases
- **Cost Efficiency**: Reduce API calls and associated costs
- **User Experience**: Faster response times improve user satisfaction

## Cache Strategies

This implementation uses the **Cache-Aside (Lazy Loading)** pattern where:

1. **Cache Miss**: When data isn't in cache, fetch from service and store in cache
2. **Cache Hit**: When data exists and isn't expired, return from cache
3. **TTL Expiration**: Cached data automatically expires after the configured TTL

```go
// The cache implementation handles this pattern automatically
func (c *cache) GetUser(ctx context.Context, id string) (service.User, error) {
    // Check cache first
    c.lock.RLock()
    cu, ok := c.entries[id]
    c.lock.RUnlock()
    if ok && !cu.IsExpired(c.clock) {
        return cu.Value, nil // Cache hit & not expired
    }

    // Cache miss or expired - fetch from service
    user, err := c.service.GetUser(ctx, id)
    if err != nil {
        return service.User{}, fmt.Errorf("failed to get user %s: %w", id, err)
    }

    // Cache the result with new expiry
    c.lock.Lock()
    c.entries[id] = entry{Value: user, ExpiresAt: c.clock.Now().Add(c.ttl)}
    c.lock.Unlock()

    return user, nil
}
```

### Extension Patterns

For production use, you might extend this to support:

**Write-Through Pattern**
```go
// Update service and cache simultaneously
func (c *cache) UpdateUser(ctx context.Context, user service.User) error {
    if err := c.service.UpdateUser(ctx, user); err != nil {
        return err
    }
    
    // Update cache with fresh TTL
    c.lock.Lock()
    c.entries[user.ID] = entry{Value: user, ExpiresAt: c.clock.Now().Add(c.ttl)}
    c.lock.Unlock()
    
    return nil
}
```

**Cache Invalidation**
```go
func (c *cache) InvalidateUser(id string) {
    c.lock.Lock()
    delete(c.entries, id)
    c.lock.Unlock()
}
```

## Real-World Considerations

### Production Enhancements
- **Distributed Caching**: Redis, Memcached for multi-instance deployments
- **Cache Warming**: Pre-populate cache with frequently accessed data
- **Metrics & Monitoring**: Hit/miss ratios, eviction rates, memory usage
- **Circuit Breaker Integration**: Fallback to cache when services fail
- **Compression**: Reduce memory usage for large cached objects

### Cache Invalidation Strategies
- **TTL-based**: Automatic expiration (implemented)
- **Event-driven**: Invalidate on data changes
- **Manual**: Explicit invalidation via API
- **Tag-based**: Group related cache entries for bulk invalidation

### Memory Management
- **Size Limits**: Implement LRU/LFU eviction policies
- **Memory Monitoring**: Track cache memory usage
- **Graceful Degradation**: Handle out-of-memory conditions

## Testing

The implementation includes comprehensive test coverage with:

### Test Features
- **Mock Dependencies**: Uses `gomock` for service mocking
- **Time Control**: Injects fake clocks for deterministic TTL testing  
- **Error Scenarios**: Tests service failures and context cancellation
- **Concurrency Safety**: Race detection enabled
- **Edge Cases**: Nil services, invalid TTLs, expired entries

### Key Test Cases
- Cache creation with valid/invalid parameters
- Cache miss/hit scenarios 
- TTL expiration behavior with fake clocks
- Service error propagation
- Context cancellation handling
- Functional options (custom clock injection)

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

The test suite uses `testify/require` for assertions and demonstrates best practices for testing time-dependent code with clock injection.
