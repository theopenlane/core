// Package pool manages reusable client pools for integration providers.
package pool

import "context"

// Key uniquely identifies a client pool entry (e.g., org + provider + scope).
type Key struct {
	OrgID     string
	Provider  string
	ScopeHash string
}

// Factory constructs new client instances when the pool misses.
type Factory[T any] interface {
	New(ctx context.Context, key Key) (T, error)
}

// Manager tracks pooled clients and coordinates acquisition and release.
type Manager[T any] interface {
	Acquire(ctx context.Context, key Key, factory Factory[T]) (Handle[T], error)
	Release(ctx context.Context, handle Handle[T]) error
	Purge(ctx context.Context, key Key) error
}

// Handle represents a pooled client checkout.
type Handle[T any] struct {
	Key    Key
	Client T
}
