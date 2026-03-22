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

var keymakerTestCredentialRef = types.NewCredentialRef("test_auth")

func TestService_BeginAndComplete(t *testing.T) {
	ctx := context.Background()

	definitionID := "github-oauth"
	installationID := "install-1"

	def := authTestDefinition(definitionID, "github-oauth", &fakeAuthFlow{
		startResult: types.AuthStartResult{
			URL:   "https://github.com/login/oauth/authorize",
			State: json.RawMessage(`{"csrf":"abc123"}`),
		},
		completeResult: types.AuthCompleteResult{
			Credential: types.CredentialSet{Data: json.RawMessage(`{"accessToken":"token-xyz"}`)},
		},
	})

	writer := &fakeInstallationWriter{}
	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, writer.PersistAuthResult, matchingInstallationResolver(installationID, definitionID).ResolveInstallation, NewInMemoryAuthStateStore())

	begin, err := svc.BeginAuth(ctx, BeginRequest{
		DefinitionID:   definitionID,
		InstallationID: installationID,
		CredentialRef:  keymakerTestCredentialRef,
	})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	if begin.AuthURL != "https://github.com/login/oauth/authorize" {
		t.Fatalf("unexpected auth URL: %q", begin.AuthURL)
	}

	if begin.State == "" {
		t.Fatalf("expected generated state token")
	}

	if begin.DefinitionID != definitionID {
		t.Fatalf("expected definition ID %q, got %q", definitionID, begin.DefinitionID)
	}

	result, err := svc.CompleteAuth(ctx, CompleteRequest{
		State:    begin.State,
		Callback: types.AuthCallbackInput{},
	})
	if err != nil {
		t.Fatalf("CompleteAuth error: %v", err)
	}

	if string(result.Credential.Data) != `{"accessToken":"token-xyz"}` {
		t.Fatalf("expected credential data, got %q", string(result.Credential.Data))
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

	svc := NewService((&fakeDefinitionResolver{}).Definition, (&fakeInstallationWriter{}).PersistAuthResult, (&fakeInstallationResolver{}).ResolveInstallation, NewInMemoryAuthStateStore())

	_, err := svc.BeginAuth(context.Background(), BeginRequest{InstallationID: "i"})
	if !errors.Is(err, ErrDefinitionIDRequired) {
		t.Fatalf("expected ErrDefinitionIDRequired, got %v", err)
	}

	_, err = svc.BeginAuth(context.Background(), BeginRequest{DefinitionID: "d"})
	if !errors.Is(err, ErrInstallationIDRequired) {
		t.Fatalf("expected ErrInstallationIDRequired, got %v", err)
	}

	def := authTestDefinition("d", "d", &fakeAuthFlow{startResult: types.AuthStartResult{URL: "https://example.com"}})
	svc = NewService((&fakeDefinitionResolver{def: def}).Definition, (&fakeInstallationWriter{}).PersistAuthResult, matchingInstallationResolver("i", "d").ResolveInstallation, NewInMemoryAuthStateStore())

	_, err = svc.BeginAuth(context.Background(), BeginRequest{DefinitionID: "d", InstallationID: "i"})
	if !errors.Is(err, ErrConnectionNotFound) {
		t.Fatalf("expected ErrConnectionNotFound, got %v", err)
	}
}

func TestService_BeginAuthDefinitionNotFound(t *testing.T) {
	t.Parallel()

	svc := NewService((&fakeDefinitionResolver{}).Definition, (&fakeInstallationWriter{}).PersistAuthResult, (&fakeInstallationResolver{}).ResolveInstallation, NewInMemoryAuthStateStore())

	_, err := svc.BeginAuth(context.Background(), BeginRequest{
		DefinitionID:   "missing",
		InstallationID: "install-1",
		CredentialRef:  keymakerTestCredentialRef,
	})
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}

func TestService_BeginAuthNoAuthRegistration(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "no-auth", Slug: "no-auth"},
		Connections: []types.ConnectionRegistration{
			{
				CredentialRef:  keymakerTestCredentialRef,
				CredentialRefs: []types.CredentialRef{keymakerTestCredentialRef},
			},
		},
	}

	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, (&fakeInstallationWriter{}).PersistAuthResult, matchingInstallationResolver("i", "no-auth").ResolveInstallation, NewInMemoryAuthStateStore())

	_, err := svc.BeginAuth(context.Background(), BeginRequest{
		DefinitionID:   "no-auth",
		InstallationID: "i",
		CredentialRef:  keymakerTestCredentialRef,
	})
	if !errors.Is(err, ErrDefinitionAuthRequired) {
		t.Fatalf("expected ErrDefinitionAuthRequired, got %v", err)
	}
}

