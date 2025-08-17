# Data Caching Pattern

The Data Caching pattern reduces dependency on external services by storing frequently accessed data in memory, improving performance and reducing load on downstream systems.

## Overview

This implementation provides a comprehensive caching solution with:
- **Thread-safe operations**: Concurrent read/write access with proper locking
- **TTL support**: Automatic expiration of cached entries
- **Service wrapper**: Easy integration with existing services
- **Generic support**: Type-safe caching with Go generics
- **Automatic cleanup**: Background removal of expired entries

## Key Components

- `Cache`: Core thread-safe cache with TTL support
- `ServiceCache[T]`: Generic service wrapper for caching function results
- `Entry`: Cache entry with expiration tracking
- Background cleanup goroutine for expired entries

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │───▶│  ServiceCache   │───▶│  External API   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Cache Store   │
                       │  (In-Memory)    │
                       └─────────────────┘
```

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

### Basic Cache Operations
```go
cache := cache.New(5 * time.Minute)

// Store values
cache.Set("user:123", user)
cache.SetWithTTL("temp", data, 30*time.Second)

// Retrieve values
if value, found := cache.Get("user:123"); found {
    user := value.(User)
    // Use user data
}

// Management
cache.Delete("user:123")
cache.Clear()
fmt.Printf("Cache size: %d\n", cache.Size())
```

### Service Caching
```go
// Wrap any service function
cachedService := cache.NewServiceCache(2*time.Minute, 
    func(ctx context.Context, id string) (User, error) {
        return userService.GetUser(ctx, id)
    })

// Use like normal service - caching is transparent
user, err := cachedService.Get(ctx, "123")

// Invalidate when data changes
cachedService.Invalidate("123")

// View cache statistics
stats := cachedService.Stats()
```

## Example Output

```
Data Caching Pattern Examples
=============================

Demo 1: Basic Cache Operations
------------------------------
Cache size: 3
Cache keys: [key1 key2 key3]
key1: Hello, World!
key2: 42
nonexistent: not found
After deletion, cache size: 2

Demo 2: Service Caching Performance
-----------------------------------
First call (cache miss):
  → Calling external service for user 1...
  User: Alice Johnson (alice@example.com)
  Duration: 501ms

Second call (cache hit):
  User: Alice Johnson (alice@example.com)
  Duration: 45µs

Cache stats: map[keys:[1] size:1]

Demo 3: Cache Invalidation and Updates
--------------------------------------
Initial fetch:
  → Service call for user 2
  User: Bob Smith

Cached fetch:
  User: Bob Smith

Updating user in service...

Fetch after update (still cached):
  User: Bob Smith

Invalidating cache...

Fetch after invalidation:
  → Service call for user 2
  User: Bob Smith Jr.

Demo 4: TTL and Expiration
-------------------------
Setting value with 2-second TTL...
Immediately: This will expire soon
Waiting 1 second...
After 1s: This will expire soon
Waiting 2 more seconds...
After 3s: expired (not found)

Setting value with custom 1-second TTL...
Immediately: Custom TTL value
After 1.5s: expired (not found)
```

## Pattern Benefits

- **Performance**: Dramatically reduce response times for repeated requests
- **Resilience**: Continue serving cached data when external services are unavailable
- **Load Reduction**: Decrease load on downstream services and databases
- **Cost Efficiency**: Reduce API calls and associated costs
- **User Experience**: Faster response times improve user satisfaction

## Cache Strategies

### Cache-Aside (Lazy Loading)
```go
func GetUser(ctx context.Context, id string) (User, error) {
    // Check cache first
    if user, found := cache.Get(id); found {
        return user.(User), nil
    }
    
    // Load from database
    user, err := db.GetUser(ctx, id)
    if err != nil {
        return User{}, err
    }
    
    // Store in cache
    cache.Set(id, user)
    return user, nil
}
```

### Write-Through
```go
func UpdateUser(ctx context.Context, user User) error {
    // Update database
    if err := db.UpdateUser(ctx, user); err != nil {
        return err
    }
    
    // Update cache
    cache.Set(user.ID, user)
    return nil
}
```

### Write-Behind (Write-Back)
```go
func UpdateUser(ctx context.Context, user User) error {
    // Update cache immediately
    cache.Set(user.ID, user)
    
    // Schedule async database update
    go func() {
        db.UpdateUser(context.Background(), user)
    }()
    
    return nil
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

The implementation includes comprehensive examples demonstrating:
- Basic cache operations and lifecycle
- Performance comparison (cached vs uncached)
- Cache invalidation and data consistency
- TTL behavior and expiration handling
- Service integration patterns

Run `make test` to execute the test suite with race detection and coverage reporting.
