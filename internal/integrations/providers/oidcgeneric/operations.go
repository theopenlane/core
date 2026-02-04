package oidcgeneric

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

// oidcOperations handles oidc operations
func oidcOperations(userInfoURL string) []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        types.OperationName("health.default"),
			Kind:        types.OperationKindHealth,
			Description: "Call the configured userinfo endpoint (when available) to validate the OIDC token.",
			Client:      ClientOIDCAPI,
			Run:         runOIDCHealth(userInfoURL),
		},
		{
			Name:        types.OperationName("claims.inspect"),
			Kind:        types.OperationKindScanSettings,
			Description: "Expose stored ID token claims for downstream checks.",
			Run:         runOIDCClaims,
		},
	}
}

// runOIDCHealth runs oidc health
func runOIDCHealth(userInfoURL string) types.OperationFunc {
	return func(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
		client, token, err := helpers.ClientAndOAuthToken(input, TypeOIDCGeneric)
		if err != nil {
			return types.OperationResult{}, err
		}

		if userInfoURL == "" {
			return types.OperationResult{
				Status:  types.OperationStatusOK,
				Summary: "OIDC token present (no userinfo endpoint configured)",
			}, nil
		}

		var resp map[string]any
		if err := helpers.GetJSONWithClient(ctx, client, userInfoURL, token, nil, &resp); err != nil {
			return helpers.OperationFailure("OIDC userinfo call failed", err), err
		}

		summary := "OIDC userinfo call succeeded"
		if subject, ok := resp["sub"].(string); ok {
			summary = fmt.Sprintf("OIDC userinfo succeeded for %s", subject)
		}

		return types.OperationResult{
			Status:  types.OperationStatusOK,
			Summary: summary,
			Details: resp,
		}, nil
	}
}

// runOIDCClaims runs oidc claims
func runOIDCClaims(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
	claims := input.Credential.Claims
	if claims == nil {
		return types.OperationResult{
			Status:  types.OperationStatusUnknown,
			Summary: "No OIDC claims stored",
		}, nil
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: "OIDC claims available",
		Details: map[string]any{"claims": claims},
	}, nil
}
