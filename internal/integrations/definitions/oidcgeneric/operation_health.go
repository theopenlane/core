package oidcgeneric

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an OIDC health check
type HealthCheck struct {
	// Subject is the OIDC subject claim
	Subject string `json:"sub,omitempty"`
	// Issuer is the OIDC issuer claim
	Issuer string `json:"iss,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := OIDCClient.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, request.Credential, c)
	}
}

// Run executes the OIDC health check
func (HealthCheck) Run(ctx context.Context, credential types.CredentialSet, c *providerkit.AuthenticatedClient) (json.RawMessage, error) {
	if credential.Claims != nil {
		sub, _ := credential.Claims["sub"].(string)
		iss, _ := credential.Claims["iss"].(string)

		return providerkit.EncodeResult(HealthCheck{Subject: sub, Issuer: iss}, ErrResultEncode)
	}

	if c.BaseURL == "" {
		return providerkit.EncodeResult(HealthCheck{}, ErrResultEncode)
	}

	var resp map[string]any
	if err := c.GetJSON(ctx, c.BaseURL, &resp); err != nil {
		return nil, ErrUserinfoCallFailed
	}

	sub, _ := resp["sub"].(string)
	iss, _ := resp["iss"].(string)

	return providerkit.EncodeResult(HealthCheck{Subject: sub, Issuer: iss}, ErrResultEncode)
}
