package operations

import (
	"context"
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
		{Name: "ok", Run: run, Provider: types.ProviderUnknown},
	}

	out := SanitizeOperationDescriptors(provider, descriptors)
	if len(out) != 1 {
		t.Fatalf("expected 1 descriptor, got %d", len(out))
	}
	if out[0].Provider != provider {
		t.Fatalf("expected provider to be set")
	}
}

func TestSanitizeClientDescriptors(t *testing.T) {
	provider := types.ProviderType("test")
	build := func(context.Context, types.CredentialPayload, map[string]any) (any, error) {
		return nil, nil
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
