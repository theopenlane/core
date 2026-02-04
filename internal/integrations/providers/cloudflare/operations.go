package cloudflare

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	cloudflareHealthOp types.OperationName = "health.default"
)

// cloudflareOperations handles cloudflare operations
func cloudflareOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        cloudflareHealthOp,
			Kind:        types.OperationKindHealth,
			Description: "Verify Cloudflare API token via /user/tokens/verify.",
			Client:      ClientCloudflareAPI,
			Run:         runCloudflareHealth,
		},
	}
}

// runCloudflareHealth runs cloudflare health
func runCloudflareHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := helpers.ClientAndAPIToken(input, TypeCloudflare)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp struct {
		Success bool `json:"success"`
		Result  struct {
			IssuedOn  string `json:"issued_on"`
			ExpiresOn string `json:"expires_on"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	endpoint := "https://api.cloudflare.com/client/v4/user/tokens/verify"
	if err := helpers.GetJSONWithClient(ctx, client, endpoint, token, headers, &resp); err != nil {
		return helpers.OperationFailure("Cloudflare token verification failed", err), err
	}

	if !resp.Success {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Cloudflare token verification returned errors",
			Details: map[string]any{"errors": resp.Errors},
		}, ErrTokenVerificationFailed
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: "Cloudflare token verified",
		Details: map[string]any{
			"issuedOn":  resp.Result.IssuedOn,
			"expiresOn": resp.Result.ExpiresOn,
		},
	}, nil
}
