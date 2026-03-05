package targetresolver

import (
	"context"
	"errors"
	"testing"

	"github.com/samber/mo"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	githubprovider "github.com/theopenlane/core/internal/integrations/providers/github"
	slackprovider "github.com/theopenlane/core/internal/integrations/providers/slack"
	"github.com/theopenlane/core/internal/integrations/types"
)

// integrationSourceStub provides deterministic integration records for resolver tests.
type integrationSourceStub struct {
	byID       map[string]*entgen.Integration
	byProvider map[types.ProviderType][]*entgen.Integration
}

func (s *integrationSourceStub) IntegrationByID(_ context.Context, ownerID string, integrationID string) (*entgen.Integration, error) {
	if ownerID == "error-owner" {
		return nil, errors.New("integration lookup failed")
	}
	if integrationID == "" {
		return nil, nil
	}

	record, ok := s.byID[integrationID]
	if !ok {
		return nil, nil
	}

	return record, nil
}

func (s *integrationSourceStub) IntegrationsByProvider(_ context.Context, ownerID string, provider types.ProviderType) ([]*entgen.Integration, error) {
	if ownerID == "error-owner" {
		return nil, errors.New("integrations by provider failed")
	}

	return append([]*entgen.Integration(nil), s.byProvider[provider]...), nil
}

func TestNewResolverRequiresSource(t *testing.T) {
	_, err := NewResolver(nil)
	if !errors.Is(err, ErrResolverSourceRequired) {
		t.Fatalf("expected ErrResolverSourceRequired, got %v", err)
	}
}

func TestResolveRequiresOwner(t *testing.T) {
	resolver, err := NewResolver(&integrationSourceStub{})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{})
	if !errors.Is(err, ErrResolverOwnerIDRequired) {
		t.Fatalf("expected ErrResolverOwnerIDRequired, got %v", err)
	}
}

func TestResolveRequiresProviderWithoutIntegrationID(t *testing.T) {
	resolver, err := NewResolver(&integrationSourceStub{})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{OwnerID: "org-1"})
	if !errors.Is(err, ErrResolverProviderRequired) {
		t.Fatalf("expected ErrResolverProviderRequired, got %v", err)
	}
}

func TestResolveRejectsEmptyIntegrationID(t *testing.T) {
	resolver, err := NewResolver(&integrationSourceStub{})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org-1",
		IntegrationID: mo.Some(""),
	})
	if !errors.Is(err, ErrResolverIntegrationIDRequired) {
		t.Fatalf("expected ErrResolverIntegrationIDRequired, got %v", err)
	}
}

func TestResolveByIntegrationID(t *testing.T) {
	provider := githubprovider.TypeGitHub
	resolver, err := NewResolver(&integrationSourceStub{
		byID: map[string]*entgen.Integration{
			"int-1": {
				ID:      "int-1",
				Kind:    string(provider),
				OwnerID: "org-1",
			},
		},
	})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	result, err := resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org-1",
		IntegrationID: mo.Some("int-1"),
	})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if result.Provider != provider {
		t.Fatalf("expected provider %s, got %s", provider, result.Provider)
	}
	if result.Integration == nil || result.Integration.ID != "int-1" {
		t.Fatalf("expected integration int-1")
	}
}

func TestResolveByIntegrationIDProviderMismatch(t *testing.T) {
	resolver, err := NewResolver(&integrationSourceStub{
		byID: map[string]*entgen.Integration{
			"int-1": {
				ID:      "int-1",
				Kind:    string(githubprovider.TypeGitHub),
				OwnerID: "org-1",
			},
		},
	})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:       "org-1",
		IntegrationID: mo.Some("int-1"),
		Provider:      mo.Some(slackprovider.TypeSlack),
	})
	if !errors.Is(err, ErrResolverProviderMismatch) {
		t.Fatalf("expected ErrResolverProviderMismatch, got %v", err)
	}
}

func TestResolveByProviderSingleAndAmbiguous(t *testing.T) {
	provider := githubprovider.TypeGitHub
	resolver, err := NewResolver(&integrationSourceStub{
		byProvider: map[types.ProviderType][]*entgen.Integration{
			provider: {
				{ID: "int-1", Kind: string(provider), OwnerID: "org-1"},
			},
		},
	})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	result, err := resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:  "org-1",
		Provider: mo.Some(provider),
	})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if result.Integration == nil || result.Integration.ID != "int-1" {
		t.Fatalf("expected integration int-1")
	}

	resolver, err = NewResolver(&integrationSourceStub{
		byProvider: map[types.ProviderType][]*entgen.Integration{
			provider: {
				{ID: "int-1", Kind: string(provider), OwnerID: "org-1"},
				{ID: "int-2", Kind: string(provider), OwnerID: "org-1"},
			},
		},
	})
	if err != nil {
		t.Fatalf("new resolver failed: %v", err)
	}

	_, err = resolver.Resolve(context.Background(), ResolveCriteria{
		OwnerID:  "org-1",
		Provider: mo.Some(provider),
	})
	if !errors.Is(err, ErrResolverIntegrationAmbiguous) {
		t.Fatalf("expected ErrResolverIntegrationAmbiguous, got %v", err)
	}
}
