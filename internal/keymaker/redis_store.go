package keymaker

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisAuthStateKeyPrefix = "keymaker:auth-state:"

// RedisAuthStateStore stores auth callback state in Redis
type RedisAuthStateStore struct {
	client *redis.Client
	now    func() time.Time
}

// NewRedisAuthStateStore returns a Redis-backed auth state store
func NewRedisAuthStateStore(client *redis.Client) *RedisAuthStateStore {
	return &RedisAuthStateStore{
		client: client,
		now:    time.Now,
	}
}

// Save records the provided definition authorization state with an expiry
func (r *RedisAuthStateStore) Save(state AuthState) error {
	if state.State == "" {
		return ErrAuthStateTokenRequired
	}

	clone := state
	now := r.now()
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = now
	}

	if clone.ExpiresAt.IsZero() {
		clone.ExpiresAt = clone.CreatedAt.Add(defaultSessionTTL)
	}

	ttl := time.Until(clone.ExpiresAt)
	if ttl <= 0 {
		return ErrAuthStateExpired
	}

	encoded, err := json.Marshal(clone)
	if err != nil {
		return err
	}

	return r.client.Set(context.Background(), redisAuthStateKey(clone.State), encoded, ttl).Err()
}

// Take retrieves and deletes authorization state associated with the given token
func (r *RedisAuthStateStore) Take(token string) (AuthState, error) {
	if token == "" {
		return AuthState{}, ErrAuthStateTokenRequired
	}

	encoded, err := r.client.GetDel(context.Background(), redisAuthStateKey(token)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return AuthState{}, ErrAuthStateNotFound
		}

		return AuthState{}, err
	}

	var state AuthState
	if err := json.Unmarshal(encoded, &state); err != nil {
		return AuthState{}, err
	}

	now := r.now()
	if !state.ExpiresAt.IsZero() && !state.ExpiresAt.After(now) {
		return AuthState{}, ErrAuthStateExpired
	}

	return state, nil
}

// redisAuthStateKey returns the Redis key for the given auth state token
func redisAuthStateKey(token string) string {
	return redisAuthStateKeyPrefix + token
}
