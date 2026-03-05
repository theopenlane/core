package providerkit

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// DefaultClientDescriptor returns a descriptor with a default object config schema
func DefaultClientDescriptor(provider types.ProviderType, name types.ClientName, description string, build types.ClientBuilderFunc) types.ClientDescriptor {
	return types.ClientDescriptor{
		Provider:     provider,
		Name:         name,
		Description:  description,
		Build:        build,
		ConfigSchema: map[string]any{"type": "object"},
	}
}

// DefaultClientDescriptors returns a single-descriptor slice with a default object config schema
func DefaultClientDescriptors(provider types.ProviderType, name types.ClientName, description string, build types.ClientBuilderFunc) []types.ClientDescriptor {
	return []types.ClientDescriptor{
		DefaultClientDescriptor(provider, name, description, build),
	}
}

// TokenClientBuilder returns a ClientBuilderFunc that extracts a token and creates an authenticated HTTP client
func TokenClientBuilder(extract func(types.CredentialPayload) (string, error), headers map[string]string) types.ClientBuilderFunc {
	return func(_ context.Context, payload types.CredentialPayload, _ json.RawMessage) (types.ClientInstance, error) {
		token, err := extract(payload)
		if err != nil {
			return types.EmptyClientInstance(), err
		}

		return types.NewClientInstance(auth.NewAuthenticatedClient(token, headers)), nil
	}
}

// SanitizeOperationDescriptors filters and cleans a slice of operation descriptors
func SanitizeOperationDescriptors(provider types.ProviderType, descriptors []types.OperationDescriptor) []types.OperationDescriptor {
	sanitized := sanitizeDescriptors(
		provider,
		descriptors,
		func(d types.OperationDescriptor) bool { return d.Run != nil && d.Name != "" },
		func(d types.OperationDescriptor) types.ProviderType { return d.Provider },
		func(d *types.OperationDescriptor, p types.ProviderType) { d.Provider = p },
	)

	return lo.Map(sanitized, func(descriptor types.OperationDescriptor, _ int) types.OperationDescriptor {
		descriptor.Ingest = sanitizeIngestContracts(descriptor.Ingest)
		return descriptor
	})
}

// SanitizeClientDescriptors filters out invalid client descriptors and assigns provider type
func SanitizeClientDescriptors(provider types.ProviderType, descriptors []types.ClientDescriptor) []types.ClientDescriptor {
	return sanitizeDescriptors(
		provider,
		descriptors,
		func(d types.ClientDescriptor) bool { return d.Build != nil },
		func(d types.ClientDescriptor) types.ProviderType { return d.Provider },
		func(d *types.ClientDescriptor, p types.ProviderType) { d.Provider = p },
	)
}

// sanitizeDescriptors filters descriptors by validity and stamps provider onto entries missing it
func sanitizeDescriptors[T any](provider types.ProviderType, descriptors []T, isValid func(T) bool, getProvider func(T) types.ProviderType, setProvider func(*T, types.ProviderType)) []T {
	if len(descriptors) == 0 {
		return nil
	}

	out := lo.FilterMap(descriptors, func(descriptor T, _ int) (T, bool) {
		if !isValid(descriptor) {
			var zero T
			return zero, false
		}
		if getProvider(descriptor) == types.ProviderUnknown {
			setProvider(&descriptor, provider)
		}

		return descriptor, true
	})
	if len(out) == 0 {
		return nil
	}

	return out
}

func sanitizeIngestContracts(contracts []types.IngestContract) []types.IngestContract {
	return lo.FilterMap(contracts, func(contract types.IngestContract, _ int) (types.IngestContract, bool) {
		contract.Schema = types.NormalizeMappingSchema(contract.Schema)

		return contract, contract.Schema != ""
	})
}
