package providers

import "github.com/theopenlane/core/common/integrations/types"

// BaseProvider stores shared provider metadata
type BaseProvider struct {
	Provider types.ProviderType
	Caps     types.ProviderCapabilities
	Ops      []types.OperationDescriptor
	Clients  []types.ClientDescriptor
}

// NewBaseProvider constructs a BaseProvider with shared metadata
func NewBaseProvider(provider types.ProviderType, caps types.ProviderCapabilities, ops []types.OperationDescriptor, clients []types.ClientDescriptor) BaseProvider {
	return BaseProvider{
		Provider: provider,
		Caps:     caps,
		Ops:      ops,
		Clients:  clients,
	}
}

// Type returns the provider identifier
func (p *BaseProvider) Type() types.ProviderType {
	return p.Provider
}

// Capabilities returns capability flags
func (p *BaseProvider) Capabilities() types.ProviderCapabilities {
	return p.Caps
}

// Operations returns provider-published operations
func (p *BaseProvider) Operations() []types.OperationDescriptor {
	if len(p.Ops) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, len(p.Ops))
	copy(out, p.Ops)
	return out
}

// ClientDescriptors returns provider-published client descriptors
func (p *BaseProvider) ClientDescriptors() []types.ClientDescriptor {
	if len(p.Clients) == 0 {
		return nil
	}

	out := make([]types.ClientDescriptor, len(p.Clients))
	copy(out, p.Clients)
	return out
}
