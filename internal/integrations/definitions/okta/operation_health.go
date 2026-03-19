package okta

import (
	"context"
	"encoding/json"

	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an Okta health check
type HealthCheck struct {
	// ID is the Okta user identifier
	ID string `json:"id"`
	// Login is the Okta user login
	Login string `json:"login"`
	// Email is the Okta user email
	Email string `json:"email"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := OktaClient.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, c)
	}
}

// Run executes the Okta health check
func (HealthCheck) Run(ctx context.Context, c *oktagosdk.APIClient) (json.RawMessage, error) {
	user, _, err := c.UserAPI.GetUser(ctx, "me").Execute()
	if err != nil {
		return nil, ErrUserLookupFailed
	}

	profile := user.GetProfile()
	login := profile.GetLogin()

	return providerkit.EncodeResult(HealthCheck{
		ID:    user.GetId(),
		Login: login,
		Email: profile.GetEmail(),
	}, ErrResultEncode)
}
