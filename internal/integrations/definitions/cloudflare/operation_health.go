package cloudflare

import (
	"context"
	"encoding/json"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// HealthCheck holds the result of a Cloudflare health check
type HealthCheck struct {
	// Status is the token verification status returned by Cloudflare
	Status string `json:"status,omitempty"`
	// ExpiresOn is the token expiry time if set
	ExpiresOn string `json:"expiresOn,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClientRequest(cloudflareClient, func(ctx context.Context, request types.OperationRequest, client *cf.Client) (json.RawMessage, error) {
		return h.Run(ctx, request.Credentials, client)
	})

}

// Run executes the Cloudflare token verification
func (HealthCheck) Run(ctx context.Context, credentials types.CredentialBindings, c *cf.Client) (json.RawMessage, error) {
	meta, err := resolveCredential(credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("gcpscc: error attempting to resolve credentials")
		return nil, err
	}

	res, err := c.Accounts.Tokens.Verify(ctx, accounts.TokenVerifyParams{
		AccountID: cf.F(meta.AccountID),
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("githubapp: healthcheck failed on token verification")
		return nil, ErrTokenVerificationFailed
	}

	if res.Status != accounts.TokenVerifyResponseStatusActive {
		return nil, ErrTokenNotActive
	}

	details := HealthCheck{
		Status: string(res.Status),
	}

	if !res.ExpiresOn.IsZero() {
		details.ExpiresOn = res.ExpiresOn.String()
	}

	return providerkit.EncodeResult(details, ErrResultEncode)
}
