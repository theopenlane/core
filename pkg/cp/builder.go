package cp

import (
	"context"
	"sync"
	"time"
)

// ProviderType represents a provider type identifier
type ProviderType string

// ClientBuilder builds client instances with credentials and configuration
type ClientBuilder[T any] interface {
	WithCredentials(credentials map[string]string) ClientBuilder[T]
	WithConfig(config map[string]any) ClientBuilder[T]
	Build(ctx context.Context) (T, error)
	ClientType() ProviderType
}

// ClientCacheKey uniquely identifies a cached client
type ClientCacheKey struct {
	TenantID        string
	IntegrationType string
	HushID          string
	IntegrationID   string
}

// ClientEntry wraps a client instance with expiration metadata
type ClientEntry[T any] struct {
	Client     T
	Expiration time.Time
}

// ClientPool holds cached client instances with TTL expiration
type ClientPool[T any] struct {
	mu      sync.RWMutex
	clients map[ClientCacheKey]*ClientEntry[T]
	ttl     time.Duration
}

// ClientService manages client builders and provides cached client instances
type ClientService[T any] struct {
	pool     *ClientPool[T]
	builders map[ProviderType]ClientBuilder[T]
	mu       sync.RWMutex
}
