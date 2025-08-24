package main

import (
	"context"
	"log"
	"time"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/cache"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/service"
)

func main() {
	log.Println("🚀 Cache Demonstration")
	log.Println("======================")

	// Create a slow user service (simulating external dependency)
	userService, err := service.NewUserService(500 * time.Millisecond)
	if err != nil {
		log.Fatalf("Failed to create user service: %v", err)
	}

	// Create cache with 30 second TTL
	userCache, err := cache.New(userService, 30*time.Second)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	ctx := context.Background()

	log.Println()

	// Demonstrate cache miss and hit scenarios
	demonstrateCacheHit(ctx, userCache)

	log.Println()
	
	// Demonstrate performance benefits
	demonstratePerformance(ctx, userCache)

	log.Println()
	
	// Demonstrate TTL expiration with shorter TTL cache
	demonstrateTTLExpiration()

	log.Println()
	log.Println("🎉 Cache demonstration complete!")
}

func demonstrateCacheHit(ctx context.Context, userCache cache.UserService) {
	log.Println("📊 Cache Miss vs Cache Hit Demo")
	log.Println("--------------------------------")
	
	userID := "1"
	
	// First call - cache miss
	log.Printf("🔍 First call (cache miss) for user %s...\n", userID)
	start := time.Now()
	user, err := userCache.GetUser(ctx, userID)
	duration := time.Since(start)
	
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return
	}
	
	log.Printf("✅ Retrieved user: %s (%s) in %v\n", user.Name, user.Email, duration)
	
	// Second call - cache hit
	log.Printf("🔍 Second call (cache hit) for user %s...\n", userID)
	start = time.Now()
	user, err = userCache.GetUser(ctx, userID)
	duration = time.Since(start)
	
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return
	}
	
	log.Printf("⚡ Retrieved user: %s (%s) in %v (from cache!)\n", user.Name, user.Email, duration)
}

func demonstratePerformance(ctx context.Context, userCache cache.UserService) {
	log.Println("🏎️  Performance Comparison")
	log.Println("---------------------------")
	
	userIDs := []string{"2", "3", "4", "5"}
	
	// Warm up the cache
	log.Println("🔥 Warming up cache...")
	for _, id := range userIDs {
		_, err := userCache.GetUser(ctx, id)
		if err != nil {
			log.Printf("Error getting user, retrying: %s: %v", id, err)
			_, err = userCache.GetUser(ctx, id)
			if err != nil {
				log.Printf("Error getting user, skipping: %s: %v", id, err)
				continue
			}
		}
	}
	
	// Benchmark cached requests
	log.Printf("⏱️  Fetching %d users from cache...\n", len(userIDs))
	start := time.Now()
	
	for _, id := range userIDs {
		user, err := userCache.GetUser(ctx, id)
		if err != nil {
			log.Printf("Error getting user %s: %v", id, err)
			continue
		}
		log.Printf("   📋 %s: %s\n", user.ID, user.Name)
	}
	
	totalDuration := time.Since(start)
	avgDuration := totalDuration / time.Duration(len(userIDs))
	
	log.Printf("🎯 Total time: %v (avg: %v per user)\n", totalDuration, avgDuration)
	log.Printf("💡 Without cache, this would take ~%v (500ms per user)\n", 
		time.Duration(len(userIDs))*500*time.Millisecond)
}

func demonstrateTTLExpiration() {
	log.Println("⏰ TTL Expiration Demo")
	log.Println("----------------------")
	
	// Create service and cache with very short TTL for demo
	userService, err := service.NewUserService(100 * time.Millisecond)
	if err != nil {
		log.Fatalf("Failed to create user service: %v", err)
	}
	
	shortTTLCache, err := cache.New(userService, 2*time.Second)
	if err != nil {
		log.Fatalf("Failed to create short TTL cache: %v", err)
	}
	
	ctx := context.Background()
	userID := "1"
	
	// First call
	log.Printf("🔍 Initial call for user %s...\n", userID)
	start := time.Now()
	user, err := shortTTLCache.GetUser(ctx, userID)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	log.Printf("✅ Got %s in %v\n", user.Name, time.Since(start))
	
	// Immediate second call (cache hit)
	log.Printf("🔍 Immediate second call (should be cached)...\n")
	start = time.Now()
	user, err = shortTTLCache.GetUser(ctx, userID)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	log.Printf("⚡ Got %s in %v (cached)\n", user.Name, time.Since(start))
	
	// Wait for TTL to expire
	log.Printf("⏳ Waiting for TTL to expire (2 seconds)...\n")
	time.Sleep(2100 * time.Millisecond)
	
	// Third call after expiration
	log.Printf("🔍 Call after TTL expiration...\n")
	start = time.Now()
	user, err = shortTTLCache.GetUser(ctx, userID)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	log.Printf("🔄 Got %s in %v (cache expired, fetched fresh)\n", user.Name, time.Since(start))	
}