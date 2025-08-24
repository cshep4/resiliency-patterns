package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// User represents a user entity
type User struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Created  time.Time `json:"created"`
}

// userService simulates a slow external user service
type userService struct {
	users map[string]User
	delay time.Duration
}

// NewUserService creates a new user service
func NewUserService(delay time.Duration) (*userService, error) {
	if delay < 0 {
		return nil, errors.New("delay must be greater than 0")
	}

	users := map[string]User{
		"1": {ID: "1", Name: "Alice Johnson", Email: "alice@example.com", Created: time.Now().Add(-24 * time.Hour)},
		"2": {ID: "2", Name: "Bob Smith", Email: "bob@example.com", Created: time.Now().Add(-12 * time.Hour)},
		"3": {ID: "3", Name: "Charlie Brown", Email: "charlie@example.com", Created: time.Now().Add(-6 * time.Hour)},
		"4": {ID: "4", Name: "Diana Prince", Email: "diana@example.com", Created: time.Now().Add(-3 * time.Hour)},
		"5": {ID: "5", Name: "Eve Wilson", Email: "eve@example.com", Created: time.Now().Add(-1 * time.Hour)},
	}
	
	s := &userService{
		users: users,
		delay: delay,
	}

	return s, nil
}

// GetUser retrieves a user by ID with simulated delay
func (s *userService) GetUser(ctx context.Context, id string) (User, error) {
	// Simulate network delay
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return User{}, ctx.Err()
	}
	
	// Simulate occasional failures
	if rand.Float32() < 0.1 { // 10% failure rate
		return User{}, errors.New("service temporarily unavailable")
	}
	
	user, exists := s.users[id]
	if !exists {
		return User{}, fmt.Errorf("user with id %s not found", id)
	}
	
	return user, nil
}