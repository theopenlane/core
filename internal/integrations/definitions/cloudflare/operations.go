package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/user"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// HealthCheck holds the result of a Cloudflare health check
type HealthCheck struct {
	// Status is the token verification status returned by Cloudflare
	Status string `json:"status,omitempty"`
	// ExpiresOn is the token expiry time if set
	ExpiresOn string `json:"expiresOn,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, c)
	}
}

// Run executes the Cloudflare token verification
func (HealthCheck) Run(ctx context.Context, c *cf.Client) (json.RawMessage, error) {
	res, err := c.User.Tokens.Verify(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: token verification failed: %w", err)
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

	return jsonx.ToRawMessage(details)
}
