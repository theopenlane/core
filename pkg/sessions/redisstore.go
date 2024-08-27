package sessions

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// PersistentStore is defining an interface for session store
type PersistentStore interface {
	Exists(ctx context.Context, key string) (int64, error)
	GetSession(ctx context.Context, key string) (string, error)
	StoreSession(ctx context.Context, key, value string) error
	StoreSessionWithExpiration(ctx context.Context, key, value string, ttl time.Duration) error
	DeleteSession(ctx context.Context, key string) error
}

var _ PersistentStore = &persistentStore{}

// persistentStore stores Sessions in a persistent data store (redis)
type persistentStore struct {
	client *redis.Client
	mu     sync.Mutex
}

// NewStore returns a new Store that stores to a persistent backend (redis)
func NewStore(client *redis.Client) PersistentStore {
	return &persistentStore{
		client: client,
	}
}

// Exists checks to see if there is an existing session for the user
func (s *persistentStore) Exists(ctx context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.client.Exists(ctx, key).Result()
}

// GetSession checks to see if there is an existing session for the user
func (s *persistentStore) GetSession(ctx context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.client.Get(ctx, key).Result()
}

// StoreSession is used to store a session in the store with a key and value
// the TTL is set to the defaultMaxAgeSeconds to align with the session cookie
func (s *persistentStore) StoreSession(ctx context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.client.Set(ctx, key, value, defaultMaxAgeSeconds).Err()
}

// StoreSessionWithExpiration is used to store a session in the store with a key and value and a time to live
func (s *persistentStore) StoreSessionWithExpiration(ctx context.Context, key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.client.Set(ctx, key, value, ttl).Err()
}

// DeleteSession is used to delete a session from the store
func (s *persistentStore) DeleteSession(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.client.Del(ctx, userID).Err()
}
