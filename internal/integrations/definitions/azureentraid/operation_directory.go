package azureentraid

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// DirectoryInspect collects Azure Entra ID tenant metadata
type DirectoryInspect struct {
	// ID is the organization identifier
	ID string `json:"id"`
	// DisplayName is the organization display name
	DisplayName string `json:"displayName"`
	// VerifiedDomains is the list of verified domains for the tenant
	VerifiedDomains any `json:"verifiedDomains"`
}

type graphOrganization struct {
	ID              string `json:"id"`
	DisplayName     string `json:"displayName"`
	TenantID        string `json:"tenantId"`
	VerifiedDomains []any  `json:"verifiedDomains"`
}

type graphOrganizationResponse struct {
	Value []graphOrganization `json:"value"`
}

// Handle adapts directory inspect to the generic operation registration boundary
func (d DirectoryInspect) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := EntraClient.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return d.Run(ctx, c)
	}
}

// Run collects Azure Entra ID tenant metadata via Microsoft Graph
func (DirectoryInspect) Run(ctx context.Context, c *providerkit.AuthenticatedClient) (json.RawMessage, error) {
	org, err := fetchOrganization(ctx, c)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(DirectoryInspect{
		ID:              org.ID,
		DisplayName:     org.DisplayName,
		VerifiedDomains: org.VerifiedDomains,
	}, ErrResultEncode)
}

// fetchOrganization retrieves the first organization entry from Microsoft Graph
func fetchOrganization(ctx context.Context, client *providerkit.AuthenticatedClient) (graphOrganization, error) {
	var resp graphOrganizationResponse
	if err := client.GetJSON(ctx, "organization?$select=id,displayName,tenantId,verifiedDomains&$top=1", &resp); err != nil {
		return graphOrganization{}, ErrOrganizationLookupFailed
	}

	if len(resp.Value) == 0 {
		return graphOrganization{}, ErrNoOrganizations
	}

	return resp.Value[0], nil
}
