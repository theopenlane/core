package pool

import (
	"context"
	"sync"
)

// simpleManager is an in-memory pool manager with per-key single entries.
type simpleManager[T any] struct {
	mu    sync.Mutex
	items map[Key]T
}

// NewSimpleManager creates an in-memory pool.
func NewSimpleManager[T any]() Manager[T] {
	return &simpleManager[T]{
		items: make(map[Key]T),
	}
}

func (m *simpleManager[T]) Acquire(ctx context.Context, key Key, factory Factory[T]) (Handle[T], error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.items[key]; ok {
		return Handle[T]{Key: key, Client: client}, nil
	}

	client, err := factory.New(ctx, key)
	if err != nil {
		return Handle[T]{}, err
	}

	m.items[key] = client
	return Handle[T]{Key: key, Client: client}, nil
}

func (m *simpleManager[T]) Release(_ context.Context, _ Handle[T]) error {
	// no-op for simple manager (clients are reused)
	return nil
}

func (m *simpleManager[T]) Purge(_ context.Context, key Key) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.items, key)
	return nil
}
