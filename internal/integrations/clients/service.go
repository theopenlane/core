package clients

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Service resolves and caches buildable clients for installations
type Service struct {
	registry    registry.DefinitionRegistry
	credentials types.CredentialResolver
	mu          sync.RWMutex
	cache       map[string]any
}

// NewService constructs the client service
func NewService(reg registry.DefinitionRegistry, credentialResolver types.CredentialResolver) (*Service, error) {
	if reg == nil {
		return nil, ErrRegistryRequired
	}

	if credentialResolver == nil {
		return nil, ErrCredentialResolverRequired
	}

	return &Service{
		registry:    reg,
		credentials: credentialResolver,
		cache:       map[string]any{},
	}, nil
}

// Get resolves one named client for an installation and caches it by installation, client, credential, and config
func (s *Service) Get(ctx context.Context, installation *ent.Integration, name types.ClientName, config json.RawMessage, force bool) (any, error) {
	if installation == nil {
		return nil, ErrInstallationRequired
	}

	if installation.DefinitionID == "" {
		return nil, ErrDefinitionIDRequired
	}

	credential, found, err := s.credentials.LoadCredential(ctx, installation)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrCredentialNotFound
	}

	return s.Build(ctx, installation, name, credential, config, force)
}

// Build resolves one named client for an installation using an explicit credential payload
func (s *Service) Build(ctx context.Context, installation *ent.Integration, name types.ClientName, credential types.CredentialSet, config json.RawMessage, force bool) (any, error) {
	if installation == nil {
		return nil, ErrInstallationRequired
	}

	if installation.DefinitionID == "" {
		return nil, ErrDefinitionIDRequired
	}

	registration, err := s.registry.Client(types.DefinitionID(installation.DefinitionID), name)
	if err != nil {
		return nil, err
	}

	cacheKey := buildCacheKey(installation.ID, name, credential, config)
	if !force {
		if cached, ok := s.cached(cacheKey); ok {
			return cached, nil
		}
	}

	client, err := registration.Build(ctx, types.ClientBuildRequest{
		Installation: installation,
		Credential:   credential,
		Config:       jsonx.CloneRawMessage(config),
	})
	if err != nil {
		return nil, err
	}

	s.store(cacheKey, client)
	return client, nil
}

// Invalidate drops all cached clients for one installation
func (s *Service) Invalidate(installationID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix := installationID + ":"
	for key := range s.cache {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(s.cache, key)
		}
	}
}

// cached resolves one cached client instance
func (s *Service) cached(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.cache[key]
	return client, ok
}

// store persists one cached client instance
func (s *Service) store(key string, client any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[key] = client
}

// buildCacheKey hashes the variable client inputs into one stable cache key
func buildCacheKey(installationID string, name types.ClientName, credential types.CredentialSet, config json.RawMessage) string {
	digest := sha256.New()
	_, _ = digest.Write([]byte(installationID))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write([]byte(name))
	_, _ = digest.Write([]byte{0})

	if encodedCredential, err := json.Marshal(credential); err == nil {
		_, _ = digest.Write(encodedCredential)
	}

	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write(config)

	return installationID + ":" + string(name) + ":" + hex.EncodeToString(digest.Sum(nil))
}
