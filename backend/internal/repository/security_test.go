package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/repository"
)

func TestRedisRefreshSessionRepository_RotateAndRevoke(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()
	repo := repository.NewRedisRefreshSessionRepository(client)
	ctx := context.Background()
	userID := uuid.New()

	require.NoError(t, repo.Save(ctx, "old-refresh-token", userID, time.Hour))
	keys, err := client.Keys(ctx, "*old-refresh-token*").Result()
	require.NoError(t, err)
	assert.Empty(t, keys, "raw refresh token must not be present in Redis keys")

	rotatedUserID, err := repo.Rotate(ctx, "old-refresh-token", "new-refresh-token", time.Hour)
	require.NoError(t, err)
	assert.Equal(t, userID, rotatedUserID)

	_, err = repo.Rotate(ctx, "old-refresh-token", "replayed-token", time.Hour)
	assert.ErrorIs(t, err, customErrors.ErrUnauthorized)

	require.NoError(t, repo.RevokeAll(ctx, userID))
	_, err = repo.Rotate(ctx, "new-refresh-token", "after-revoke", time.Hour)
	assert.ErrorIs(t, err, customErrors.ErrUnauthorized)
}

func TestRedisRateLimiter(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()
	repo := repository.NewRedisRateLimiter(client)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		allowed, _, err := repo.Allow(ctx, "login:ip:127.0.0.1", 3, time.Minute)
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	allowed, retryAfter, err := repo.Allow(ctx, "login:ip:127.0.0.1", 3, time.Minute)
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Positive(t, retryAfter)
}
