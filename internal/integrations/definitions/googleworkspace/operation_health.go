package googleworkspace

import (
	"context"
	"encoding/json"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// healthMaxResults is the maximum number of users to fetch during a health probe
const healthMaxResults = int64(1)

// HealthCheck holds the result of a Google Workspace health check
type HealthCheck struct {
	// UserCount is the number of users returned by the health probe
	UserCount int `json:"userCount"`
}

// checkHealth executes the health check using the Google Admin SDK
func checkHealth(ctx context.Context, _ types.OperationRequest, svc *admin.Service) (json.RawMessage, error) {
	resp, err := svc.Users.List().
		Customer(defaultCustomerID).
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
