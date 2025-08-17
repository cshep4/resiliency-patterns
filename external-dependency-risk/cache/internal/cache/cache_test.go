package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cshep4/resiliency-patterns/external-dependency-risk/cache/internal/mocks"
)

func TestNew(t *testing.T) {
	t.Run("valid TTL", func(t *testing.T) {
		cache, err := New(mocks.NewUserService(t), 5*time.Minute)
		require.NoError(t, err)
		require.NotNil(t, cache)
	})

	t.Run("invalid service", func(t *testing.T) {
		cache, err := New(nil, 5*time.Minute)
		require.Error(t, err)
		require.Nil(t, cache)
	})

	t.Run("invalid TTL", func(t *testing.T) {
		cache, err := New(mocks.NewUserService(t), 0)
		require.Error(t, err)
		require.Nil(t, cache)
	})
}

func TestGetUser(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		cache, err := New(mocks.NewUserService(t), 5*time.Minute)
		require.NoError(t, err)
		require.NotNil(t, cache)

		user, err := cache.GetUser(context.Background(), "1")
		require.NoError(t, err)
		require.NotNil(t, user)
	})
}
