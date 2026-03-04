package operations

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/theopenlane/core/common/integrations/types"
)

func TestSanitizeOperationDescriptors(t *testing.T) {
	provider := types.ProviderType("test")
	run := func(context.Context, types.OperationInput) (types.OperationResult, error) {
		return types.OperationResult{}, nil
	}

	descriptors := []types.OperationDescriptor{
		{Name: "", Run: run},
		{Name: "missing-run"},
		{
			Name:     "ok",
			Run:      run,
			Provider: types.ProviderUnknown,
			Ingest: []types.IngestContract{
				{Schema: "   "},
				{Schema: types.MappingSchemaVulnerability, EnsurePayloads: true},
			},
		},
	}

	out := SanitizeOperationDescriptors(provider, descriptors)
	if len(out) != 1 {
		t.Fatalf("expected 1 descriptor, got %d", len(out))
	}
	if out[0].Provider != provider {
		t.Fatalf("expected provider to be set")
	}
	if len(out[0].Ingest) != 1 {
		t.Fatalf("expected one ingest contract, got %d", len(out[0].Ingest))
	}
	if out[0].Ingest[0].Schema != types.MappingSchemaVulnerability {
		t.Fatalf("expected normalized ingest schema")
	}
}

func TestSanitizeClientDescriptors(t *testing.T) {
	provider := types.ProviderType("test")
	build := func(context.Context, types.CredentialPayload, json.RawMessage) (types.ClientInstance, error) {
		return types.EmptyClientInstance(), nil
	}

	descriptors := []types.ClientDescriptor{
		{Name: "missing-build"},
		{Name: "ok", Build: build, Provider: types.ProviderUnknown},
	}

	out := SanitizeClientDescriptors(provider, descriptors)
	if len(out) != 1 {
		t.Fatalf("expected 1 descriptor, got %d", len(out))
	}
	if out[0].Provider != provider {
		t.Fatalf("expected provider to be set")
	}
}
