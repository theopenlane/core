package keystore

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	integrationops "github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// OperationManager executes provider-published operations using stored credentials and optional client pools
type OperationManager struct {
	// source provides credential retrieval and refresh capabilities
	source IntegrationCredentialSource
	// clients provides access to pooled provider clients
	clients *ClientPoolManager
	// mu protects concurrent access to the descriptors map
	mu sync.RWMutex
	// descriptors indexes registered operations by provider and name
	descriptors map[operationKey]types.OperationDescriptor
}

// IntegrationCredentialSource extends CredentialSource with integration-specific lookup methods.
type IntegrationCredentialSource interface {
	CredentialSource
	// GetForIntegration retrieves credentials scoped to a specific integration.
	GetForIntegration(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialSet, error)
	// MintForIntegration refreshes credentials scoped to a specific integration.
	MintForIntegration(ctx context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialSet, error)
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
func NewOperationManager(source IntegrationCredentialSource, descriptors []types.OperationDescriptor, opts ...OperationManagerOption) (*OperationManager, error) {
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
	descriptor, err := m.resolveRequestDescriptor(req)
	if err != nil {
		return types.OperationResult{}, err
	}
	if err := integrationops.ValidateConfig(descriptor.ConfigSchema, req.Config); err != nil {
		return types.OperationResult{}, err
	}

	payload, err := m.resolveCredential(ctx, req)
	if err != nil {
		return types.OperationResult{}, err
	}

	client, err := m.resolveClient(ctx, req, descriptor, payload, req.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	input := types.OperationInput{
		OrgID:      req.OrgID,
		Provider:   req.Provider,
		Credential: payload,
		Client:     client,
		Config:     jsonx.CloneRawMessage(req.Config),
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

// RunWithCredential executes the requested operation using the provided credential set instead of loading from the store.
func (m *OperationManager) RunWithCredential(ctx context.Context, req types.OperationRequest, credential types.CredentialSet) (types.OperationResult, error) {
	descriptor, err := m.resolveRequestDescriptor(req)
	if err != nil {
		return types.OperationResult{}, err
	}
	if err := integrationops.ValidateConfig(descriptor.ConfigSchema, req.Config); err != nil {
		return types.OperationResult{}, err
	}

	client, err := m.resolveClientFromCredential(ctx, req, descriptor, credential, req.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	input := types.OperationInput{
		OrgID:      req.OrgID,
		Provider:   req.Provider,
		Credential: credential,
		Client:     client,
		Config:     jsonx.CloneRawMessage(req.Config),
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

func validateOperationRequest(req types.OperationRequest) error {
	if req.OrgID == "" {
		return ErrOrgIDRequired
	}
	if req.Provider == types.ProviderUnknown {
		return ErrProviderRequired
	}
	if req.Name == "" {
		return ErrOperationNameRequired
	}

	return nil
}

func (m *OperationManager) resolveRequestDescriptor(req types.OperationRequest) (types.OperationDescriptor, error) {
	if err := validateOperationRequest(req); err != nil {
		return types.OperationDescriptor{}, err
	}

	key := operationKey{
		provider: req.Provider,
		name:     req.Name,
	}

	m.mu.RLock()
	descriptor, ok := m.descriptors[key]
	m.mu.RUnlock()

	if !ok {
		return types.OperationDescriptor{}, ErrOperationNotRegistered
	}

	return descriptor, nil
}

// ValidateProviderHealth executes the default health operation with a supplied credential payload.
func (m *OperationManager) ValidateProviderHealth(ctx context.Context, orgID string, provider types.ProviderType, credential types.CredentialSet) (types.OperationResult, error) {
	return m.RunWithCredential(ctx, types.OperationRequest{
		OrgID:    orgID,
		Provider: provider,
		Name:     types.OperationHealthDefault,
	}, credential)
}

// resolveClientFromCredential builds a client from the provided credential when the operation requires one.
func (m *OperationManager) resolveClientFromCredential(ctx context.Context, req types.OperationRequest, descriptor types.OperationDescriptor, credential types.CredentialSet, config json.RawMessage) (types.ClientInstance, error) {
	if descriptor.Client == "" {
		return types.EmptyClientInstance(), nil
	}

	if m.clients == nil {
		return types.EmptyClientInstance(), ErrOperationClientManagerRequired
	}

	return m.clients.BuildFromPayload(ctx, req.Provider, descriptor.Client, credential, jsonx.CloneRawMessage(config))
}

// resolveCredential retrieves or refreshes the credential based on the request flags
func (m *OperationManager) resolveCredential(ctx context.Context, req types.OperationRequest) (types.CredentialSet, error) {
	if req.IntegrationID != "" {
		return resolveCredentialWithPolicy(
			ctx,
			req.Force,
			time.Now,
			func(callCtx context.Context) (types.CredentialSet, error) {
				return m.source.GetForIntegration(callCtx, req.OrgID, req.Provider, req.IntegrationID)
			},
			func(callCtx context.Context, _ types.CredentialSet) (types.CredentialSet, error) {
				return m.source.MintForIntegration(callCtx, req.OrgID, req.Provider, req.IntegrationID)
			},
		)
	}

	return resolveCredentialWithPolicy(
		ctx,
		req.Force,
		time.Now,
		func(callCtx context.Context) (types.CredentialSet, error) {
			return m.source.Get(callCtx, req.OrgID, req.Provider)
		},
		func(callCtx context.Context, _ types.CredentialSet) (types.CredentialSet, error) {
			return m.source.Mint(callCtx, req.OrgID, req.Provider)
		},
	)
}

// resolveClient retrieves a client instance if the operation requires one
func (m *OperationManager) resolveClient(ctx context.Context, req types.OperationRequest, descriptor types.OperationDescriptor, credential types.CredentialSet, config json.RawMessage) (types.ClientInstance, error) {
	if descriptor.Client == "" {
		return types.EmptyClientInstance(), nil
	}

	if req.IntegrationID != "" {
		return m.resolveClientFromCredential(ctx, req, descriptor, credential, config)
	}

	if m.clients == nil {
		return types.EmptyClientInstance(), ErrOperationClientManagerRequired
	}

	opts := []ClientRequestOption{}
	if len(config) > 0 {
		opts = append(opts, WithClientConfig(jsonx.CloneRawMessage(config)))
	}
	if req.ClientForce {
		opts = append(opts, WithClientForceRefresh())
	}

	return m.clients.Get(ctx, req.OrgID, req.Provider, descriptor.Client, opts...)
}

// Descriptors returns a copy of all registered operations keyed by provider
func (m *OperationManager) Descriptors() map[types.ProviderType][]types.OperationDescriptor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return groupDescriptors(m.descriptors, func(k operationKey) types.ProviderType { return k.provider })
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
	if descriptor.Name == "" {
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
	return flattenDescriptors(entries)
}
