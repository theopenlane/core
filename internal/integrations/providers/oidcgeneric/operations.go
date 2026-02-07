package oidcgeneric

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

// oidcOperations returns OIDC operation descriptors
func oidcOperations(userInfoURL string) []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(types.OperationName("health.default"), "Call the configured userinfo endpoint (when available) to validate the OIDC token.", ClientOIDCAPI, runOIDCHealth(userInfoURL)),
		{
			Name:        types.OperationName("claims.inspect"),
			Kind:        types.OperationKindScanSettings,
			Description: "Expose stored ID token claims for downstream checks.",
			Run:         runOIDCClaims,
		},
	}
}

// runOIDCHealth builds a health check function for OIDC tokens
func runOIDCHealth(userInfoURL string) types.OperationFunc {
	return func(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
		client, token, err := auth.ClientAndOAuthToken(input, TypeOIDCGeneric)
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
		if err := auth.GetJSONWithClient(ctx, client, userInfoURL, token, nil, &resp); err != nil {
			return operations.OperationFailure("OIDC userinfo call failed", err), err
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

// runOIDCClaims returns stored OIDC claims for inspection
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
