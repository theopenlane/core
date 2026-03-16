package microsoftteams

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ProfileResponse is the raw Microsoft Graph /me response shape
type ProfileResponse struct {
	// ID is the Microsoft Graph user identifier
	ID string `json:"id"`
	// DisplayName is the user's display name
	DisplayName string `json:"displayName"`
	// Mail is the user's email address
	Mail string `json:"mail"`
}

// HealthCheck holds the result of a Microsoft Teams health check
type HealthCheck struct {
	// ID is the Microsoft Graph user identifier
	ID string `json:"id"`
	// Mail is the user's email address
	Mail string `json:"mail"`
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

// Run executes the Microsoft Teams health check
func (HealthCheck) Run(ctx context.Context, c *providerkit.AuthenticatedClient) (json.RawMessage, error) {
	var profile ProfileResponse
	if err := c.GetJSON(ctx, "me", &profile); err != nil {
		return nil, ErrProfileLookupFailed
	}

	return providerkit.EncodeResult(HealthCheck{
		ID:   profile.ID,
		Mail: profile.Mail,
	}, ErrResultEncode)
}
