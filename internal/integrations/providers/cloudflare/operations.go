package cloudflare

import (
	"context"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/user"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	cloudflareHealthOp types.OperationName = "health.default"
)

type cloudflareHealthDetails struct {
	Status    string `json:"status,omitempty"`
	ExpiresOn string `json:"expiresOn,omitempty"`
}

// cloudflareOperations returns Cloudflare operation descriptors
func cloudflareOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(cloudflareHealthOp, "Verify Cloudflare API token via /user/tokens/verify.", ClientCloudflareAPI, runCloudflareHealth),
	}
}

// resolveCloudflareClient returns a pooled Cloudflare client or builds one from the credential payload.
func resolveCloudflareClient(input types.OperationInput) (*cf.Client, error) {
	if c, ok := types.ClientInstanceAs[*cf.Client](input.Client); ok {
		return c, nil
	}

	token, err := auth.APITokenFromPayload(input.Credential)
	if err != nil {
		return nil, err
	}

	return cf.NewClient(option.WithAPIToken(token)), nil
}

// runCloudflareHealth validates Cloudflare credentials via token verification
func runCloudflareHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveCloudflareClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	res, err := client.User.Tokens.Verify(ctx)
	if err != nil {
		return operations.OperationFailure("Cloudflare token verification failed", err, nil)
	}

	if res.Status != user.TokenVerifyResponseStatusActive {
		return operations.OperationFailure("Cloudflare token is not active", ErrTokenVerificationFailed, cloudflareHealthDetails{
			Status: string(res.Status),
		})
	}

	details := cloudflareHealthDetails{
		Status: string(res.Status),
	}

	if !res.ExpiresOn.IsZero() {
		details.ExpiresOn = res.ExpiresOn.String()
	}

	return operations.OperationSuccess(fmt.Sprintf("Cloudflare token verified, status: %s", res.Status), details), nil
}
