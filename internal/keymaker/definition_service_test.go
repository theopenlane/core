package keymaker

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestService_BeginAndComplete(t *testing.T) {
	ctx := context.Background()

	definitionID := "github-oauth"
	installationID := "install-1"

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: definitionID, Slug: "github-oauth", Version: "1.0"},
		Auth: &types.AuthRegistration{
			Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
				return types.AuthStartResult{
					URL:   "https://github.com/login/oauth/authorize",
					State: json.RawMessage(`{"csrf":"abc123"}`),
				}, nil
			},
			Complete: func(_ context.Context, _ json.RawMessage, _ json.RawMessage) (types.AuthCompleteResult, error) {
				return types.AuthCompleteResult{
					Credential: types.CredentialSet{OAuthAccessToken: "token-xyz"},
				}, nil
			},
		},
	}

	writer := &fakeInstallationWriter{}
	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, writer.SaveInstallationCredential, matchingInstallationResolver(installationID, definitionID).ResolveInstallation, NewInMemoryAuthStateStore(), 0)

	begin, err := svc.BeginAuth(ctx, BeginRequest{
		DefinitionID:   definitionID,
		InstallationID: installationID,
	})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	if begin.AuthURL != "https://github.com/login/oauth/authorize" {
		t.Fatalf("unexpected auth URL: %q", begin.AuthURL)
	}

	if begin.State == "" {
		t.Fatal("expected non-empty state token")
	}

	if begin.DefinitionID != definitionID {
		t.Fatalf("expected definition ID %q, got %q", definitionID, begin.DefinitionID)
	}

	result, err := svc.CompleteAuth(ctx, CompleteRequest{
		State: begin.State,
		Input: json.RawMessage(`{"code":"code-123"}`),
	})
	if err != nil {
		t.Fatalf("CompleteAuth error: %v", err)
	}

	if result.Credential.OAuthAccessToken != "token-xyz" {
		t.Fatalf("expected credential token, got %q", result.Credential.OAuthAccessToken)
	}

	if result.InstallationID != installationID {
		t.Fatalf("expected installation ID %q, got %q", installationID, result.InstallationID)
	}

	if len(writer.saves) != 1 {
		t.Fatalf("expected one credential save, got %d", len(writer.saves))
	}

	if writer.saves[0].installationID != installationID {
		t.Fatalf("expected installation ID %q in save, got %q", installationID, writer.saves[0].installationID)
	}
}

func TestService_BeginAuthRequiresDefinitionAndInstallation(t *testing.T) {
	t.Parallel()

	svc := NewService((&fakeDefinitionResolver{}).Definition, (&fakeInstallationWriter{}).SaveInstallationCredential, (&fakeInstallationResolver{}).ResolveInstallation, NewInMemoryAuthStateStore(), 0)

	_, err := svc.BeginAuth(context.Background(), BeginRequest{InstallationID: "i"})
	if !errors.Is(err, ErrDefinitionIDRequired) {
		t.Fatalf("expected ErrDefinitionIDRequired, got %v", err)
	}

	_, err = svc.BeginAuth(context.Background(), BeginRequest{DefinitionID: "d"})
	if !errors.Is(err, ErrInstallationIDRequired) {
		t.Fatalf("expected ErrInstallationIDRequired, got %v", err)
	}
}

func TestService_BeginAuthDefinitionNotFound(t *testing.T) {
	t.Parallel()

	svc := NewService((&fakeDefinitionResolver{}).Definition, (&fakeInstallationWriter{}).SaveInstallationCredential, (&fakeInstallationResolver{}).ResolveInstallation, NewInMemoryAuthStateStore(), 0)

	_, err := svc.BeginAuth(context.Background(), BeginRequest{
		DefinitionID:   "missing",
		InstallationID: "install-1",
	})
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}

func TestService_BeginAuthNoAuthRegistration(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "no-auth", Slug: "no-auth", Version: "1.0"},
	}

	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, (&fakeInstallationWriter{}).SaveInstallationCredential, matchingInstallationResolver("i", "no-auth").ResolveInstallation, NewInMemoryAuthStateStore(), 0)

	_, err := svc.BeginAuth(context.Background(), BeginRequest{
		DefinitionID:   "no-auth",
		InstallationID: "i",
	})
	if !errors.Is(err, ErrDefinitionAuthRequired) {
		t.Fatalf("expected ErrDefinitionAuthRequired, got %v", err)
	}
}

