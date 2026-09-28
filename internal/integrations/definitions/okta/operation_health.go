package okta

import (
	"context"
	"encoding/json"

	oktagosdk "github.com/okta/okta-sdk-golang/v6/okta"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
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

// checkHealth executes the Okta health check
func checkHealth(ctx context.Context, _ types.OperationRequest, c *oktagosdk.APIClient) (json.RawMessage, error) {
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