func TestService_BeginAuthGeneratesSessionStateToken(t *testing.T) {
	t.Parallel()

	def := authTestDefinition("d1", "d1", &fakeAuthFlow{
		startResult: types.AuthStartResult{URL: "https://example.com"},
	})

	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, (&fakeInstallationWriter{}).PersistAuthResult, matchingInstallationResolver("i1", "d1").ResolveInstallation, NewInMemoryAuthStateStore())

	begin, err := svc.BeginAuth(context.Background(), BeginRequest{
		DefinitionID:   "d1",
		InstallationID: "i1",
		CredentialRef:  keymakerTestCredentialRef,
	})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	if begin.State == "" {
		t.Fatalf("expected generated state token")
	}
}

func TestService_BeginAuthInstallationDefinitionMismatch(t *testing.T) {
	t.Parallel()

	def := authTestDefinition("d1", "d1", &fakeAuthFlow{
		startResult: types.AuthStartResult{URL: "https://example.com"},
	})

	svc := NewService(
		(&fakeDefinitionResolver{def: def}).Definition,
		(&fakeInstallationWriter{}).PersistAuthResult,
		matchingInstallationResolver("i1", "other-definition").ResolveInstallation,
		NewInMemoryAuthStateStore(),
	)

	_, err := svc.BeginAuth(context.Background(), BeginRequest{
		DefinitionID:   "d1",
		InstallationID: "i1",
		CredentialRef:  keymakerTestCredentialRef,
	})
	if !errors.Is(err, ErrInstallationDefinitionMismatch) {
		t.Fatalf("expected ErrInstallationDefinitionMismatch, got %v", err)
	}
}

func TestService_CompleteAuthExpired(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	clock := func() time.Time { return now }

	def := authTestDefinition("slack", "slack", &fakeAuthFlow{
		startResult:    types.AuthStartResult{URL: "https://slack.com/oauth"},
		completeResult: types.AuthCompleteResult{},
	})

	store := NewInMemoryAuthStateStore()
	store.now = clock

	svc := NewService(
		(&fakeDefinitionResolver{def: def}).Definition,
		(&fakeInstallationWriter{}).PersistAuthResult,
		matchingInstallationResolver("install-2", "slack").ResolveInstallation,
		store,
	)

	begin, err := svc.BeginAuth(ctx, BeginRequest{
		DefinitionID:   "slack",
		InstallationID: "install-2",
		CredentialRef:  keymakerTestCredentialRef,
	})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	now = now.Add(defaultSessionTTL + time.Minute)

	_, err = svc.CompleteAuth(ctx, CompleteRequest{State: begin.State})
	if !errors.Is(err, ErrAuthStateExpired) {
		t.Fatalf("expected ErrAuthStateExpired, got %v", err)
	}
}

func TestService_CompleteAuthStateTokenRequired(t *testing.T) {
	t.Parallel()

	svc := NewService((&fakeDefinitionResolver{}).Definition, (&fakeInstallationWriter{}).PersistAuthResult, (&fakeInstallationResolver{}).ResolveInstallation, NewInMemoryAuthStateStore())

	_, err := svc.CompleteAuth(context.Background(), CompleteRequest{})
	if !errors.Is(err, ErrAuthStateTokenRequired) {
		t.Fatalf("expected ErrAuthStateTokenRequired, got %v", err)
	}
}

func TestService_CompleteAuthSaveError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	def := authTestDefinition("okta", "okta", &fakeAuthFlow{
		startResult:    types.AuthStartResult{URL: "https://okta.com"},
		completeResult: types.AuthCompleteResult{Credential: types.CredentialSet{}},
	})

	writer := &fakeInstallationWriter{err: errors.New("db unavailable")}
	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, writer.PersistAuthResult, matchingInstallationResolver("install-3", "okta").ResolveInstallation, NewInMemoryAuthStateStore())

	begin, err := svc.BeginAuth(ctx, BeginRequest{
		DefinitionID:   "okta",
		InstallationID: "install-3",
		CredentialRef:  keymakerTestCredentialRef,
	})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	_, err = svc.CompleteAuth(ctx, CompleteRequest{State: begin.State})
	if err == nil || !strings.Contains(err.Error(), "keymaker: auth complete hook") {
		t.Fatalf("expected wrapped save error, got %v", err)
	}
}

