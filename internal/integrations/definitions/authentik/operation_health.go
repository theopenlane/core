package authentik

import (
	"context"
	"encoding/json"

	authentikSDK "goauthentik.io/api/v3"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an Authentik health check
type HealthCheck struct {
	// PK is the Authentik user identifier
	PK int32 `json:"pk"`
	// Username is the Authentik service account username
	Username string `json:"username"`
	// Email is the Authentik service account email
	Email string `json:"email"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(authentikClient, h.Run)
}

// Run executes the Authentik health check
func (HealthCheck) Run(ctx context.Context, c *authentikSDK.APIClient) (json.RawMessage, error) {
	me, resp, err := c.CoreApi.CoreUsersMeRetrieve(ctx).Execute()
	if resp != nil {
		_ = resp.Body.Close()
	}

	if err != nil {
		return nil, ErrHealthCheckFailed
	}

	user := me.GetUser()

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	return providerkit.EncodeResult(HealthCheck{
		PK:       user.GetPk(),
		Username: user.GetUsername(),
		Email:    email,
	}, ErrResultEncode)
}
