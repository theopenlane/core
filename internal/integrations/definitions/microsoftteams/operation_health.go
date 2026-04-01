package microsoftteams

import (
	"context"
	"encoding/json"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of a Microsoft Teams health check
type HealthCheck struct {
	// ID is the Microsoft Graph user identifier
	ID string `json:"id"`
	// Mail is the user's email address
	Mail string `json:"mail"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(teamsClient, h.Run)
}

// Run executes the Microsoft Teams health check via Microsoft Graph
func (HealthCheck) Run(ctx context.Context, c *msgraphsdk.GraphServiceClient) (json.RawMessage, error) {
	me, err := c.Me().Get(ctx, nil)
	if err != nil {
		return nil, ErrProfileLookupFailed
	}

	return providerkit.EncodeResult(HealthCheck{
		ID:   lo.FromPtr(me.GetId()),
		Mail: lo.FromPtr(me.GetMail()),
	}, ErrResultEncode)
}
