package providerkit

import (
	"context"
	"encoding/json"
	"maps"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// defaultObjectSchema is the default JSON schema for client configuration
var defaultObjectSchema = json.RawMessage(`{"type":"object"}`)

// BearerClient holds a bearer token and static headers for authenticated HTTP calls.
// Providers unwrap this from a ClientInstance using types.ClientInstanceAs[*BearerClient].
type BearerClient struct {
	// BearerToken is the authorization bearer token
	BearerToken string
	// Headers contains additional static HTTP headers sent with each request
	Headers map[string]string
}

// DefaultClientDescriptor returns a descriptor with a default object config schema
func DefaultClientDescriptor(provider types.ProviderType, name types.ClientName, description string, build types.ClientBuilderFunc) types.ClientDescriptor {
	return types.ClientDescriptor{
		Provider:     provider,
		Name:         name,
		Description:  description,
		Build:        build,
		ConfigSchema: jsonx.CloneRawMessage(defaultObjectSchema),
	}
}

// DefaultClientDescriptors returns a single-element slice containing a descriptor with a default object config schema
func DefaultClientDescriptors(provider types.ProviderType, name types.ClientName, description string, build types.ClientBuilderFunc) []types.ClientDescriptor {
	return []types.ClientDescriptor{
		DefaultClientDescriptor(provider, name, description, build),
	}
}

// TokenClientBuilder returns a ClientBuilderFunc that extracts a bearer token from a CredentialSet
// and wraps a BearerClient with the token and provided static headers
func TokenClientBuilder(extract func(types.CredentialSet) (string, error), headers map[string]string) types.ClientBuilderFunc {
	return func(_ context.Context, credential types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
		token, err := extract(credential)
		if err != nil {
			return types.EmptyClientInstance(), err
		}

		return types.NewClientInstance(&BearerClient{
			BearerToken: token,
			Headers:     maps.Clone(headers),
		}), nil
	}
}

// SanitizeOperationDescriptors filters out invalid operation descriptors and stamps the provider
// type on entries that are missing it. Ingest contracts within each descriptor are also normalized.
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

// SanitizeClientDescriptors filters out invalid client descriptors and stamps the provider type
// on entries that are missing it
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
func sanitizeDescriptors[T any](
	provider types.ProviderType,
	descriptors []T,
	isValid func(T) bool,
	getProvider func(T) types.ProviderType,
	setProvider func(*T, types.ProviderType),
) []T {
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

// sanitizeIngestContracts normalizes schema names and filters out contracts with empty schemas
func sanitizeIngestContracts(contracts []types.IngestContract) []types.IngestContract {
	return lo.FilterMap(contracts, func(contract types.IngestContract, _ int) (types.IngestContract, bool) {
		contract.Schema = normalizeMappingSchema(contract.Schema)
		return contract, contract.Schema != ""
	})
}

// normalizeMappingSchema trims whitespace from a schema name and returns the result
func normalizeMappingSchema(schema types.MappingSchema) types.MappingSchema {
	value := strings.TrimSpace(string(schema))
	if value == "" {
		return ""
	}

	return types.MappingSchema(value)
}
