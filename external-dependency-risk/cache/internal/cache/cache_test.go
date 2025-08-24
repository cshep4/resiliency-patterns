package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/cache"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/mocks"
	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/service"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("valid TTL", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, 5*time.Minute)
		require.NoError(t, err)
		require.NotNil(t, c)
	})	

	t.Run("nil service", func(t *testing.T) {
		c, err := cache.New(nil, 5*time.Minute)
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "service is nil")
	})

	t.Run("invalid TTL", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, 0)
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "ttl must be greater than 0")
	})

	t.Run("negative TTL", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, -time.Minute)
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "ttl must be greater than 0")
	})

	t.Run("nil clock", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, 5*time.Minute, cache.WithClock(nil))
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "clock is nil")
	})
}

func TestNewWithOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("with custom clock", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		fakeClock := clockwork.NewFakeClock()
		c, err := cache.New(mockService, 5*time.Minute, cache.WithClock(fakeClock))
		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("with nil clock option", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, 5*time.Minute, cache.WithClock(nil))
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "clock is nil")
	})

	t.Run("without options - uses real clock", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, 5*time.Minute)
		require.NoError(t, err)
		require.NotNil(t, c)
	})
}

func TestGetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedUser := service.User{
		ID:      "1",
		Name:    "Test User",
		Email:   "test@example.com",
		Created: time.Now(),
	}

	t.Run("cache miss - service success", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, 5*time.Minute)
		require.NoError(t, err)

		ctx := context.Background()

		mockService.EXPECT().
			GetUser(ctx, "1").
			Return(expectedUser, nil).
			Times(1)

		user, err := c.GetUser(ctx, "1")
		require.NoError(t, err)
		require.Equal(t, expectedUser, user)
	})

	t.Run("cache hit - service not called", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, 5*time.Minute)
		require.NoError(t, err)

		ctx := context.Background()

		mockService.EXPECT().
			GetUser(ctx, "1").
			Return(expectedUser, nil).
			Times(1)

		// First call - cache miss
		user1, err := c.GetUser(ctx, "1")
		require.NoError(t, err)
		require.Equal(t, expectedUser, user1)

		// Second call - cache hit (service should not be called again)
		user2, err := c.GetUser(ctx, "1")
		require.NoError(t, err)
		require.Equal(t, expectedUser, user2)
	})

	t.Run("cache expired - service called again", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		fakeClock := clockwork.NewFakeClock()
		c, err := cache.New(mockService, 10*time.Minute, cache.WithClock(fakeClock))
		require.NoError(t, err)

		updatedUser := expectedUser
		updatedUser.Name = "Updated User"

		ctx := context.Background()

		mockService.EXPECT().
			GetUser(ctx, "1").
			Return(expectedUser, nil).
			Times(1)

		mockService.EXPECT().
			GetUser(ctx, "1").
			Return(updatedUser, nil).
			Times(1)

		// First call - cache miss
		user1, err := c.GetUser(ctx, "1")
		require.NoError(t, err)
		require.Equal(t, expectedUser, user1)

		// Advance time to expire cache
		fakeClock.Advance(11 * time.Minute)

		// Second call - cache expired
		user2, err := c.GetUser(ctx, "1")
		require.NoError(t, err)
		require.Equal(t, updatedUser, user2)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, 5*time.Minute)
		require.NoError(t, err)

		ctx := context.Background()

		serviceErr := errors.New("service unavailable")
		mockService.EXPECT().
			GetUser(ctx, "1").
			Return(service.User{}, serviceErr).
			Times(1)

		user, err := c.GetUser(ctx, "1")
		require.Error(t, err)
		require.Equal(t, service.User{}, user)
		require.Contains(t, err.Error(), "failed to get user")
		require.Contains(t, err.Error(), "service unavailable")
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockService := mocks.NewMockUserService(ctrl)
		c, err := cache.New(mockService, 5*time.Minute)
		require.NoError(t, err)

		ctx := context.Background()

		mockService.EXPECT().
			GetUser(ctx, "1").
			Return(service.User{}, context.Canceled).
			Times(1)

		user, err := c.GetUser(ctx, "1")
		require.Error(t, err)
		require.Equal(t, service.User{}, user)
		require.Contains(t, err.Error(), "failed to get user")
	})
}
