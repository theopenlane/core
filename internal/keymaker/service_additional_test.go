package keymaker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations"
)

func TestNewServiceValidatesDependencies(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil, &fakeKeystore{}, NewMemorySessionStore(), ServiceOptions{})
	if !errors.Is(err, integrations.ErrProviderRegistryUninitialized) {
		t.Fatalf("expected ErrProviderRegistryUninitialized, got %v", err)
	}

	resolver := fakeResolver{provider: &fakeProvider{providerType: types.ProviderType("acme")}}

	_, err = NewService(resolver, nil, NewMemorySessionStore(), ServiceOptions{})
	if !errors.Is(err, integrations.ErrKeystoreRequired) {
		t.Fatalf("expected ErrKeystoreRequired, got %v", err)
	}

	_, err = NewService(resolver, &fakeKeystore{}, nil, ServiceOptions{})
	if !errors.Is(err, integrations.ErrSessionStoreRequired) {
		t.Fatalf("expected ErrSessionStoreRequired, got %v", err)
	}
}

func TestBeginAuthorizationClonesRequestDataAndSetsTTL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Date(2024, 11, 24, 10, 0, 0, 0, time.UTC)
	providerType := types.ProviderType("acme")
	provider := &fakeProvider{
		providerType: providerType,
		state:        "state-1",
		authURL:      "https://example.com/auth",
		payload: types.CredentialPayload{
			Provider: providerType,
			Kind:     types.CredentialKindOAuthToken,
		},
	}

	store := &recordingSessionStore{}
	service, err := NewService(
		fakeResolver{provider: provider},
		&fakeKeystore{},
		store,
		ServiceOptions{
			SessionTTL: 10 * time.Minute,
			Now:        func() time.Time { return now },
		},
	)
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	scopes := []string{"repo"}
	metadata := map[string]any{"env": "prod"}
	labels := map[string]string{"color": "blue"}

	if _, err := service.BeginAuthorization(ctx, BeginRequest{
		OrgID:          "org-1",
		IntegrationID:  "int-1",
		Provider:       providerType,
		Scopes:         scopes,
		Metadata:       metadata,
		LabelOverrides: labels,
	}); err != nil {
		t.Fatalf("BeginAuthorization error: %v", err)
	}

	saved := store.saved
	if saved.State != "state-1" {
		t.Fatalf("expected state to be persisted")
	}
	if !saved.CreatedAt.Equal(now) {
		t.Fatalf("expected creation time %s, got %s", now, saved.CreatedAt)
	}
	if !saved.ExpiresAt.Equal(now.Add(10 * time.Minute)) {
		t.Fatalf("expected expiry 10m later, got %s", saved.ExpiresAt)
	}

	// Mutate the original inputs and ensure the saved session did not change.
	scopes[0] = "mutated"
	metadata["env"] = "dev"
	labels["color"] = "red"

	if saved.Scopes[0] != "repo" {
		t.Fatalf("expected scopes to be cloned, got %v", saved.Scopes)
	}
	if saved.Metadata["env"] != "prod" {
		t.Fatalf("expected metadata to be cloned, got %v", saved.Metadata)
	}
	if saved.LabelOverrides["color"] != "blue" {
		t.Fatalf("expected label overrides to be cloned, got %v", saved.LabelOverrides)
	}
}

func TestBeginAuthorizationSaveError(t *testing.T) {
	t.Parallel()

	providerType := types.ProviderType("acme")
	service, err := NewService(
		fakeResolver{provider: &fakeProvider{providerType: providerType, state: "state-xyz"}},
		&fakeKeystore{},
		&recordingSessionStore{saveErr: errors.New("store down")},
		ServiceOptions{},
	)
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	_, err = service.BeginAuthorization(context.Background(), BeginRequest{
		OrgID:         "org-1",
		IntegrationID: "int-1",
		Provider:      providerType,
	})
	if err == nil || !strings.Contains(err.Error(), "keymaker: save auth session") {
		t.Fatalf("expected wrapped save error, got %v", err)
	}
}

func TestBeginAuthorizationRequiresProviderState(t *testing.T) {
	t.Parallel()

	providerType := types.ProviderType("acme")
	service, err := NewService(
		fakeResolver{provider: &fakeProvider{providerType: providerType, state: "   "}},
		&fakeKeystore{},
		&recordingSessionStore{},
		ServiceOptions{},
	)
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	_, err = service.BeginAuthorization(context.Background(), BeginRequest{
		OrgID:         "org-1",
		IntegrationID: "int-1",
		Provider:      providerType,
	})
	if !errors.Is(err, integrations.ErrStateRequired) {
		t.Fatalf("expected ErrStateRequired, got %v", err)
	}
}

