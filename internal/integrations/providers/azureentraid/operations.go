package azureentraid

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	azureEntraHealthOp types.OperationName = "health.default"
	azureEntraTenantOp types.OperationName = "directory.inspect"
)

// azureOperations returns the Azure Entra ID operations supported by this provider.
func azureOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(azureEntraHealthOp, "Call Microsoft Graph /organization to verify tenant access.", ClientAzureEntraAPI, runAzureEntraHealth),
		{
			Name:        azureEntraTenantOp,
			Kind:        types.OperationKindScanSettings,
			Description: "Collect basic tenant metadata via Microsoft Graph.",
			Client:      ClientAzureEntraAPI,
			Run:         runAzureEntraTenantInspect,
		},
	}
}

// runAzureEntraHealth performs a basic tenant reachability check
func runAzureEntraHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndToken(input, auth.OAuthTokenFromPayload)
	if err != nil {
		return types.OperationResult{}, err
	}

	org, err := fetchOrganization(ctx, token, client)
	if err != nil {
		return operations.OperationFailure("Graph organization lookup failed", err, nil)
	}

	summary := fmt.Sprintf("Tenant %s reachable", org.DisplayName)
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: map[string]any{
			"id":          org.ID,
			"tenantId":    org.TenantID,
			"displayName": org.DisplayName,
		},
	}, nil
}

// runAzureEntraTenantInspect collects tenant metadata from Microsoft Graph
func runAzureEntraTenantInspect(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndToken(input, auth.OAuthTokenFromPayload)
	if err != nil {
		return types.OperationResult{}, err
	}

	org, err := fetchOrganization(ctx, token, client)
	if err != nil {
		return operations.OperationFailure("Graph organization lookup failed", err, nil)
	}

	details := map[string]any{
		"id":              org.ID,
		"displayName":     org.DisplayName,
		"verifiedDomains": org.VerifiedDomains,
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected metadata for tenant %s", org.DisplayName),
		Details: details,
	}, nil
}

type graphOrganization struct {
	// ID is the organization identifier
	ID string `json:"id"`
	// DisplayName is the organization display name
	DisplayName string `json:"displayName"`
	// TenantID is the Azure AD tenant identifier
	TenantID string `json:"tenantId"`
	// VerifiedDomains lists verified domains for the tenant
	VerifiedDomains []interface{} `json:"verifiedDomains"`
}

type graphOrganizationResponse struct {
	// Value holds organization entries returned by Graph
	Value []graphOrganization `json:"value"`
}

// fetchOrganization retrieves the first organization entry from Microsoft Graph.
func fetchOrganization(ctx context.Context, token string, client *auth.AuthenticatedClient) (graphOrganization, error) {
	endpoint := "https://graph.microsoft.com/v1.0/organization?$select=id,displayName,tenantId,verifiedDomains&$top=1"
	var resp graphOrganizationResponse
	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
		return graphOrganization{}, err
	}

	if len(resp.Value) == 0 {
		return graphOrganization{}, ErrNoOrganizations
	}

	return resp.Value[0], nil
}
