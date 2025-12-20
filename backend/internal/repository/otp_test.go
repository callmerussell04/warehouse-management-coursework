package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/repository"
)

const testRedisAddr = "localhost:6378"

func newTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: testRedisAddr,
	})

	err := client.Ping(context.Background()).Err()
	require.NoError(t, err)

	err = client.FlushDB(context.Background()).Err()
	require.NoError(t, err)

	return client
}

func TestRedisOTPRepository_Save(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()
	repo := repository.NewRedisOTPRepository(client, newDiscardLogger())

	err := repo.Save(context.Background(), "test@mail.com", "123456", time.Minute)
	require.NoError(t, err)

	valid, err := repo.Verify(context.Background(), "test@mail.com", "123456", 5)
	assert.NoError(t, err)
	assert.True(t, valid)

	valid, err = repo.Verify(context.Background(), "test@mail.com", "123456", 5)
	assert.ErrorIs(t, err, customErrors.ErrNotFound)
	assert.False(t, valid)
}

func TestRedisOTPRepository_AttemptLimit(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()
	repo := repository.NewRedisOTPRepository(client, newDiscardLogger())

	err := repo.Save(context.Background(), "attempts@mail.com", "654321", time.Minute)
	require.NoError(t, err)

	for attempt := 1; attempt <= 5; attempt++ {
		valid, verifyErr := repo.Verify(context.Background(), "attempts@mail.com", "000000", 5)
		assert.NoError(t, verifyErr)
		assert.False(t, valid)
	}

	valid, err := repo.Verify(context.Background(), "attempts@mail.com", "654321", 5)
	assert.ErrorIs(t, err, customErrors.ErrNotFound)
	assert.False(t, valid)
}
