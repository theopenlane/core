// Package mappingtest provides shared test helpers for integration mapping tests.
package mappingtest

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// LoadExample reads a JSON file from the given directory and returns it as a raw message
func LoadExample(t *testing.T, dir, name string) json.RawMessage {
	t.Helper()

	data, err := os.ReadFile(dir + "/" + name)
	assert.NilError(t, err)

	return json.RawMessage(data)
}

// MappingSpec returns the MappingOverride for the first mapping matching schema in the list
func MappingSpec(t *testing.T, mappings []types.MappingRegistration, schema string) types.MappingOverride {
	t.Helper()

	for _, m := range mappings {
		if m.Schema == schema && m.Variant == "" {
			return m.Spec
		}
	}

	t.Fatalf("no mapping found for schema %s", schema)

	return types.MappingOverride{}
}

// MappingSpecForVariant returns the MappingOverride for the mapping matching both schema and variant
func MappingSpecForVariant(t *testing.T, mappings []types.MappingRegistration, schema, variant string) types.MappingOverride {
	t.Helper()

	for _, m := range mappings {
		if m.Schema == schema && m.Variant == variant {
			return m.Spec
		}
	}

	t.Fatalf("no mapping found for schema %s variant %s", schema, variant)

	return types.MappingOverride{}
}

// AssertFiltered evaluates the filter expression against the envelope and returns whether it matched
func AssertFiltered(t *testing.T, spec types.MappingOverride, envelope types.MappingEnvelope) bool {
	t.Helper()

	matched, err := providerkit.EvalFilter(context.Background(), spec.FilterExpr, envelope)
	assert.NilError(t, err)

	return matched
}

// EvalMap evaluates the map expression for the given spec and envelope, returning the result as a map
func EvalMap(t *testing.T, spec types.MappingOverride, envelope types.MappingEnvelope) map[string]any {
	t.Helper()

	raw, err := providerkit.EvalMap(context.Background(), spec.MapExpr, envelope)
	assert.NilError(t, err)

	mapped, err := jsonx.ToMap(raw)
	assert.NilError(t, err)

	return mapped
}
