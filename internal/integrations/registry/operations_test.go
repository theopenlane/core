package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/theopenlane/core/internal/integrations/types"
)

func testRun(_ context.Context, _ types.OperationInput) (types.OperationResult, error) {
	return types.OperationResult{}, nil
}

func TestResolveOperationByName(t *testing.T) {
	provider := types.ProviderType("test")
	reg := &Registry{
		operations: map[types.ProviderType][]types.OperationDescriptor{
			provider: {
				{Name: "collect", Kind: types.OperationKindCollectFindings, Run: testRun},
			},
		},
	}

	out, err := reg.ResolveOperation(provider, "collect", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Name != "collect" {
		t.Fatalf("expected collect operation, got %s", out.Name)
	}
}

func TestResolveOperationRequiresCriteria(t *testing.T) {
	reg := &Registry{}

	_, err := reg.ResolveOperation(types.ProviderType("test"), "", "")
	if !errors.Is(err, ErrOperationCriteriaRequired) {
		t.Fatalf("expected ErrOperationCriteriaRequired, got %v", err)
	}
}

func TestResolveOperationKindMismatch(t *testing.T) {
	provider := types.ProviderType("test")
	reg := &Registry{
		operations: map[types.ProviderType][]types.OperationDescriptor{
			provider: {
				{Name: "notify", Kind: types.OperationKindNotify, Run: testRun},
			},
		},
	}

	_, err := reg.ResolveOperation(provider, "notify", types.OperationKindCollectFindings)
	if !errors.Is(err, ErrOperationKindMismatch) {
		t.Fatalf("expected ErrOperationKindMismatch, got %v", err)
	}
}

func TestResolveOperationAmbiguousByKind(t *testing.T) {
	provider := types.ProviderType("test")
	reg := &Registry{
		operations: map[types.ProviderType][]types.OperationDescriptor{
			provider: {
				{Name: "collect.a", Kind: types.OperationKindCollectFindings, Run: testRun},
				{Name: "collect.b", Kind: types.OperationKindCollectFindings, Run: testRun},
			},
		},
	}

	_, err := reg.ResolveOperation(provider, "", types.OperationKindCollectFindings)
	if !errors.Is(err, ErrOperationDescriptorAmbiguous) {
		t.Fatalf("expected ErrOperationDescriptorAmbiguous, got %v", err)
	}
}
