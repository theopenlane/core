package cloudflare

import (
	"context"
	"encoding/json"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/user"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
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
	return providerkit.WithClient(cloudflareClient, h.Run)
}

// Run executes the Cloudflare token verification
func (HealthCheck) Run(ctx context.Context, c *cf.Client) (json.RawMessage, error) {
	res, err := c.User.Tokens.Verify(ctx)
	if err != nil {
		return nil, ErrTokenVerificationFailed
	}

	if res.Status != user.TokenVerifyResponseStatusActive {
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
