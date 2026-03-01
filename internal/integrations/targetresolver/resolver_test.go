package targetresolver

import (
	"context"
	"errors"
	"testing"

	"github.com/samber/mo"

	"github.com/theopenlane/core/common/integrations/types"
	entgen "github.com/theopenlane/core/internal/ent/generated"
)

// integrationSourceStub provides deterministic integration records for resolver tests
type integrationSourceStub struct {
	integrationByID           map[string]*entgen.Integration
	integrationsByProvider    map[types.ProviderType][]*entgen.Integration
	integrationByIDErr        error
	integrationsByProviderErr error
}

// IntegrationByID resolves one integration from the stub map
func (s *integrationSourceStub) IntegrationByID(_ context.Context, _ string, integrationID string) (*entgen.Integration, error) {
	if s.integrationByIDErr != nil {
		return nil, s.integrationByIDErr
	}

	record, ok := s.integrationByID[integrationID]
	if !ok {
		return nil, nil
	}

	return record, nil
}

// IntegrationsByProvider resolves integrations from the stub provider map
func (s *integrationSourceStub) IntegrationsByProvider(_ context.Context, _ string, provider types.ProviderType) ([]*entgen.Integration, error) {
	if s.integrationsByProviderErr != nil {
		return nil, s.integrationsByProviderErr
	}

	return s.integrationsByProvider[provider], nil
}

// operationRegistryStub provides deterministic operation descriptors for tests
type operationRegistryStub struct {
	descriptors map[types.ProviderType][]types.OperationDescriptor
}

// OperationDescriptors returns provider descriptors from the stub map
func (r operationRegistryStub) OperationDescriptors(provider types.ProviderType) []types.OperationDescriptor {
	return r.descriptors[provider]
}

// TestNewResolverRequiresDependencies verifies resolver constructor dependency validation
func TestNewResolverRequiresDependencies(t *testing.T) {
	_, err := NewResolver(nil, operationRegistryStub{})
	if !errors.Is(err, ErrResolverSourceRequired) {
		t.Fatalf("expected ErrResolverSourceRequired, got %v", err)
	}

	_, err = NewResolver(&integrationSourceStub{}, nil)
	if !errors.Is(err, ErrResolverRegistryRequired) {
		t.Fatalf("expected ErrResolverRegistryRequired, got %v", err)
	}
}

// TestResolveRequiresOwnerID verifies owner id validation
func TestResolveRequiresOwnerID(t *testing.T) {
	resolver, err := NewResolver(&integrationSourceStub{}, operationRegistryStub{})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{})
	if !errors.Is(err, ErrResolverOwnerIDRequired) {
		t.Fatalf("expected ErrResolverOwnerIDRequired, got %v", err)
	}
}

// TestResolveRequiresOperationCriteria verifies operation criteria validation
func TestResolveRequiresOperationCriteria(t *testing.T) {
	resolver, err := NewResolver(&integrationSourceStub{}, operationRegistryStub{})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID: "org_123",
	})
	if !errors.Is(err, ErrResolverOperationCriteriaRequired) {
		t.Fatalf("expected ErrResolverOperationCriteriaRequired, got %v", err)
	}
}

// TestResolveRequiresProviderWithoutIntegrationID verifies provider requirement when no integration id is specified
func TestResolveRequiresProviderWithoutIntegrationID(t *testing.T) {
	resolver, err := NewResolver(&integrationSourceStub{}, operationRegistryStub{})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		OperationName: mo.Some(types.OperationName("message.send")),
	})
	if !errors.Is(err, ErrResolverProviderRequired) {
		t.Fatalf("expected ErrResolverProviderRequired, got %v", err)
	}
}

// TestResolveRejectsEmptyIntegrationID verifies integration id validation
func TestResolveRejectsEmptyIntegrationID(t *testing.T) {
	resolver, err := NewResolver(&integrationSourceStub{}, operationRegistryStub{})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		IntegrationID: mo.Some(""),
		OperationName: mo.Some(types.OperationName("message.send")),
	})
	if !errors.Is(err, ErrResolverIntegrationIDRequired) {
		t.Fatalf("expected ErrResolverIntegrationIDRequired, got %v", err)
	}
}