func TestCompleteAuthorizationValidatesInputs(t *testing.T) {
	t.Parallel()

	service, err := NewService(
		fakeResolver{provider: &fakeProvider{providerType: types.ProviderType("acme")}},
		&fakeKeystore{},
		NewMemorySessionStore(),
		ServiceOptions{},
	)
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	_, err = service.CompleteAuthorization(context.Background(), CompleteRequest{State: "", Code: "code"})
	if !errors.Is(err, integrations.ErrStateRequired) {
		t.Fatalf("expected ErrStateRequired, got %v", err)
	}

	_, err = service.CompleteAuthorization(context.Background(), CompleteRequest{State: "state", Code: ""})
	if !errors.Is(err, integrations.ErrAuthorizationCodeRequired) {
		t.Fatalf("expected ErrAuthorizationCodeRequired, got %v", err)
	}
}

func TestCompleteAuthorizationSessionErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	providerType := types.ProviderType("acme")
	now := time.Now()
	baseActivation := ActivationSession{
		State:          "state-123",
		Provider:       providerType,
		OrgID:          "org-1",
		IntegrationID:  "int-1",
		CreatedAt:      now.Add(-time.Minute),
		ExpiresAt:      now.Add(time.Minute),
		LabelOverrides: map[string]string{},
		Metadata:       map[string]any{},
	}

	tests := []struct {
		name      string
		store     *recordingSessionStore
		keystore  *fakeKeystore
		wantError error
		checkErr  func(error) bool
	}{
		{
			name:      "take error",
			store:     &recordingSessionStore{takeErr: integrations.ErrAuthorizationStateNotFound},
			keystore:  &fakeKeystore{},
			wantError: integrations.ErrAuthorizationStateNotFound,
		},
		{
			name: "missing auth session",
			store: &recordingSessionStore{
				takeResponse: baseActivation,
			},
			keystore:  &fakeKeystore{},
			wantError: integrations.ErrAuthSessionInvalid,
		},
		{
			name: "finish error",
			store: &recordingSessionStore{
				takeResponse: func() ActivationSession {
					act := baseActivation
					act.AuthSession = &fakeAuthSession{
						provider:  providerType,
						state:     act.State,
						finishErr: errors.New("finish failed"),
					}
					return act
				}(),
			},
			keystore: &fakeKeystore{},
			checkErr: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "keymaker: finish auth")
			},
		},
		{
			name: "save credential error",
			store: &recordingSessionStore{
				takeResponse: func() ActivationSession {
					act := baseActivation
					act.AuthSession = &fakeAuthSession{
						provider: providerType,
						state:    act.State,
						payload: types.CredentialPayload{
							Provider: providerType,
						},
					}
					return act
				}(),
			},
			keystore: &fakeKeystore{err: errors.New("db unavailable")},
			checkErr: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "keymaker: save credential")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service, err := NewService(
				fakeResolver{provider: &fakeProvider{providerType: providerType}},
				tc.keystore,
				tc.store,
				ServiceOptions{
					Now: func() time.Time { return now },
				},
			)
			if err != nil {
				t.Fatalf("NewService error: %v", err)
			}

			_, err = service.CompleteAuthorization(ctx, CompleteRequest{
				State: baseActivation.State,
				Code:  "code-1",
			})

			switch {
			case tc.checkErr != nil:
				if !tc.checkErr(err) {
					t.Fatalf("unexpected error: %v", err)
				}
			default:
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("expected %v, got %v", tc.wantError, err)
				}
			}
		})
	}
}

type recordingSessionStore struct {
	saved        ActivationSession
	saveErr      error
	takeErr      error
	takeResponse ActivationSession
}

func (s *recordingSessionStore) Save(session ActivationSession) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = session
	return nil
}

func (s *recordingSessionStore) Take(_ string) (ActivationSession, error) {
	if s.takeErr != nil {
		return ActivationSession{}, s.takeErr
	}
	if s.takeResponse.State != "" || s.takeResponse.AuthSession != nil {
		return s.takeResponse, nil
	}
	return s.saved, nil
}
