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

	type args struct {
		email    string
		code     string
		duration time.Duration
	}

	tests := []struct {
		name      string
		args      args
		wantError error
		checkRes  func(*testing.T)
	}{
		{
			name: "Success",
			args: args{
				email:    "test@mail.com",
				code:     "123456",
				duration: time.Minute,
			},
			wantError: nil,
			checkRes: func(t *testing.T) {
				val, err := client.Get(context.Background(), "otp:test@mail.com").Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Save(context.Background(), tc.args.email, tc.args.code, tc.args.duration)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
				if tc.checkRes != nil {
					tc.checkRes(t)
				}
			}
		})
	}
}

func TestRedisOTPRepository_Get(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()
	repo := repository.NewRedisOTPRepository(client, newDiscardLogger())

	email := "get@mail.com"
	err := client.Set(context.Background(), "otp:get@mail.com", "654321", time.Minute).Err()
	require.NoError(t, err)

	tests := []struct {
		name      string
		email     string
		wantCode  string
		wantError error
	}{
		{
			name:      "Success",
			email:     email,
			wantCode:  "654321",
			wantError: nil,
		},
		{
			name:      "Not Found",
			email:     "missing@mail.com",
			wantCode:  "",
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, err := repo.Get(context.Background(), tc.email)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
				assert.Empty(t, val)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantCode, val)
			}
		})
	}
}

func TestRedisOTPRepository_Delete(t *testing.T) {
	client := newTestRedis(t)
	defer client.Close()
	repo := repository.NewRedisOTPRepository(client, newDiscardLogger())

	email := "del@mail.com"
	err := client.Set(context.Background(), "otp:del@mail.com", "111111", time.Minute).Err()
	require.NoError(t, err)

	tests := []struct {
		name      string
		email     string
		wantError error
		checkRes  func(*testing.T)
	}{
		{
			name:      "Success",
			email:     email,
			wantError: nil,
			checkRes: func(t *testing.T) {
				_, err := client.Get(context.Background(), "otp:del@mail.com").Result()
				assert.Equal(t, redis.Nil, err)
			},
		},
		{
			name:      "Delete Non-Existent",
			email:     "ghost@mail.com",
			wantError: nil,
			checkRes:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Delete(context.Background(), tc.email)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
				if tc.checkRes != nil {
					tc.checkRes(t)
				}
			}
		})
	}
}
