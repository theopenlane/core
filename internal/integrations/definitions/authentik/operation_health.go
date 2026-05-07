package authentik

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an Authentik health check
type HealthCheck struct {
	// PK is the Authentik user identifier
	PK int `json:"pk"`
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
func (HealthCheck) Run(ctx context.Context, c *Client) (json.RawMessage, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, authentikMeEndpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, ErrRequestBuildFailed
	}

	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, ErrHealthCheckFailed
	}

	defer resp.Body.Close()

	var me MeResponse
	if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
		return nil, ErrHealthCheckFailed
	}

	return providerkit.EncodeResult(HealthCheck{
		PK:       me.User.PK,
		Username: me.User.Username,
		Email:    me.User.Email,
	}, ErrResultEncode)
}
