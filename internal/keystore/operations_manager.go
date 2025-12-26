package keystore

import (
	"context"
	"maps"
	"strings"
	"sync"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/types"
)

// OperationManager executes provider-published operations using stored credentials and optional client pools
type OperationManager struct {
	// source provides credential retrieval and refresh capabilities
	source CredentialSource
	// clients provides access to pooled provider clients
	clients *ClientPoolManager
	// mu protects concurrent access to the descriptors map
	mu sync.RWMutex
	// descriptors indexes registered operations by provider and name
	descriptors map[operationKey]types.OperationDescriptor
}

// OperationManagerOption customizes manager construction
type OperationManagerOption func(*OperationManager)

// WithOperationClients registers the client pool manager used to satisfy operation client dependencies
func WithOperationClients(clients *ClientPoolManager) OperationManagerOption {
	return func(manager *OperationManager) {
		manager.clients = clients
	}
}

// NewOperationManager builds an OperationManager from the supplied credential source and descriptors
func NewOperationManager(source CredentialSource, descriptors []types.OperationDescriptor, opts ...OperationManagerOption) (*OperationManager, error) {
	if source == nil {
		return nil, ErrBrokerRequired
	}

	manager := &OperationManager{
		source:      source,
		descriptors: map[operationKey]types.OperationDescriptor{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(manager)
		}
	}

	for _, descriptor := range descriptors {
		if err := manager.RegisterDescriptor(descriptor); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

// RegisterDescriptor registers an operation descriptor and makes it available to callers
func (m *OperationManager) RegisterDescriptor(descriptor types.OperationDescriptor) error {
	key, err := operationDescriptorKey(descriptor)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.descriptors[key] = descriptor

	return nil
}

// Run executes the requested provider operation using stored credentials and optional clients
func (m *OperationManager) Run(ctx context.Context, req types.OperationRequest) (types.OperationResult, error) {
	if strings.TrimSpace(req.OrgID) == "" {
		return types.OperationResult{}, ErrOrgIDRequired
	}
	if req.Provider == types.ProviderUnknown {
		return types.OperationResult{}, ErrProviderRequired
	}
	if strings.TrimSpace(string(req.Name)) == "" {
		return types.OperationResult{}, ErrOperationNameRequired
	}

	key := operationKey{
		provider: req.Provider,
		name:     req.Name,
	}

	m.mu.RLock()
	descriptor, ok := m.descriptors[key]
	m.mu.RUnlock()

	if !ok {
		return types.OperationResult{}, ErrOperationNotRegistered
	}

	payload, err := m.resolveCredential(ctx, req)
	if err != nil {
		return types.OperationResult{}, err
	}

	client, err := m.resolveClient(ctx, req, descriptor)
	if err != nil {
		return types.OperationResult{}, err
	}

	input := types.OperationInput{
		OrgID:      req.OrgID,
		Provider:   req.Provider,
		Credential: payload,
		Client:     client,
		Config:     cloneConfigMap(req.Config),
	}

	result, runErr := descriptor.Run(ctx, input)
	if result.Status == "" {
		result.Status = types.OperationStatusUnknown
	}

	if runErr != nil {
		return result, runErr
	}

	return result, nil
}

// resolveCredential retrieves or refreshes the credential based on the request flags
func (m *OperationManager) resolveCredential(ctx context.Context, req types.OperationRequest) (types.CredentialPayload, error) {
	if req.Force {
		return m.source.Mint(ctx, req.OrgID, req.Provider)
	}

	return m.source.Get(ctx, req.OrgID, req.Provider)
}

// resolveClient retrieves a client instance if the operation requires one
func (m *OperationManager) resolveClient(ctx context.Context, req types.OperationRequest, descriptor types.OperationDescriptor) (any, error) {
	if descriptor.Client == "" {
		return nil, nil
	}

	if m.clients == nil {
		return nil, ErrOperationClientManagerRequired
	}

	opts := []ClientRequestOption[map[string]any]{}
	if len(req.Config) > 0 {
		opts = append(opts, WithClientConfig(maps.Clone(req.Config)))
	}
	if req.ClientForce {
		opts = append(opts, WithClientForceRefresh[map[string]any]())
	}

	return m.clients.Get(ctx, req.OrgID, req.Provider, descriptor.Client, opts...)
}

// Descriptors returns a copy of all registered operations keyed by provider
func (m *OperationManager) Descriptors() map[types.ProviderType][]types.OperationDescriptor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	grouped := map[types.ProviderType][]types.OperationDescriptor{}
	for key, descriptor := range m.descriptors {
		grouped[key.provider] = append(grouped[key.provider], descriptor)
	}

	for provider, descriptors := range grouped {
		copied := make([]types.OperationDescriptor, len(descriptors))
		copy(copied, descriptors)
		grouped[provider] = copied
	}

	return grouped
}

// operationKey uniquely identifies an operation by provider and name
type operationKey struct {
	// provider identifies which provider publishes the operation
	provider types.ProviderType
	// name identifies the specific operation within the provider
	name types.OperationName
}

// operationDescriptorKey extracts and validates the unique key from an operation descriptor
func operationDescriptorKey(descriptor types.OperationDescriptor) (operationKey, error) {
	if descriptor.Provider == types.ProviderUnknown {
		return operationKey{}, ErrProviderRequired
	}
	if strings.TrimSpace(string(descriptor.Name)) == "" {
		return operationKey{}, ErrOperationDescriptorInvalid
	}
	if descriptor.Run == nil {
		return operationKey{}, ErrOperationDescriptorInvalid
	}

	return operationKey{
		provider: descriptor.Provider,
		name:     descriptor.Name,
	}, nil
}

// FlattenOperationDescriptors converts a map of provider operations into a single slice for manager construction
func FlattenOperationDescriptors(entries map[types.ProviderType][]types.OperationDescriptor) []types.OperationDescriptor {
	if len(entries) == 0 {
		return nil
	}

	return lo.Flatten(lo.Values(entries))
}
