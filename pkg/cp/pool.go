package cp

import (
	"time"

	"github.com/samber/mo"
)

// NewClientPool creates a new client pool with the specified TTL
func NewClientPool[T any](ttl time.Duration) *ClientPool[T] {
	return &ClientPool[T]{
		clients: make(map[ClientCacheKey]*ClientEntry[T]),
		ttl:     ttl,
	}
}

// GetClient retrieves a client from the pool if it exists and hasn't expired
func (p *ClientPool[T]) GetClient(key ClientCacheKey) mo.Option[T] {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if entry, exists := p.clients[key]; exists && time.Now().Before(entry.Expiration) {
		return mo.Some(entry.Client)
	}

	return mo.None[T]()
}

// SetClient stores a client in the pool with TTL expiration
func (p *ClientPool[T]) SetClient(key ClientCacheKey, client T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.clients[key] = &ClientEntry[T]{
		Client:     client,
		Expiration: time.Now().Add(p.ttl),
	}
}

// RemoveClient removes a client from the pool
func (p *ClientPool[T]) RemoveClient(key ClientCacheKey) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.clients, key)
}


// CleanExpired removes expired clients from the pool and returns the count of removed clients
func (p *ClientPool[T]) CleanExpired() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	removed := 0
	for key, entry := range p.clients {
		if now.After(entry.Expiration) {
			delete(p.clients, key)
			removed++
		}
	}
	return removed
}
