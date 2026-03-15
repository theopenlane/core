package keymaker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisAuthStatePrefix = "integrationsv2:keymaker:authstate:"

// RedisAuthStateStore stores auth state in Redis so OAuth callbacks work across instances.
type RedisAuthStateStore struct {
	client *redis.Client
	prefix string
	now    func() time.Time
}

// NewRedisAuthStateStore constructs a Redis-backed auth state store.
func NewRedisAuthStateStore(client *redis.Client) *RedisAuthStateStore {
	return &RedisAuthStateStore{
		client: client,
		prefix: redisAuthStatePrefix,
		now:    time.Now,
	}
}

// Save persists one auth state payload with an expiry aligned to the auth session lifetime.
func (s *RedisAuthStateStore) Save(state AuthState) error {
	if state.State == "" {
		return ErrAuthStateTokenRequired
	}

	clone := state
	now := s.now()

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

	payload, err := json.Marshal(clone)
	if err != nil {
		return fmt.Errorf("keymaker: marshal auth state: %w", err)
	}

	return s.client.Set(context.Background(), s.key(clone.State), payload, ttl).Err()
}

// Take loads and deletes one auth state payload.
func (s *RedisAuthStateStore) Take(token string) (AuthState, error) {
	if token == "" {
		return AuthState{}, ErrAuthStateTokenRequired
	}

	data, err := s.client.GetDel(context.Background(), s.key(token)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return AuthState{}, ErrAuthStateNotFound
		}

		return AuthState{}, err
	}

	var state AuthState
	if err := json.Unmarshal(data, &state); err != nil {
		return AuthState{}, fmt.Errorf("keymaker: unmarshal auth state: %w", err)
	}

	if !state.ExpiresAt.IsZero() && !state.ExpiresAt.After(s.now()) {
		return AuthState{}, ErrAuthStateExpired
	}

	return state, nil
}

func (s *RedisAuthStateStore) key(token string) string {
	return s.prefix + token
}
