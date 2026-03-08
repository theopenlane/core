package keystore

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ClientPoolManager manages client pools constructed from provider-published descriptors
type ClientPoolManager struct {
	// source provides credential retrieval and refresh capabilities
	source CredentialSource
	// mu protects concurrent access to pools and descriptors maps
	mu sync.RWMutex
	// pools indexes client pools by provider and client name
	pools map[clientDescriptorKey]*ClientPool[types.ClientInstance]
	// descriptors stores registered client descriptors
	descriptors map[clientDescriptorKey]types.ClientDescriptor
}

// NewClientPoolManager builds a manager from the supplied credential source and descriptors
func NewClientPoolManager(source CredentialSource, descriptors []types.ClientDescriptor) (*ClientPoolManager, error) {
	if source == nil {
		return nil, ErrBrokerRequired
	}

	manager := &ClientPoolManager{
		source:      source,
		pools:       map[clientDescriptorKey]*ClientPool[types.ClientInstance]{},
		descriptors: map[clientDescriptorKey]types.ClientDescriptor{},
	}

	for _, descriptor := range descriptors {
		if err := manager.RegisterDescriptor(descriptor); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

// RegisterDescriptor registers a single client descriptor and lazily constructs its pool
func (m *ClientPoolManager) RegisterDescriptor(descriptor types.ClientDescriptor) error {
	key, err := descriptorKey(descriptor)
	if err != nil {
		return err
	}

	builder := descriptorClientBuilder{descriptor: descriptor}

	pool, err := NewClientPool(m.source, builder)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.pools[key] = pool
	m.descriptors[key] = descriptor

	return nil
}

// Get retrieves a client for the given provider/client name pair
func (m *ClientPoolManager) Get(ctx context.Context, orgID string, provider types.ProviderType, client types.ClientName, opts ...ClientRequestOption) (types.ClientInstance, error) {
	m.mu.RLock()
	pool := m.pools[clientDescriptorKey{Provider: provider, Name: client}]
	m.mu.RUnlock()

	if pool == nil {
		return types.EmptyClientInstance(), ErrClientNotRegistered
	}

	clientResult, err := pool.Get(ctx, orgID, opts...)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return clientResult, nil
}

// Descriptors returns a copy of all registered descriptors keyed by provider
func (m *ClientPoolManager) Descriptors() map[types.ProviderType][]types.ClientDescriptor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return groupDescriptors(m.descriptors, func(k clientDescriptorKey) types.ProviderType { return k.Provider })
}

// descriptorKey extracts and validates the unique key from a client descriptor
func descriptorKey(descriptor types.ClientDescriptor) (clientDescriptorKey, error) {
	if descriptor.Provider == types.ProviderUnknown {
		return clientDescriptorKey{}, ErrProviderRequired
	}
	if descriptor.Name == "" {
		return clientDescriptorKey{}, ErrClientDescriptorInvalid
	}
	if descriptor.Build == nil {
		return clientDescriptorKey{}, ErrClientBuilderRequired
	}

	return clientDescriptorKey{
		Provider: descriptor.Provider,
		Name:     descriptor.Name,
	}, nil
}

// descriptorClientBuilder adapts a ClientDescriptor to the ClientBuilder interface
type descriptorClientBuilder struct {
	// descriptor contains the provider-published client configuration
	descriptor types.ClientDescriptor
}

// Build constructs a client using the descriptor's build function
func (b descriptorClientBuilder) Build(ctx context.Context, payload models.CredentialSet, config json.RawMessage) (types.ClientInstance, error) {
	return b.descriptor.Build(ctx, payload, append(json.RawMessage(nil), config...))
}

// ProviderType returns the provider identifier for this client builder
func (b descriptorClientBuilder) ProviderType() types.ProviderType {
	return b.descriptor.Provider
}

// clientDescriptorKey uniquely identifies a client descriptor by provider and name
type clientDescriptorKey struct {
	// Provider identifies which provider publishes the client
	Provider types.ProviderType
	// Name identifies the specific client type within the provider
	Name types.ClientName
}

// BuildFromPayload constructs a client directly from the provided payload without using the credential store or pool
func (m *ClientPoolManager) BuildFromPayload(ctx context.Context, provider types.ProviderType, client types.ClientName, payload models.CredentialSet, config json.RawMessage) (types.ClientInstance, error) {
	m.mu.RLock()
	descriptor, ok := m.descriptors[clientDescriptorKey{Provider: provider, Name: client}]
	m.mu.RUnlock()

	if !ok {
		return types.EmptyClientInstance(), ErrClientNotRegistered
	}

	return descriptor.Build(ctx, payload, append(json.RawMessage(nil), config...))
}

// FlattenDescriptors converts a map of provider descriptors into a single slice for manager construction
func FlattenDescriptors(entries map[types.ProviderType][]types.ClientDescriptor) []types.ClientDescriptor {
	return flattenDescriptors(entries)
}
