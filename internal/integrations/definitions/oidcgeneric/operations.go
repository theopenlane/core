package oidcgeneric

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

type oidcHealthDetails struct {
	Subject string `json:"sub,omitempty"`
	Issuer  string `json:"iss,omitempty"`
}

type oidcClaimsDetails struct {
	Claims map[string]any `json:"claims"`
}

// buildOIDCClient builds an authenticated HTTP client for OIDC userinfo calls
func buildOIDCClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	token := req.Credential.OAuthAccessToken
	if token == "" {
		return nil, ErrOAuthTokenMissing
	}

	return providerkit.NewAuthenticatedClient("", token, nil), nil
}

// runHealthOperation calls the OIDC userinfo endpoint to validate the token
func runHealthOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	if credential.Claims != nil {
		sub, _ := credential.Claims["sub"].(string)
		iss, _ := credential.Claims["iss"].(string)
		return jsonx.ToRawMessage(oidcHealthDetails{Subject: sub, Issuer: iss})
	}

	if c.BaseURL == "" {
		return jsonx.ToRawMessage(oidcHealthDetails{})
	}

	var resp map[string]any
	if err := c.GetJSON(ctx, c.BaseURL, &resp); err != nil {
		return nil, fmt.Errorf("oidcgeneric: userinfo call failed: %w", err)
	}

	sub, _ := resp["sub"].(string)
	iss, _ := resp["iss"].(string)

	return jsonx.ToRawMessage(oidcHealthDetails{Subject: sub, Issuer: iss})
}

// runClaimsInspectOperation returns stored OIDC claims for inspection
func runClaimsInspectOperation(_ context.Context, _ *generated.Integration, credential types.CredentialSet, _ any, _ json.RawMessage) (json.RawMessage, error) {
	if credential.Claims == nil {
		return jsonx.ToRawMessage(oidcClaimsDetails{Claims: map[string]any{}})
	}

	return jsonx.ToRawMessage(oidcClaimsDetails{Claims: credential.Claims})
}
