package main

import (
	"context"
	"fmt"
	"time"
)

type User struct{ Name string }

func NewCachedUserService(svc UserService, ttl time.Duration) *CacheUserService {
	return &CacheUserService{service: svc, ttl: ttl, cache: make(map[string]cachedEntry)}
}

type UserService interface {
	GetUser(ctx context.Context, id string) (User, error)
}

type cachedEntry struct {
	user User
	exp  time.Time
}

type CacheUserService struct {
	service UserService
	ttl     time.Duration
	cache   map[string]cachedEntry
}

func (c *CacheUserService) GetUser(ctx context.Context, id string) (User, error) {

	// Check cache first
	cu, ok := c.cache[id]
	if ok && time.Now().Before(cu.exp) {
		return cu.user, nil // Cache hit & not expired
	}

	// Miss/expired: call underlying service
	user, err := c.service.GetUser(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache the result with new expiry
	c.cache[id] = cachedEntry{user: user, exp: time.Now().Add(c.ttl)}

	return user, nil
}
