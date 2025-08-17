package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/cache"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/service"
)

func main() {
	fmt.Println("Data Caching Pattern Examples")
	fmt.Println("=============================")
	fmt.Println()

	// Demo 1: Basic cache operations
	basicCacheDemo()
	
	fmt.Println()
	
	// Demo 2: Service caching with performance comparison
	serviceCacheDemo()
	
	fmt.Println()
	
	// Demo 3: Cache invalidation and updates
	cacheInvalidationDemo()
	
	fmt.Println()
	
	// Demo 4: TTL and expiration behavior
	ttlDemo()
}

func basicCacheDemo() {
	fmt.Println("Demo 1: Basic Cache Operations")
	fmt.Println("------------------------------")
	
	cache := cache.New(5 * time.Minute)
	
	// Set some values
	cache.Set("key1", "Hello, World!")
	cache.Set("key2", 42)
	cache.Set("key3", []string{"apple", "banana", "cherry"})
	
	fmt.Printf("Cache size: %d\n", cache.Size())
	fmt.Printf("Cache keys: %v\n", cache.Keys())
	
	// Get values
	if value, found := cache.Get("key1"); found {
		fmt.Printf("key1: %v\n", value)
	}
	
	if value, found := cache.Get("key2"); found {
		fmt.Printf("key2: %v\n", value)
	}
	
	if value, found := cache.Get("nonexistent"); found {
		fmt.Printf("nonexistent: %v\n", value)
	} else {
		fmt.Println("nonexistent: not found")
	}
	
	// Delete a key
	cache.Delete("key2")
	fmt.Printf("After deletion, cache size: %d\n", cache.Size())
}

func serviceCacheDemo() {
	fmt.Println("Demo 2: Service Caching Performance")
	fmt.Println("-----------------------------------")
	
	// Create a slow mock service (500ms delay)
	mockService := service.NewMockUserService(500 * time.Millisecond)
	
	// Create cached service wrapper
	cachedService := cache.NewServiceCache(2*time.Minute, func(ctx context.Context, id string) (service.User, error) {
		fmt.Printf("  → Calling external service for user %s...\n", id)
		return mockService.GetUser(ctx, id)
	})

	ctx := context.Background()
	userID := "1"
	
	fmt.Println("First call (cache miss):")
	start := time.Now()
	user, err := cachedService.Get(ctx, userID)
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("  User: %s (%s)\n", user.Name, user.Email)
	fmt.Printf("  Duration: %v\n", duration)
	
	fmt.Println("\nSecond call (cache hit):")
	start = time.Now()
	user, err = cachedService.Get(ctx, userID)
	duration = time.Since(start)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("  User: %s (%s)\n", user.Name, user.Email)
	fmt.Printf("  Duration: %v\n", duration)
	
	// Show cache stats
	stats := cachedService.Stats()
	fmt.Printf("\nCache stats: %+v\n", stats)
}

func cacheInvalidationDemo() {
	fmt.Println("Demo 3: Cache Invalidation and Updates")
	fmt.Println("--------------------------------------")
	
	mockService := service.NewMockUserService(100 * time.Millisecond)
	cachedService := cache.NewServiceCache(5*time.Minute, func(ctx context.Context, id string) (service.User, error) {
		fmt.Printf("  → Service call for user %s\n", id)
		return mockService.GetUser(ctx, id)
	})
	
	ctx := context.Background()
	userID := "2"
	
	// Initial fetch
	fmt.Println("Initial fetch:")
	user, err := cachedService.Get(ctx, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("  User: %s\n", user.Name)
	
	// Cached fetch
	fmt.Println("\nCached fetch:")
	user, err = cachedService.Get(ctx, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("  User: %s\n", user.Name)
	
	// Update user in service
	fmt.Println("\nUpdating user in service...")
	updatedUser := user
	updatedUser.Name = "Bob Smith Jr."
	updatedUser.Email = "bob.jr@example.com"
	err = mockService.UpdateUser(ctx, updatedUser)
	if err != nil {
		fmt.Printf("Error updating user: %v\n", err)
		return
	}
	
	// Fetch again (still cached, old data)
	fmt.Println("\nFetch after update (still cached):")
	user, err = cachedService.Get(ctx, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("  User: %s\n", user.Name)
	
	// Invalidate cache
	fmt.Println("\nInvalidating cache...")
	cachedService.Invalidate(userID)
	
	// Fetch again (cache miss, fresh data)
	fmt.Println("\nFetch after invalidation:")
	user, err = cachedService.Get(ctx, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("  User: %s\n", user.Name)
}

func ttlDemo() {
	fmt.Println("Demo 4: TTL and Expiration")
	fmt.Println("-------------------------")
	
	// Create cache with short TTL for demo
	shortCache := cache.New(2 * time.Second)
	
	fmt.Println("Setting value with 2-second TTL...")
	shortCache.Set("temp", "This will expire soon")
	
	// Check immediately
	if value, found := shortCache.Get("temp"); found {
		fmt.Printf("Immediately: %v\n", value)
	}
	
	// Wait 1 second
	fmt.Println("Waiting 1 second...")
	time.Sleep(1 * time.Second)
	
	if value, found := shortCache.Get("temp"); found {
		fmt.Printf("After 1s: %v\n", value)
	}
	
	// Wait another 2 seconds (total 3 seconds, should be expired)
	fmt.Println("Waiting 2 more seconds...")
	time.Sleep(2 * time.Second)
	
	if value, found := shortCache.Get("temp"); found {
		fmt.Printf("After 3s: %v\n", value)
	} else {
		fmt.Println("After 3s: expired (not found)")
	}
	
	// Demo custom TTL
	fmt.Println("\nSetting value with custom 1-second TTL...")
	shortCache.SetWithTTL("custom", "Custom TTL value", 1*time.Second)
	
	if value, found := shortCache.Get("custom"); found {
		fmt.Printf("Immediately: %v\n", value)
	}
	
	time.Sleep(1500 * time.Millisecond)
	
	if value, found := shortCache.Get("custom"); found {
		fmt.Printf("After 1.5s: %v\n", value)
	} else {
		fmt.Println("After 1.5s: expired (not found)")
	}
}
