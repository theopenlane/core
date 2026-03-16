package oidcgeneric

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// HealthCheck holds the result of an OIDC health check
type HealthCheck struct {
	// Subject is the OIDC subject claim
	Subject string `json:"sub,omitempty"`
	// Issuer is the OIDC issuer claim
	Issuer string `json:"iss,omitempty"`
}

// ClaimsInspect returns stored OIDC claims for inspection
type ClaimsInspect struct {
	// Claims is the full set of OIDC claims
	Claims map[string]any `json:"claims"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
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

		return jsonx.ToRawMessage(HealthCheck{Subject: sub, Issuer: iss})
	}

	if c.BaseURL == "" {
		return jsonx.ToRawMessage(HealthCheck{})
	}

	var resp map[string]any
	if err := c.GetJSON(ctx, c.BaseURL, &resp); err != nil {
		return nil, fmt.Errorf("oidcgeneric: userinfo call failed: %w", err)
	}

	sub, _ := resp["sub"].(string)
	iss, _ := resp["iss"].(string)

	return jsonx.ToRawMessage(HealthCheck{Subject: sub, Issuer: iss})
}

// Handle adapts claims inspection to the generic operation registration boundary
func (ci ClaimsInspect) Handle() types.OperationHandler {
	return func(_ context.Context, request types.OperationRequest) (json.RawMessage, error) {
		return ci.Run(request.Credential)
	}
}

// Run returns the stored OIDC claims
func (ClaimsInspect) Run(credential types.CredentialSet) (json.RawMessage, error) {
	if credential.Claims == nil {
		return jsonx.ToRawMessage(ClaimsInspect{Claims: map[string]any{}})
	}

	return jsonx.ToRawMessage(ClaimsInspect{Claims: credential.Claims})
}
