package cloudflare

import (
	"context"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	cloudflareHealthOp types.OperationName = "health.default"
)

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

func runCloudflareHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
	token, err := helpers.APITokenFromPayload(input.Credential, string(TypeCloudflare))
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
	if client != nil {
		if err := client.GetJSON(ctx, endpoint, &resp); err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "Cloudflare token verification failed",
				Details: map[string]any{"error": err.Error()},
			}, err
		}
	} else if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, headers, &resp); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Cloudflare token verification failed",
			Details: map[string]any{"error": err.Error()},
		}, err
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
