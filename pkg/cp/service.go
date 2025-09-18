package cp

import (
	"context"

	"github.com/samber/mo"
)

// NewClientService creates a new client service with the specified pool
func NewClientService[T any](pool *ClientPool[T]) *ClientService[T] {
	return &ClientService[T]{
		pool:     pool,
		builders: make(map[ProviderType]ClientBuilder[T]),
	}
}

// RegisterBuilder registers a client builder for a specific client type
func (s *ClientService[T]) RegisterBuilder(clientType ProviderType, builder ClientBuilder[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.builders[clientType] = builder
}

// GetClient retrieves a client from cache or builds a new one
func (s *ClientService[T]) GetClient(ctx context.Context, key ClientCacheKey, clientType ProviderType, credentials map[string]string, config map[string]any) mo.Option[T] {
	// Try cache first
	if cached := s.pool.GetClient(key); cached.IsPresent() {
		return cached
	}

	// Build new client
	s.mu.RLock()
	builderPtr, exists := s.builders[clientType]
	s.mu.RUnlock()

	if !exists {
		return mo.None[T]()
	}

	client, err := builderPtr.
		WithCredentials(credentials).
		WithConfig(config).
		Build(ctx)
	if err != nil {
		return mo.None[T]()
	}

	// Cache the new client
	s.pool.SetClient(key, client)

	return mo.Some(client)
}

// Pool returns the underlying client pool
func (s *ClientService[T]) Pool() *ClientPool[T] {
	return s.pool
}
