package cloudflare

import (
	"context"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	cloudflareHealthOp types.OperationName = "health.default"
)

// cloudflareOperations returns Cloudflare operation descriptors
func cloudflareOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(cloudflareHealthOp, "Verify Cloudflare API token via /user/tokens/verify.", ClientCloudflareAPI, runCloudflareHealth),
	}
}

// runCloudflareHealth validates Cloudflare credentials via token verification
func runCloudflareHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndAPIToken(input, TypeCloudflare)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp struct {
		// Success indicates whether the API call succeeded
		Success bool `json:"success"`
		// Result holds token metadata returned by the API
		Result  struct {
			// IssuedOn is the token issued timestamp
			IssuedOn  string `json:"issued_on"`
			// ExpiresOn is the token expiration timestamp
			ExpiresOn string `json:"expires_on"`
		} `json:"result"`
		// Errors lists any API errors returned by the verification call
		Errors []struct {
			// Message is the error message returned by the API
			Message string `json:"message"`
		} `json:"errors"`
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	endpoint := "https://api.cloudflare.com/client/v4/user/tokens/verify"
	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, headers, &resp); err != nil {
		return operations.OperationFailure("Cloudflare token verification failed", err), err
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
