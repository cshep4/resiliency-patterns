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

- [`cache`](internal/cache/cache.go): Core thread-safe cache with TTL support for user data
- [`entry`](internal/cache/cache.go): Cache entry with expiration tracking
- [`userService`](internal/service/user.go): Mock user service with configurable delay to simulate a network call

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
    log.Fatalf("Error: %v", err)
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