func TestService_CallbackStatePassedToComplete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var receivedCallbackState json.RawMessage

	def := authTestDefinition("az", "az", &fakeAuthFlow{
		startResult: types.AuthStartResult{
			URL:   "https://login.microsoftonline.com",
			State: json.RawMessage(`{"nonce":"n1","tenant":"t1"}`),
		},
		completeResult: types.AuthCompleteResult{
			Credential: types.CredentialSet{Data: json.RawMessage(`{"accessToken":"az-token"}`)},
		},
		onComplete: func(state json.RawMessage) {
			receivedCallbackState = state
		},
	})

	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, (&fakeInstallationWriter{}).PersistAuthResult, matchingInstallationResolver("i1", "az").ResolveInstallation, NewInMemoryAuthStateStore())

	begin, err := svc.BeginAuth(ctx, BeginRequest{DefinitionID: "az", InstallationID: "i1", CredentialRef: keymakerTestCredentialRef})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	_, err = svc.CompleteAuth(ctx, CompleteRequest{State: begin.State, Callback: types.AuthCallbackInput{}})
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

	def := authTestDefinition("az", "az", &fakeAuthFlow{
		startResult:    types.AuthStartResult{URL: "https://login.microsoftonline.com"},
		completeResult: types.AuthCompleteResult{Credential: types.CredentialSet{Data: json.RawMessage(`{"accessToken":"az-token"}`)}},
	})

	svc := NewService((&fakeDefinitionResolver{def: def}).Definition, (&fakeInstallationWriter{}).PersistAuthResult, installations.ResolveInstallation, NewInMemoryAuthStateStore())

	begin, err := svc.BeginAuth(ctx, BeginRequest{DefinitionID: "az", InstallationID: "i1", CredentialRef: keymakerTestCredentialRef})
	if err != nil {
		t.Fatalf("BeginAuth error: %v", err)
	}

	installations.installation.DefinitionID = "changed"

	_, err = svc.CompleteAuth(ctx, CompleteRequest{State: begin.State, Callback: types.AuthCallbackInput{}})
	if !errors.Is(err, ErrInstallationDefinitionMismatch) {
		t.Fatalf("expected ErrInstallationDefinitionMismatch, got %v", err)
	}
}

// fakeAuthFlow holds configurable auth behavior for testing
type fakeAuthFlow struct {
	startResult    types.AuthStartResult
	startErr       error
	completeResult types.AuthCompleteResult
	completeErr    error
	onComplete     func(state json.RawMessage)
}

func authTestDefinition(definitionID string, slug string, flow *fakeAuthFlow) types.Definition {
	return types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: definitionID, Slug: slug},
		Connections: []types.ConnectionRegistration{
			{
				CredentialRef:  keymakerTestCredentialRef,
				CredentialRefs: []types.CredentialRef{keymakerTestCredentialRef},
				Auth:           flow.registration(keymakerTestCredentialRef),
			},
		},
	}
}

// registration returns an AuthRegistration wired to the fake's configured behavior
func (f *fakeAuthFlow) registration(credentialRef types.CredentialRef) *types.AuthRegistration {
	return &types.AuthRegistration{
		CredentialRef: credentialRef,
		Start: func(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
			return f.startResult, f.startErr
		},
		Complete: func(_ context.Context, state json.RawMessage, _ types.AuthCallbackInput) (types.AuthCompleteResult, error) {
			if f.onComplete != nil {
				f.onComplete(state)
			}

			return f.completeResult, f.completeErr
		},
		Refresh: func(_ context.Context, credential types.CredentialSet) (types.CredentialSet, error) {
			return credential, nil
		},
		TokenView: func(_ context.Context, _ types.CredentialSet) (*types.TokenView, error) {
			return nil, nil
		},
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
	credentialRef  types.CredentialRef
	definitionID   string
	result         types.AuthCompleteResult
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

func (f *fakeInstallationWriter) PersistAuthResult(_ context.Context, installationID string, credentialRef types.CredentialRef, definition types.Definition, result types.AuthCompleteResult) error {
	f.saves = append(f.saves, installationSave{
		installationID: installationID,
		credentialRef:  credentialRef,
		definitionID:   definition.ID,
		result:         result,
	})
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
