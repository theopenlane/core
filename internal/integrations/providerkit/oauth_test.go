package providerkit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestCompleteOAuthFlowMissingCallbackStateFallsThroughExchange(t *testing.T) {
	t.Parallel()

	cfg := OAuthFlowConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		AuthURL:      "https://example.com/oauth/authorize",
		TokenURL:     "https://example.com/oauth/token",
		RedirectURL:  "https://app.example.com/callback",
	}

	startState, err := json.Marshal(oauthStartState{State: "expected-state"})
	if err != nil {
		t.Fatalf("Marshal(startState) error = %v", err)
	}

	callbackInput, err := json.Marshal(OAuthCallbackInput{Code: "code-123"})
	if err != nil {
		t.Fatalf("Marshal(callbackInput) error = %v", err)
	}

	_, err = CompleteOAuthFlow(context.Background(), cfg, startState, callbackInput)
	if !errors.Is(err, ErrOAuthCodeExchange) {
		t.Fatalf("CompleteOAuthFlow() error = %v, want %v", err, ErrOAuthCodeExchange)
	}
}
