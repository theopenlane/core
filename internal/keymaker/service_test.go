package keymaker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestService_BeginAndComplete(t *testing.T) {
	ctx := context.Background()
	providerType := types.ProviderType("github")

	provider := &fakeProvider{
		providerType: providerType,
		state:        "state-123",
		authURL:      "https://example.com/auth",
		payload: types.CredentialSet{
			OAuthAccessToken: "token-123",
		},
	}

	keystore := &fakeKeystore{}
	store := NewInMemoryAuthStateStore()

	svc, err := NewService(fakeResolver{provider: provider}, keystore, store, ServiceOptions{})
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	begin, err := svc.BeginAuthorization(ctx, BeginRequest{
		OrgID:         "org-1",
		IntegrationID: "int-1",
		Provider:      providerType,
		Scopes:        []string{"repo"},
		Metadata: json.RawMessage(`{"label":"value"}`),
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
	if result.Credential.OAuthAccessToken != "token-123" {
		t.Fatalf("expected credential data to persist token")
	}
	if len(keystore.integrationSaves) != 1 {
		t.Fatalf("expected integration-scoped keystore call once, got %d", len(keystore.integrationSaves))
	}
	if keystore.integrationSaves[0].orgID != "org-1" {
		t.Fatalf("expected org ID to propagate to keystore")
	}
	if keystore.integrationSaves[0].integrationID != "int-1" {
		t.Fatalf("expected integration ID to propagate to keystore")
	}
}

func TestService_BeginAuthorizationProviderMissing(t *testing.T) {
	svc, err := NewService(fakeResolver{}, &fakeKeystore{}, NewInMemoryAuthStateStore(), ServiceOptions{})
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
		payload:      types.CredentialSet{},
	}

	now := time.Now()
	clock := func() time.Time {
		return now
	}

	svc, err := NewService(fakeResolver{provider: provider}, &fakeKeystore{}, NewInMemoryAuthStateStore(), ServiceOptions{
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
	saves            []saveCall
	integrationSaves []saveIntegrationCall
	err              error
}

type saveCall struct {
	orgID      string
	provider   types.ProviderType
	authKind   types.AuthKind
	credential types.CredentialSet
}

type saveIntegrationCall struct {
	orgID         string
	integrationID string
	provider      types.ProviderType
	authKind      types.AuthKind
	credential    types.CredentialSet
}

func (f *fakeKeystore) SaveCredential(_ context.Context, orgID string, provider types.ProviderType, authKind types.AuthKind, credential types.CredentialSet) (types.CredentialSet, error) {
	f.saves = append(f.saves, saveCall{
		orgID:      orgID,
		provider:   provider,
		authKind:   authKind,
		credential: credential,
	})
	if f.err != nil {
		return types.CredentialSet{}, f.err
	}
	return credential, nil
}

func (f *fakeKeystore) SaveCredentialForIntegration(_ context.Context, orgID string, integrationID string, provider types.ProviderType, authKind types.AuthKind, credential types.CredentialSet) (types.CredentialSet, error) {
	f.integrationSaves = append(f.integrationSaves, saveIntegrationCall{
		orgID:         orgID,
		integrationID: integrationID,
		provider:      provider,
		authKind:      authKind,
		credential:    credential,
	})
	if f.err != nil {
		return types.CredentialSet{}, f.err
	}

	return credential, nil
}

type fakeProvider struct {
	providerType types.ProviderType
	state        string
	authURL      string
	payload      types.CredentialSet
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

func (p *fakeProvider) Mint(context.Context, types.CredentialMintRequest) (types.CredentialSet, error) {
	return types.CredentialSet{}, errors.New("not implemented")
}

type fakeAuthSession struct {
	provider  types.ProviderType
	state     string
	authURL   string
	payload   types.CredentialSet
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

func (s *fakeAuthSession) Finish(context.Context, string) (types.CredentialSet, error) {
	if s.finishErr != nil {
		return types.CredentialSet{}, s.finishErr
	}
	return s.payload, nil
}
