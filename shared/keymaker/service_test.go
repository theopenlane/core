package keymaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/theopenlane/shared/integrations"
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/types"
	"github.com/theopenlane/shared/models"
)

func TestService_BeginAndComplete(t *testing.T) {
	ctx := context.Background()
	providerType := types.ProviderType("github")

	provider := &fakeProvider{
		providerType: providerType,
		state:        "state-123",
		authURL:      "https://example.com/auth",
		payload: types.CredentialPayload{
			Provider: providerType,
			Kind:     types.CredentialKindOAuthToken,
			Data: models.CredentialSet{
				APIToken: "token-123",
			},
		},
	}

	keystore := &fakeKeystore{}
	store := NewMemorySessionStore()

	svc, err := NewService(fakeResolver{provider: provider}, keystore, store, ServiceOptions{})
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	begin, err := svc.BeginAuthorization(ctx, BeginRequest{
		OrgID:         "org-1",
		IntegrationID: "int-1",
		Provider:      providerType,
		Scopes:        []string{"repo"},
		Metadata: map[string]any{
			"label": "value",
		},
	})
	if err != nil {
		t.Fatalf("BeginAuthorization error: %v", err)
	}
	if begin.AuthURL != provider.authURL {
		t.Fatalf("expected auth url %q, got %q", provider.authURL, begin.AuthURL)
	}

	result, err := svc.CompleteAuthorization(ctx, CompleteRequest{
		State: begin.State,
		Code:  "code-123",
	})
	if err != nil {
		t.Fatalf("CompleteAuthorization error: %v", err)
	}
	if result.Credential.Data.APIToken != "token-123" {
		t.Fatalf("expected credential data to persist token")
	}
	if len(keystore.saves) != 1 {
		t.Fatalf("expected keystore to be called once, got %d", len(keystore.saves))
	}
	if keystore.saves[0].orgID != "org-1" {
		t.Fatalf("expected org ID to propagate to keystore")
	}
}

func TestService_BeginAuthorizationProviderMissing(t *testing.T) {
	svc, err := NewService(fakeResolver{}, &fakeKeystore{}, NewMemorySessionStore(), ServiceOptions{})
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	_, err = svc.BeginAuthorization(context.Background(), BeginRequest{
		OrgID:         "org-1",
		IntegrationID: "int-1",
		Provider:      types.ProviderType("unknown"),
	})
	if !errors.Is(err, integrations.ErrProviderNotFound) {
		t.Fatalf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestService_CompleteAuthorizationExpired(t *testing.T) {
	ctx := context.Background()
	providerType := types.ProviderType("slack")
	provider := &fakeProvider{
		providerType: providerType,
		state:        "state-456",
		authURL:      "https://example.com/authorize",
		payload: types.CredentialPayload{
			Provider: providerType,
			Kind:     types.CredentialKindOAuthToken,
			Data:     models.CredentialSet{},
		},
	}

	now := time.Now()
	clock := func() time.Time {
		return now
	}

	svc, err := NewService(fakeResolver{provider: provider}, &fakeKeystore{}, NewMemorySessionStore(), ServiceOptions{
		SessionTTL: time.Minute,
		Now:        clock,
	})
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	begin, err := svc.BeginAuthorization(ctx, BeginRequest{
		OrgID:         "org-2",
		IntegrationID: "int-2",
		Provider:      providerType,
	})
	if err != nil {
		t.Fatalf("BeginAuthorization error: %v", err)
	}

	now = now.Add(2 * time.Minute)

	_, err = svc.CompleteAuthorization(ctx, CompleteRequest{
		State: begin.State,
		Code:  "code-456",
	})
	if !errors.Is(err, integrations.ErrAuthorizationStateExpired) {
		t.Fatalf("expected ErrAuthorizationStateExpired, got %v", err)
	}
}

type fakeResolver struct {
	provider providers.Provider
}

func (r fakeResolver) Provider(pt types.ProviderType) (types.Provider, bool) {
	if r.provider == nil {
		return nil, false
	}
	if r.provider.Type() != pt {
		return nil, false
	}
	return r.provider, true
}

type fakeKeystore struct {
	saves []saveCall
	err   error
}

type saveCall struct {
	orgID   string
	payload types.CredentialPayload
}

func (f *fakeKeystore) SaveCredential(_ context.Context, orgID string, payload types.CredentialPayload) (types.CredentialPayload, error) {
	f.saves = append(f.saves, saveCall{
		orgID:   orgID,
		payload: payload,
	})
	if f.err != nil {
		return types.CredentialPayload{}, f.err
	}
	return payload, nil
}

type fakeProvider struct {
	providerType types.ProviderType
	state        string
	authURL      string
	payload      types.CredentialPayload
	beginErr     error
	finishErr    error
}

func (p *fakeProvider) Type() types.ProviderType {
	return p.providerType
}

func (p *fakeProvider) Capabilities() types.ProviderCapabilities {
	return types.ProviderCapabilities{}
}

func (p *fakeProvider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	if p.beginErr != nil {
		return nil, p.beginErr
	}
	return &fakeAuthSession{
		provider:  p.providerType,
		state:     p.state,
		authURL:   p.authURL,
		payload:   p.payload,
		finishErr: p.finishErr,
	}, nil
}

func (p *fakeProvider) Mint(context.Context, types.CredentialSubject) (types.CredentialPayload, error) {
	return types.CredentialPayload{}, errors.New("not implemented")
}

type fakeAuthSession struct {
	provider  types.ProviderType
	state     string
	authURL   string
	payload   types.CredentialPayload
	finishErr error
}

func (s *fakeAuthSession) ProviderType() types.ProviderType {
	return s.provider
}

func (s *fakeAuthSession) State() string {
	return s.state
}

func (s *fakeAuthSession) AuthURL() string {
	return s.authURL
}

func (s *fakeAuthSession) Finish(context.Context, string) (types.CredentialPayload, error) {
	if s.finishErr != nil {
		return types.CredentialPayload{}, s.finishErr
	}
	return s.payload, nil
}
