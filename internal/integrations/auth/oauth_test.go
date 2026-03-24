package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/theopenlane/core/internal/integrations/types"
)

var testOAuthCfg = OAuthConfig{
	ClientID:     "client-id",
	ClientSecret: "client-secret",
	AuthURL:      "https://example.com/oauth/authorize",
	TokenURL:     "https://example.com/oauth/token",
	RedirectURL:  "https://app.example.com/callback",
}

func TestStartOAuthReturnsURLAndState(t *testing.T) {
	t.Parallel()

	result, err := StartOAuth(context.Background(), testOAuthCfg)
	if err != nil {
		t.Fatalf("StartOAuth() error = %v", err)
	}

	if result.URL == "" {
		t.Fatal("StartOAuth() URL is empty")
	}

	parsed, err := url.Parse(result.URL)
	if err != nil {
		t.Fatalf("StartOAuth() URL is not valid: %v", err)
	}

	if !strings.HasPrefix(parsed.String(), testOAuthCfg.AuthURL) {
		t.Fatalf("StartOAuth() URL = %v, want prefix %v", result.URL, testOAuthCfg.AuthURL)
	}

	var state oauthStartState
	if err := json.Unmarshal(result.State, &state); err != nil {
		t.Fatalf("StartOAuth() State is not valid JSON: %v", err)
	}

	if state.State == "" {
		t.Fatal("StartOAuth() State.state is empty")
	}
}

func TestStartOAuthStateIsUnique(t *testing.T) {
	t.Parallel()

	r1, err := StartOAuth(context.Background(), testOAuthCfg)
	if err != nil {
		t.Fatalf("StartOAuth() first call error = %v", err)
	}

	r2, err := StartOAuth(context.Background(), testOAuthCfg)
	if err != nil {
		t.Fatalf("StartOAuth() second call error = %v", err)
	}

	var s1, s2 oauthStartState
	_ = json.Unmarshal(r1.State, &s1)
	_ = json.Unmarshal(r2.State, &s2)

	if s1.State == s2.State {
		t.Fatal("StartOAuth() produced the same state on two calls")
	}
}

func TestCompleteOAuthStateMismatchReturnsError(t *testing.T) {
	t.Parallel()

	startState, err := json.Marshal(oauthStartState{State: "expected-state"})
	if err != nil {
		t.Fatalf("Marshal(startState) error = %v", err)
	}

	_, err = CompleteOAuth(context.Background(), testOAuthCfg, startState, types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "code", Values: []string{"code-123"}},
			{Name: "state", Values: []string{"different-state"}},
		},
	})
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Fatalf("CompleteOAuth() error = %v, want %v", err, ErrOAuthStateMismatch)
	}
}

func TestCompleteOAuthMissingCodeReturnsError(t *testing.T) {
	t.Parallel()

	startState, err := json.Marshal(oauthStartState{State: "some-state"})
	if err != nil {
		t.Fatalf("Marshal(startState) error = %v", err)
	}

	_, err = CompleteOAuth(context.Background(), testOAuthCfg, startState, types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "state", Values: []string{"some-state"}},
		},
	})
	if !errors.Is(err, ErrOAuthCodeMissing) {
		t.Fatalf("CompleteOAuth() error = %v, want %v", err, ErrOAuthCodeMissing)
	}
}

func TestCompleteOAuthMissingCallbackStateReturnsError(t *testing.T) {
	t.Parallel()

	startState, err := json.Marshal(oauthStartState{State: "expected-state"})
	if err != nil {
		t.Fatalf("Marshal(startState) error = %v", err)
	}

	_, err = CompleteOAuth(context.Background(), testOAuthCfg, startState, types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "code", Values: []string{"code-123"}},
		},
	})
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Fatalf("CompleteOAuth() error = %v, want %v", err, ErrOAuthStateMismatch)
	}
}

func TestCompleteOAuthInvalidStartStateReturnsError(t *testing.T) {
	t.Parallel()

	_, err := CompleteOAuth(context.Background(), testOAuthCfg, json.RawMessage(`{invalid`), types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "code", Values: []string{"code-123"}},
			{Name: "state", Values: []string{"some-state"}},
		},
	})
	if !errors.Is(err, ErrOAuthStateInvalid) {
		t.Fatalf("CompleteOAuth() error = %v, want %v", err, ErrOAuthStateInvalid)
	}
}
