package providerkit

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"
)

var testOAuthCfg = OAuthFlowConfig{
	ClientID:     "client-id",
	ClientSecret: "client-secret",
	AuthURL:      "https://example.com/oauth/authorize",
	TokenURL:     "https://example.com/oauth/token",
	RedirectURL:  "https://app.example.com/callback",
}

func TestStartOAuthFlowReturnsURLAndState(t *testing.T) {
	t.Parallel()

	result, err := StartOAuthFlow(context.Background(), testOAuthCfg)
	if err != nil {
		t.Fatalf("StartOAuthFlow() error = %v", err)
	}

	if result.URL == "" {
		t.Fatal("StartOAuthFlow() URL is empty")
	}

	parsed, err := url.Parse(result.URL)
	if err != nil {
		t.Fatalf("StartOAuthFlow() URL is not valid: %v", err)
	}

	if !strings.HasPrefix(parsed.String(), testOAuthCfg.AuthURL) {
		t.Fatalf("StartOAuthFlow() URL = %v, want prefix %v", result.URL, testOAuthCfg.AuthURL)
	}

	var state oauthStartState
	if err := json.Unmarshal(result.State, &state); err != nil {
		t.Fatalf("StartOAuthFlow() State is not valid JSON: %v", err)
	}

	if state.State == "" {
		t.Fatal("StartOAuthFlow() State.state is empty")
	}
}

func TestStartOAuthFlowStateIsUnique(t *testing.T) {
	t.Parallel()

	r1, err := StartOAuthFlow(context.Background(), testOAuthCfg)
	if err != nil {
		t.Fatalf("StartOAuthFlow() first call error = %v", err)
	}

	r2, err := StartOAuthFlow(context.Background(), testOAuthCfg)
	if err != nil {
		t.Fatalf("StartOAuthFlow() second call error = %v", err)
	}

	var s1, s2 oauthStartState
	_ = json.Unmarshal(r1.State, &s1)
	_ = json.Unmarshal(r2.State, &s2)

	if s1.State == s2.State {
		t.Fatal("StartOAuthFlow() produced the same state on two calls")
	}
}

func TestCompleteOAuthFlowStateMismatchReturnsError(t *testing.T) {
	t.Parallel()

	startState, err := json.Marshal(oauthStartState{State: "expected-state"})
	if err != nil {
		t.Fatalf("Marshal(startState) error = %v", err)
	}

	callbackInput, err := json.Marshal(OAuthCallbackInput{Code: "code-123", State: "different-state"})
	if err != nil {
		t.Fatalf("Marshal(callbackInput) error = %v", err)
	}

	_, err = CompleteOAuthFlow(context.Background(), testOAuthCfg, startState, callbackInput)
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Fatalf("CompleteOAuthFlow() error = %v, want %v", err, ErrOAuthStateMismatch)
	}
}

func TestCompleteOAuthFlowMissingCodeReturnsError(t *testing.T) {
	t.Parallel()

	startState, err := json.Marshal(oauthStartState{State: "some-state"})
	if err != nil {
		t.Fatalf("Marshal(startState) error = %v", err)
	}

	callbackInput, err := json.Marshal(OAuthCallbackInput{State: "some-state"})
	if err != nil {
		t.Fatalf("Marshal(callbackInput) error = %v", err)
	}

	_, err = CompleteOAuthFlow(context.Background(), testOAuthCfg, startState, callbackInput)
	if !errors.Is(err, ErrOAuthCodeMissing) {
		t.Fatalf("CompleteOAuthFlow() error = %v, want %v", err, ErrOAuthCodeMissing)
	}
}

func TestCompleteOAuthFlowMissingCallbackStateFallsThroughExchange(t *testing.T) {
	t.Parallel()

	startState, err := json.Marshal(oauthStartState{State: "expected-state"})
	if err != nil {
		t.Fatalf("Marshal(startState) error = %v", err)
	}

	callbackInput, err := json.Marshal(OAuthCallbackInput{Code: "code-123"})
	if err != nil {
		t.Fatalf("Marshal(callbackInput) error = %v", err)
	}

	_, err = CompleteOAuthFlow(context.Background(), testOAuthCfg, startState, callbackInput)
	if !errors.Is(err, ErrOAuthCodeExchange) {
		t.Fatalf("CompleteOAuthFlow() error = %v, want %v", err, ErrOAuthCodeExchange)
	}
}
