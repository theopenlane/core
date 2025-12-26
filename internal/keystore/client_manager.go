package keystore

import (
	"context"
	"strings"
	"sync"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

// ClientPoolManager manages client pools constructed from provider-published descriptors
type ClientPoolManager struct {
	// source provides credential retrieval and refresh capabilities
	source CredentialSource
	// mu protects concurrent access to pools and descriptors maps
	mu sync.RWMutex
	// pools indexes client pools by provider and client name
	pools map[clientDescriptorKey]*ClientPool[any, map[string]any]
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
		pools:       map[clientDescriptorKey]*ClientPool[any, map[string]any]{},
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

	pool, err := NewClientPool(m.source, builder,
		WithClientConfigClone[any](cloneConfigMap))
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
func (m *ClientPoolManager) Get(ctx context.Context, orgID string, provider types.ProviderType, client types.ClientName, opts ...ClientRequestOption[map[string]any]) (any, error) {
	m.mu.RLock()
	pool := m.pools[clientDescriptorKey{Provider: provider, Name: client}]
	m.mu.RUnlock()

	if pool == nil {
		return nil, ErrClientNotRegistered
	}

	clientResult, err := pool.Get(ctx, orgID, opts...)
	if err != nil {
		return nil, err
	}

	return clientResult, nil
}

// Descriptors returns a copy of all registered descriptors keyed by provider
func (m *ClientPoolManager) Descriptors() map[types.ProviderType][]types.ClientDescriptor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	grouped := map[types.ProviderType][]types.ClientDescriptor{}
	for key, descriptor := range m.descriptors {
		grouped[key.Provider] = append(grouped[key.Provider], descriptor)
	}

	for provider, descriptors := range grouped {
		copied := make([]types.ClientDescriptor, len(descriptors))
		copy(copied, descriptors)
		grouped[provider] = copied
	}

	return grouped
}

// descriptorKey extracts and validates the unique key from a client descriptor
func descriptorKey(descriptor types.ClientDescriptor) (clientDescriptorKey, error) {
	if descriptor.Provider == types.ProviderUnknown {
		return clientDescriptorKey{}, ErrProviderRequired
	}
	if strings.TrimSpace(string(descriptor.Name)) == "" {
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
func (b descriptorClientBuilder) Build(ctx context.Context, payload types.CredentialPayload, config map[string]any) (any, error) {
	cloned := cloneConfigMap(config)
	return b.descriptor.Build(ctx, payload, cloned)
}

// ProviderType returns the provider identifier for this client builder
func (b descriptorClientBuilder) ProviderType() types.ProviderType {
	return b.descriptor.Provider
}

// cloneConfigMap creates a deep copy of the configuration map
func cloneConfigMap(input map[string]any) map[string]any {
	return helpers.DeepCloneMap(input)
}

// clientDescriptorKey uniquely identifies a client descriptor by provider and name
type clientDescriptorKey struct {
	// Provider identifies which provider publishes the client
	Provider types.ProviderType
	// Name identifies the specific client type within the provider
	Name types.ClientName
}

// FlattenDescriptors converts a map of provider descriptors into a single slice for manager construction
func FlattenDescriptors(entries map[types.ProviderType][]types.ClientDescriptor) []types.ClientDescriptor {
	if len(entries) == 0 {
		return nil
	}

	return lo.Flatten(lo.Values(entries))
}
