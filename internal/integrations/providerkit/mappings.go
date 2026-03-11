package providerkit

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
)

// ResolveMappingOverride returns the user-configured MappingOverride for the given schema from
// config.MappingOverrides if present; otherwise returns defaultSpec.
func ResolveMappingOverride(config types.IntegrationConfig, schema types.MappingSchema, defaultSpec types.MappingOverride) types.MappingOverride {
	key := string(normalizeMappingSchema(schema))
	if key == "" {
		return defaultSpec
	}

	override, ok := lo.Find(lo.Entries(config.MappingOverrides), func(e lo.Entry[string, types.MappingOverride]) bool {
		return e.Key == key
	})
	if !ok {
		return defaultSpec
	}

	return override.Value
}
