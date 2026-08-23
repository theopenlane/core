//go:build test

package integrations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// OAuthAccessToken is the access token minted by the OAuth completion fixture
	OAuthAccessToken = "test-access-token"
	// OAuthRefreshToken is the refresh token minted by the OAuth completion fixture
	OAuthRefreshToken = "test-refresh-token"
)

// oauthCallbackPayload carries the OAuth state and code
type oauthCallbackPayload struct {
	State string `json:"state"`
	Code  string `json:"code,omitempty"`
}

// oauthStart mints an authorize URL and opaque state
func oauthStart(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	oauthState := ulids.New().String()

	stateBytes, err := json.Marshal(oauthCallbackPayload{State: oauthState})
	if err != nil {
		return types.AuthStartResult{}, err
	}

	return types.AuthStartResult{
		URL:   fmt.Sprintf("https://example.com/oauth/authorize?state=%s", oauthState),
		State: stateBytes,
	}, nil
}

// oauthComplete validates the callback and mints the OAuth credential
func oauthComplete(_ context.Context, callbackState json.RawMessage, input types.AuthCallbackInput) (types.AuthCompleteResult, error) {
	var stored oauthCallbackPayload
	if err := json.Unmarshal(callbackState, &stored); err != nil {
		return types.AuthCompleteResult{}, err
	}

	if input.First("code") == "" {
		return types.AuthCompleteResult{}, ErrOAuthCodeMissing
	}

	if stored.State != input.First("state") {
		return types.AuthCompleteResult{}, ErrOAuthStateMismatch
	}

	tokenData, err := json.Marshal(oauthTokenCred{
		AccessToken:  OAuthAccessToken,
		RefreshToken: OAuthRefreshToken,
	})
	if err != nil {
		return types.AuthCompleteResult{}, err
	}

	return types.AuthCompleteResult{
		Credential: types.CredentialSet{Data: tokenData},
	}, nil
}
