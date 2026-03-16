package googleworkspace

import (
	"context"
	"encoding/json"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const healthMaxResults = int64(1)

// HealthCheck holds the result of a Google Workspace health check
type HealthCheck struct {
	// UserCount is the number of users returned by the health probe
	UserCount int `json:"userCount"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		svc, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, svc)
	}
}

// Run executes the health check using the Google Admin SDK
func (HealthCheck) Run(ctx context.Context, svc *admin.Service) (json.RawMessage, error) {
	resp, err := svc.Users.List().
		Customer("my_customer").
		MaxResults(healthMaxResults).
		Projection("basic").
		ViewType("admin_view").
		Fields(googleapi.Field("users(id),nextPageToken")).
		Context(ctx).
		Do()
	if err != nil {
		return nil, ErrHealthCheckFailed
	}

	return providerkit.EncodeResult(HealthCheck{UserCount: len(resp.Users)}, ErrResultEncode)
}
