package azureentraid

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an Azure Entra ID health check
type HealthCheck struct {
	// ID is the organization identifier
	ID string `json:"id"`
	// TenantID is the Azure tenant identifier
	TenantID string `json:"tenantId"`
	// DisplayName is the organization display name
	DisplayName string `json:"displayName"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := EntraClient.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, c)
	}
}

// Run executes the Azure Entra ID health check
func (HealthCheck) Run(ctx context.Context, c *providerkit.AuthenticatedClient) (json.RawMessage, error) {
	org, err := fetchOrganization(ctx, c)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(HealthCheck{
		ID:          org.ID,
		TenantID:    org.TenantID,
		DisplayName: org.DisplayName,
	}, ErrResultEncode)
}
