package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/service"
)

// entry represents a cached item with expiration
type entry struct {
	Value     service.User
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// UserService defines the interface for user operations
type UserService interface {
	GetUser(ctx context.Context, id string) (service.User, error)
}

// cache provides a thread-safe in-memory cache with TTL support
type cache struct {
	service UserService
	lock    sync.RWMutex
	entries map[string]entry
	ttl     time.Duration
}

// New creates a new cache with the specified TTL
func New(service UserService, ttl time.Duration) (*cache, error) {
	if ttl <= 0 {
		return nil, fmt.Errorf("ttl must be greater than 0")
	}

	c := &cache{
		service: service,
		entries: make(map[string]entry),
		ttl:     ttl,
	}

	return c, nil
}

// GetUser retrieves a value from the cache
func (c *cache) GetUser(ctx context.Context, id string) (service.User, error) {

	// Check cache first
	c.lock.RLock()
	cu, ok := c.entries[id]
	c.lock.RUnlock()
	if ok && !cu.IsExpired() {
		return cu.Value, nil // Cache hit & not expired
	}

	// Miss/expired: call underlying service
	user, err := c.service.GetUser(ctx, id)
	if err != nil {
		return service.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache the result with new expiry
	c.lock.Lock()
	c.entries[id] = entry{Value: user, ExpiresAt: time.Now().Add(c.ttl)}
	c.lock.Unlock()

	return user, nil
}
