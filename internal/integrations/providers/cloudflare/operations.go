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

type cloudflareHealthDetails struct {
	IssuedOn  string `json:"issuedOn,omitempty"`
	ExpiresOn string `json:"expiresOn,omitempty"`
}

type cloudflareAPIError struct {
	// Message is the error message returned by the API
	Message string `json:"message"`
}

type cloudflareTokenVerificationResponse struct {
	// Success indicates whether the API call succeeded
	Success bool `json:"success"`
	// Result holds token metadata returned by the API
	Result struct {
		// IssuedOn is the token issued timestamp
		IssuedOn string `json:"issued_on"`
		// ExpiresOn is the token expiration timestamp
		ExpiresOn string `json:"expires_on"`
	} `json:"result"`
	// Errors lists any API errors returned by the verification call
	Errors []cloudflareAPIError `json:"errors"`
}

type cloudflareVerificationFailureDetails struct {
	Errors []cloudflareAPIError `json:"errors,omitempty"`
}

// cloudflareOperations returns Cloudflare operation descriptors
func cloudflareOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(cloudflareHealthOp, "Verify Cloudflare API token via /user/tokens/verify.", ClientCloudflareAPI, runCloudflareHealth),
	}
}

// runCloudflareHealth validates Cloudflare credentials via token verification
func runCloudflareHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndToken(input, auth.APITokenFromPayload)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp cloudflareTokenVerificationResponse

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	endpoint := "https://api.cloudflare.com/client/v4/user/tokens/verify"
	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, headers, &resp); err != nil {
		return operations.OperationFailure("Cloudflare token verification failed", err, nil)
	}

	if !resp.Success {
		return operations.OperationFailure("Cloudflare token verification returned errors", ErrTokenVerificationFailed, cloudflareVerificationFailureDetails{
			Errors: resp.Errors,
		})
	}

	return operations.OperationSuccess("Cloudflare token verified", cloudflareHealthDetails{
		IssuedOn:  resp.Result.IssuedOn,
		ExpiresOn: resp.Result.ExpiresOn,
	}), nil
}
