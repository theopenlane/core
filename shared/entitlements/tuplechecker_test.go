package entitlements

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type mockFGA struct {
	checkResult bool
	checkErr    error
	createErr   error
	deleteErr   error
}

func (m *mockFGA) CheckTuple(ctx context.Context, tuple FeatureTuple) (bool, error) {
	return m.checkResult, m.checkErr
}
func (m *mockFGA) CreateTuple(ctx context.Context, tuple FeatureTuple) error {
	return m.createErr
}
func (m *mockFGA) DeleteTuple(ctx context.Context, tuple FeatureTuple) error {
	return m.deleteErr
}

// newTestRedis returns a redis.Client backed by a fresh miniredis instance for each test.
func newTestRedis(t *testing.T) *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	return redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
}

func TestCheckFeatureTuple_CacheHit(t *testing.T) {
	ctx := context.Background()
	redisClient := newTestRedis(t)

	tuple := FeatureTuple{UserID: "u1", Feature: "f1", Context: map[string]any{"k": "v"}}
	key := "feature:f1:user:u1:ctx:{\"k\":\"v\"}"
	err := redisClient.Set(ctx, key, "1", 5*time.Minute).Err()
	assert.NoError(t, err)

	tc := NewTupleChecker(WithRedisClient(redisClient), WithFGAClient(&mockFGA{}))
	ok, err := tc.CheckFeatureTuple(ctx, tuple)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestCheckFeatureTuple_FallbackToFGA(t *testing.T) {
	ctx := context.Background()
	redisClient := newTestRedis(t)

	tuple := FeatureTuple{UserID: "u2", Feature: "f2", Context: map[string]any{"k": "v2"}}
	mock := &mockFGA{checkResult: true}
	tc := NewTupleChecker(WithRedisClient(redisClient), WithFGAClient(mock))
	ok, err := tc.CheckFeatureTuple(ctx, tuple)
	assert.NoError(t, err)
	assert.True(t, ok)
	// Should be cached now
	ok2, err := tc.CheckFeatureTuple(ctx, tuple)
	assert.NoError(t, err)
	assert.True(t, ok2)
}

func TestCreateFeatureTuple(t *testing.T) {
	ctx := context.Background()
	redisClient := newTestRedis(t)

	tuple := FeatureTuple{UserID: "u3", Feature: "f3", Context: map[string]any{"k": "v3"}}
	mock := &mockFGA{}
	tc := NewTupleChecker(WithRedisClient(redisClient), WithFGAClient(mock))
	err := tc.CreateFeatureTuple(ctx, tuple)
	assert.NoError(t, err)
	// Should be cached
	ok, err := tc.CheckFeatureTuple(ctx, tuple)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestTupleChecker_NotConfigured(t *testing.T) {
	tc := NewTupleChecker()
	_, err := tc.CheckFeatureTuple(context.Background(), FeatureTuple{})
	assert.Error(t, err)
	err = tc.CreateFeatureTuple(context.Background(), FeatureTuple{})
	assert.Error(t, err)
}

func TestCheckFeatureTuple_FGAError(t *testing.T) {
	ctx := context.Background()
	redisClient := newTestRedis(t)

	tuple := FeatureTuple{UserID: "u4", Feature: "f4", Context: map[string]any{"k": "v4"}}
	mock := &mockFGA{checkErr: errors.New("fga error")}
	tc := NewTupleChecker(WithRedisClient(redisClient), WithFGAClient(mock))
	_, err := tc.CheckFeatureTuple(ctx, tuple)
	assert.Error(t, err)
}

func TestDeleteFeatureTuple(t *testing.T) {
	ctx := context.Background()
	redisClient := newTestRedis(t)

	tuple := FeatureTuple{UserID: "u5", Feature: "f5", Context: map[string]any{"k": "v5"}}
	mock := &mockFGA{}
	tc := NewTupleChecker(WithRedisClient(redisClient), WithFGAClient(mock))

	// Pre-populate cache
	key := tc.cacheKey(tuple)
	err := redisClient.Set(ctx, key, "1", 5*time.Minute).Err()
	assert.NoError(t, err)

	err = tc.DeleteFeatureTuple(ctx, tuple)
	assert.NoError(t, err)
	// Should be gone from cache
	_, err = redisClient.Get(ctx, key).Result()
	assert.ErrorIs(t, err, redis.Nil)
}

func TestDeleteFeatureTuple_FGAError(t *testing.T) {
	ctx := context.Background()
	redisClient := newTestRedis(t)
	tuple := FeatureTuple{UserID: "u6", Feature: "f6", Context: map[string]any{"k": "v6"}}
	mock := &mockFGA{deleteErr: errors.New("fga delete error")}
	tc := NewTupleChecker(WithRedisClient(redisClient), WithFGAClient(mock))
	err := tc.DeleteFeatureTuple(ctx, tuple)
	assert.Error(t, err)
}

func TestDeleteFeatureTuple_CacheError(t *testing.T) {
	ctx := context.Background()
	redisClient := newTestRedis(t)
	tuple := FeatureTuple{UserID: "u7", Feature: "f7", Context: map[string]any{"k": "v7"}}
	mock := &mockFGA{}
	tc := NewTupleChecker(WithRedisClient(redisClient), WithFGAClient(mock))

	// Close Redis to force error
	redisClient.Close()
	err := tc.DeleteFeatureTuple(ctx, tuple)
	assert.Error(t, err)
}