// TestResolveWithIntegrationIDDerivesProvider verifies provider resolution from integration kind
func TestResolveWithIntegrationIDDerivesProvider(t *testing.T) {
	resolver, err := NewResolver(
		&integrationSourceStub{
			integrationByID: map[string]*entgen.Integration{
				"int_123": {
					ID:   "int_123",
					Kind: "slack",
				},
			},
		},
		operationRegistryStub{
			descriptors: map[types.ProviderType][]types.OperationDescriptor{
				"slack": {
					{
						Provider: "slack",
						Name:     "message.send",
						Kind:     types.OperationKindNotify,
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	out, err := resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		IntegrationID: mo.Some("int_123"),
		OperationName: mo.Some(types.OperationName("message.send")),
	})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if out.Provider != "slack" {
		t.Fatalf("expected provider slack, got %s", out.Provider)
	}
	if out.Integration == nil || out.Integration.ID != "int_123" {
		t.Fatalf("expected integration int_123")
	}
	if out.Operation.Name != "message.send" {
		t.Fatalf("expected operation message.send, got %s", out.Operation.Name)
	}
}

// TestResolveWithIntegrationIDProviderMismatch verifies provider mismatch detection
func TestResolveWithIntegrationIDProviderMismatch(t *testing.T) {
	resolver, err := NewResolver(
		&integrationSourceStub{
			integrationByID: map[string]*entgen.Integration{
				"int_123": {
					ID:   "int_123",
					Kind: "slack",
				},
			},
		},
		operationRegistryStub{},
	)
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		IntegrationID: mo.Some("int_123"),
		Provider:      mo.Some(types.ProviderType("github")),
		OperationName: mo.Some(types.OperationName("message.send")),
	})
	if !errors.Is(err, ErrResolverProviderMismatch) {
		t.Fatalf("expected ErrResolverProviderMismatch, got %v", err)
	}
}

// TestResolveWithIntegrationIDNotFound verifies missing integration handling
func TestResolveWithIntegrationIDNotFound(t *testing.T) {
	resolver, err := NewResolver(
		&integrationSourceStub{
			integrationByID: map[string]*entgen.Integration{},
		},
		operationRegistryStub{
			descriptors: map[types.ProviderType][]types.OperationDescriptor{
				"slack": {
					{
						Provider: "slack",
						Name:     "message.send",
						Kind:     types.OperationKindNotify,
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		IntegrationID: mo.Some("int_missing"),
		OperationName: mo.Some(types.OperationName("message.send")),
	})
	if !errors.Is(err, ErrResolverIntegrationNotFound) {
		t.Fatalf("expected ErrResolverIntegrationNotFound, got %v", err)
	}
}

// TestResolveByProviderAndOperationKind verifies provider plus kind selection
func TestResolveByProviderAndOperationKind(t *testing.T) {
	resolver, err := NewResolver(
		&integrationSourceStub{
			integrationsByProvider: map[types.ProviderType][]*entgen.Integration{
				"slack": {
					{
						ID:   "int_123",
						Kind: "slack",
					},
				},
			},
		},
		operationRegistryStub{
			descriptors: map[types.ProviderType][]types.OperationDescriptor{
				"slack": {
					{
						Provider: "slack",
						Name:     "health.default",
						Kind:     types.OperationKindHealth,
					},
					{
						Provider: "slack",
						Name:     "message.send",
						Kind:     types.OperationKindNotify,
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	out, err := resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		Provider:      mo.Some(types.ProviderType("slack")),
		OperationKind: mo.Some(types.OperationKindNotify),
	})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if out.Operation.Name != "message.send" {
		t.Fatalf("expected operation message.send, got %s", out.Operation.Name)
	}
}

// TestResolveByProviderAmbiguousIntegration verifies ambiguity detection for multiple installed integrations
func TestResolveByProviderAmbiguousIntegration(t *testing.T) {
	resolver, err := NewResolver(
		&integrationSourceStub{
			integrationsByProvider: map[types.ProviderType][]*entgen.Integration{
				"slack": {
					{ID: "int_1", Kind: "slack"},
					{ID: "int_2", Kind: "slack"},
				},
			},
		},
		operationRegistryStub{
			descriptors: map[types.ProviderType][]types.OperationDescriptor{
				"slack": {
					{
						Provider: "slack",
						Name:     "message.send",
						Kind:     types.OperationKindNotify,
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		Provider:      mo.Some(types.ProviderType("slack")),
		OperationName: mo.Some(types.OperationName("message.send")),
	})
	if !errors.Is(err, ErrResolverIntegrationAmbiguous) {
		t.Fatalf("expected ErrResolverIntegrationAmbiguous, got %v", err)
	}
}

// TestResolveByOperationKindAmbiguousDescriptor verifies ambiguity detection for kind-only descriptor resolution
func TestResolveByOperationKindAmbiguousDescriptor(t *testing.T) {
	resolver, err := NewResolver(
		&integrationSourceStub{
			integrationsByProvider: map[types.ProviderType][]*entgen.Integration{
				"slack": {
					{ID: "int_1", Kind: "slack"},
				},
			},
		},
		operationRegistryStub{
			descriptors: map[types.ProviderType][]types.OperationDescriptor{
				"slack": {
					{
						Provider: "slack",
						Name:     "message.send",
						Kind:     types.OperationKindNotify,
					},
					{
						Provider: "slack",
						Name:     "message.alt",
						Kind:     types.OperationKindNotify,
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		Provider:      mo.Some(types.ProviderType("slack")),
		OperationKind: mo.Some(types.OperationKindNotify),
	})
	if !errors.Is(err, ErrResolverOperationDescriptorAmbiguous) {
		t.Fatalf("expected ErrResolverOperationDescriptorAmbiguous, got %v", err)
	}
}

// TestResolveWithOperationNameAndKindMismatch verifies mismatch detection between explicit name and kind constraints
func TestResolveWithOperationNameAndKindMismatch(t *testing.T) {
	resolver, err := NewResolver(
		&integrationSourceStub{
			integrationsByProvider: map[types.ProviderType][]*entgen.Integration{
				"slack": {
					{ID: "int_1", Kind: "slack"},
				},
			},
		},
		operationRegistryStub{
			descriptors: map[types.ProviderType][]types.OperationDescriptor{
				"slack": {
					{
						Provider: "slack",
						Name:     "message.send",
						Kind:     types.OperationKindNotify,
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		Provider:      mo.Some(types.ProviderType("slack")),
		OperationName: mo.Some(types.OperationName("message.send")),
		OperationKind: mo.Some(types.OperationKindHealth),
	})
	if !errors.Is(err, ErrResolverOperationKindMismatch) {
		t.Fatalf("expected ErrResolverOperationKindMismatch, got %v", err)
	}
}

// TestResolveReturnsOperationNotRegistered verifies operation registration validation
func TestResolveReturnsOperationNotRegistered(t *testing.T) {
	resolver, err := NewResolver(
		&integrationSourceStub{
			integrationsByProvider: map[types.ProviderType][]*entgen.Integration{
				"slack": {
					{ID: "int_1", Kind: "slack"},
				},
			},
		},
		operationRegistryStub{
			descriptors: map[types.ProviderType][]types.OperationDescriptor{
				"slack": {
					{
						Provider: "slack",
						Name:     "health.default",
						Kind:     types.OperationKindHealth,
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org_123",
		Provider:      mo.Some(types.ProviderType("slack")),
		OperationName: mo.Some(types.OperationName("message.send")),
	})
	if !errors.Is(err, ErrResolverOperationNotRegistered) {
		t.Fatalf("expected ErrResolverOperationNotRegistered, got %v", err)
	}
}
