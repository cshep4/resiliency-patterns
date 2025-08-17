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

// UserService defines the interface for user operations
type UserService interface {
	GetUser(ctx context.Context, id string) (User, error)
	CreateUser(ctx context.Context, user User) error
	UpdateUser(ctx context.Context, user User) error
	DeleteUser(ctx context.Context, id string) error
}

// MockUserService simulates a slow external user service
type MockUserService struct {
	users map[string]User
	delay time.Duration
}

// NewMockUserService creates a new mock user service
func NewMockUserService(delay time.Duration) *MockUserService {
	users := map[string]User{
		"1": {ID: "1", Name: "Alice Johnson", Email: "alice@example.com", Created: time.Now().Add(-24 * time.Hour)},
		"2": {ID: "2", Name: "Bob Smith", Email: "bob@example.com", Created: time.Now().Add(-12 * time.Hour)},
		"3": {ID: "3", Name: "Charlie Brown", Email: "charlie@example.com", Created: time.Now().Add(-6 * time.Hour)},
		"4": {ID: "4", Name: "Diana Prince", Email: "diana@example.com", Created: time.Now().Add(-3 * time.Hour)},
		"5": {ID: "5", Name: "Eve Wilson", Email: "eve@example.com", Created: time.Now().Add(-1 * time.Hour)},
	}
	
	return &MockUserService{
		users: users,
		delay: delay,
	}
}

// GetUser retrieves a user by ID with simulated delay
func (s *MockUserService) GetUser(ctx context.Context, id string) (User, error) {
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

// CreateUser creates a new user
func (s *MockUserService) CreateUser(ctx context.Context, user User) error {
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	if _, exists := s.users[user.ID]; exists {
		return fmt.Errorf("user with id %s already exists", user.ID)
	}
	
	user.Created = time.Now()
	s.users[user.ID] = user
	return nil
}

// UpdateUser updates an existing user
func (s *MockUserService) UpdateUser(ctx context.Context, user User) error {
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	if _, exists := s.users[user.ID]; !exists {
		return fmt.Errorf("user with id %s not found", user.ID)
	}
	
	// Preserve creation time
	existing := s.users[user.ID]
	user.Created = existing.Created
	s.users[user.ID] = user
	return nil
}

// DeleteUser removes a user
func (s *MockUserService) DeleteUser(ctx context.Context, id string) error {
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	if _, exists := s.users[id]; !exists {
		return fmt.Errorf("user with id %s not found", id)
	}
	
	delete(s.users, id)
	return nil
}

// ListUsers returns all users (for demo purposes)
func (s *MockUserService) ListUsers() []User {
	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}