func TestService_BeginAuthUsesCustomStateToken(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "d1", Slug: "d1", Version: "1.0"},
		Auth: &types.AuthRegistration{
			Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
				return types.AuthStartResult{URL: "https://example.com"}, nil
			},
		},
	}

	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, (&fakeInstallationWriter{}).SaveInstallationCredential, matchingInstallationResolver("i1", "d1").ResolveInstallation, NewInMemoryAuthStateStore(), 0)

	begin, err := svc.BeginAuth(context.Background(), BeginRequest{
		DefinitionID:   "d1",
		InstallationID: "i1",
		State:          "custom-csrf-token",
	})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	if begin.State != "custom-csrf-token" {
		t.Fatalf("expected custom state token, got %q", begin.State)
	}
}

func TestService_BeginAuthInstallationDefinitionMismatch(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "d1", Slug: "d1", Version: "1.0"},
		Auth: &types.AuthRegistration{
			Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
				return types.AuthStartResult{URL: "https://example.com"}, nil
			},
		},
	}

	svc := NewService(
		(&fakeDefinitionResolver{def: def}).Definition,
		(&fakeInstallationWriter{}).SaveInstallationCredential,
		matchingInstallationResolver("i1", "other-definition").ResolveInstallation,
		NewInMemoryAuthStateStore(),
		0,
	)

	_, err := svc.BeginAuth(context.Background(), BeginRequest{
		DefinitionID:   "d1",
		InstallationID: "i1",
	})
	if !errors.Is(err, ErrInstallationDefinitionMismatch) {
		t.Fatalf("expected ErrInstallationDefinitionMismatch, got %v", err)
	}
}

func TestService_CompleteAuthExpired(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	clock := func() time.Time { return now }

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "slack", Slug: "slack", Version: "1.0"},
		Auth: &types.AuthRegistration{
			Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
				return types.AuthStartResult{URL: "https://slack.com/oauth"}, nil
			},
			Complete: func(_ context.Context, _ json.RawMessage, _ json.RawMessage) (types.AuthCompleteResult, error) {
				return types.AuthCompleteResult{}, nil
			},
		},
	}

	svc := NewService(
		(&fakeDefinitionResolver{def: def}).Definition,
		(&fakeInstallationWriter{}).SaveInstallationCredential,
		matchingInstallationResolver("install-2", "slack").ResolveInstallation,
		NewInMemoryAuthStateStore(),
		time.Minute,
	)
	svc.now = clock

	begin, err := svc.BeginAuth(ctx, BeginRequest{
		DefinitionID:   "slack",
		InstallationID: "install-2",
	})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	now = now.Add(2 * time.Minute)

	_, err = svc.CompleteAuth(ctx, CompleteRequest{State: begin.State})
	if !errors.Is(err, ErrAuthStateExpired) {
		t.Fatalf("expected ErrAuthStateExpired, got %v", err)
	}
}

func TestService_CompleteAuthStateTokenRequired(t *testing.T) {
	t.Parallel()

	svc := NewService((&fakeDefinitionResolver{}).Definition, (&fakeInstallationWriter{}).SaveInstallationCredential, (&fakeInstallationResolver{}).ResolveInstallation, NewInMemoryAuthStateStore(), 0)

	_, err := svc.CompleteAuth(context.Background(), CompleteRequest{})
	if !errors.Is(err, ErrAuthStateTokenRequired) {
		t.Fatalf("expected ErrAuthStateTokenRequired, got %v", err)
	}
}

func TestService_CompleteAuthSaveError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "okta", Slug: "okta", Version: "1.0"},
		Auth: &types.AuthRegistration{
			Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
				return types.AuthStartResult{URL: "https://okta.com"}, nil
			},
			Complete: func(_ context.Context, _ json.RawMessage, _ json.RawMessage) (types.AuthCompleteResult, error) {
				return types.AuthCompleteResult{Credential: types.CredentialSet{}}, nil
			},
		},
	}

	writer := &fakeInstallationWriter{err: errors.New("db unavailable")}
	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, writer.SaveInstallationCredential, matchingInstallationResolver("install-3", "okta").ResolveInstallation, NewInMemoryAuthStateStore(), 0)

	begin, err := svc.BeginAuth(ctx, BeginRequest{
		DefinitionID:   "okta",
		InstallationID: "install-3",
	})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	_, err = svc.CompleteAuth(ctx, CompleteRequest{State: begin.State})
	if err == nil || !strings.Contains(err.Error(), "keymaker: save definition credential") {
		t.Fatalf("expected wrapped save error, got %v", err)
	}
}

