package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
	
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/service"
)

// entry represents a cached item with expiration
type entry struct {
	Value     service.User
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e entry) IsExpired(clock clockwork.Clock) bool {
	return clock.Now().After(e.ExpiresAt)
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
	clock   clockwork.Clock
}

// Option is a functional option for configuring the cache
type Option func(*cache) error

// WithClock sets a custom clock for the cache
func WithClock(clock clockwork.Clock) Option {
	return func(c *cache) error {
		if clock == nil {
			return errors.New("clock is nil")
		}
		c.clock = clock
		return nil
	}
}

// New creates a new cache with the specified TTL and optional configurations
func New(service UserService, ttl time.Duration, opts ...Option) (*cache, error) {
	switch {
	case service == nil:
		return nil, errors.New("service is nil")
	case ttl <= 0:
		return nil, errors.New("ttl must be greater than 0")
	}

	c := &cache{
		service: service,
		entries: make(map[string]entry),
		ttl:     ttl,
		clock:   clockwork.NewRealClock(), // Default to real clock
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// GetUser retrieves a value from the cache
func (c *cache) GetUser(ctx context.Context, id string) (service.User, error) {

	// Check cache first
	c.lock.RLock()
	cu, ok := c.entries[id]
	c.lock.RUnlock()
	if ok && !cu.IsExpired(c.clock) {
		return cu.Value, nil // Cache hit & not expired
	}

	// Miss/expired: call underlying service
	user, err := c.service.GetUser(ctx, id)
	if err != nil {
		return service.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache the result with new expiry
	c.lock.Lock()
	c.entries[id] = entry{Value: user, ExpiresAt: c.clock.Now().Add(c.ttl)}
	c.lock.Unlock()

	return user, nil
}