func TestService_CallbackStatePassedToComplete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var receivedCallbackState json.RawMessage

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "az", Slug: "az", Version: "1.0"},
		Auth: &types.AuthRegistration{
			Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
				return types.AuthStartResult{
					URL:   "https://login.microsoftonline.com",
					State: json.RawMessage(`{"nonce":"n1","tenant":"t1"}`),
				}, nil
			},
			Complete: func(_ context.Context, state json.RawMessage, _ json.RawMessage) (types.AuthCompleteResult, error) {
				receivedCallbackState = state
				return types.AuthCompleteResult{Credential: types.CredentialSet{OAuthAccessToken: "az-token"}}, nil
			},
		},
	}

	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, (&fakeInstallationWriter{}).SaveInstallationCredential, matchingInstallationResolver("i1", "az").ResolveInstallation, NewInMemoryAuthStateStore(), 0)

	begin, err := svc.BeginAuth(ctx, BeginRequest{DefinitionID: "az", InstallationID: "i1"})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	_, err = svc.CompleteAuth(ctx, CompleteRequest{State: begin.State, Input: json.RawMessage(`{"code":"c1"}`)})
	if err != nil {
		t.Fatalf("CompleteAuth error: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(receivedCallbackState, &parsed); err != nil {
		t.Fatalf("expected valid callback state, got error: %v", err)
	}

	if parsed["nonce"] != "n1" || parsed["tenant"] != "t1" {
		t.Fatalf("expected callback state to contain nonce and tenant, got %v", parsed)
	}
}

func TestService_CompleteAuthInstallationDefinitionMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	installations := matchingInstallationResolver("i1", "az")

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "az", Slug: "az", Version: "1.0"},
		Auth: &types.AuthRegistration{
			Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
				return types.AuthStartResult{URL: "https://login.microsoftonline.com"}, nil
			},
			Complete: func(_ context.Context, state json.RawMessage, _ json.RawMessage) (types.AuthCompleteResult, error) {
				return types.AuthCompleteResult{Credential: types.CredentialSet{OAuthAccessToken: string(state)}}, nil
			},
		},
	}

	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, (&fakeInstallationWriter{}).SaveInstallationCredential, installations.ResolveInstallation, NewInMemoryAuthStateStore(), 0)

	begin, err := svc.BeginAuth(ctx, BeginRequest{DefinitionID: "az", InstallationID: "i1"})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	installations.installation.DefinitionID = "changed"

	_, err = svc.CompleteAuth(ctx, CompleteRequest{State: begin.State, Input: json.RawMessage(`{"code":"c1"}`)})
	if !errors.Is(err, ErrInstallationDefinitionMismatch) {
		t.Fatalf("expected ErrInstallationDefinitionMismatch, got %v", err)
	}
}

type fakeDefinitionResolver struct {
	def types.Definition
}

func (r *fakeDefinitionResolver) Definition(id string) (types.Definition, bool) {
	if r.def.ID == "" || r.def.ID != id {
		return types.Definition{}, false
	}

	return r.def, true
}

type installationSave struct {
	installationID string
	credential     types.CredentialSet
}

type fakeInstallationResolver struct {
	installation InstallationRecord
	err          error
}

func (f *fakeInstallationResolver) ResolveInstallation(context.Context, string) (InstallationRecord, error) {
	if f.err != nil {
		return InstallationRecord{}, f.err
	}

	return f.installation, nil
}

type fakeInstallationWriter struct {
	saves []installationSave
	err   error
}

func (f *fakeInstallationWriter) SaveInstallationCredential(_ context.Context, installationID string, credential types.CredentialSet) error {
	f.saves = append(f.saves, installationSave{installationID: installationID, credential: credential})
	return f.err
}

func matchingInstallationResolver(installationID string, definitionID string) *fakeInstallationResolver {
	return &fakeInstallationResolver{
		installation: InstallationRecord{
			ID:           installationID,
			OwnerID:      "org-1",
			DefinitionID: definitionID,
		},
	}
}
