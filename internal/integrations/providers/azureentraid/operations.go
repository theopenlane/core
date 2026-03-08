package azureentraid

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
)

const azureEntraGraphBaseURL = "https://graph.microsoft.com/v1.0/"

const (
	azureEntraHealthOp types.OperationName = types.OperationHealthDefault
	azureEntraTenantOp types.OperationName = "directory.inspect"
)

type azureEntraHealthDetails struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	DisplayName string `json:"displayName"`
}

type azureEntraTenantDetails struct {
	ID              string      `json:"id"`
	DisplayName     string      `json:"displayName"`
	VerifiedDomains interface{} `json:"verifiedDomains"`
}

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
	client, err := auth.ResolveAuthenticatedClient(input, auth.OAuthTokenFromPayload, azureEntraGraphBaseURL, nil)
	if err != nil {
		return types.OperationResult{}, err
	}

	org, err := fetchOrganization(ctx, client)
	if err != nil {
		return operations.OperationFailure("Graph organization lookup failed", err, nil)
	}

	summary := fmt.Sprintf("Tenant %s reachable", org.DisplayName)
	return operations.OperationSuccess(summary, azureEntraHealthDetails{
		ID:          org.ID,
		TenantID:    org.TenantID,
		DisplayName: org.DisplayName,
	}), nil
}

// runAzureEntraTenantInspect collects tenant metadata from Microsoft Graph
func runAzureEntraTenantInspect(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := auth.ResolveAuthenticatedClient(input, auth.OAuthTokenFromPayload, azureEntraGraphBaseURL, nil)
	if err != nil {
		return types.OperationResult{}, err
	}

	org, err := fetchOrganization(ctx, client)
	if err != nil {
		return operations.OperationFailure("Graph organization lookup failed", err, nil)
	}

	return operations.OperationSuccess(fmt.Sprintf("Collected metadata for tenant %s", org.DisplayName), azureEntraTenantDetails{
		ID:              org.ID,
		DisplayName:     org.DisplayName,
		VerifiedDomains: org.VerifiedDomains,
	}), nil
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
func fetchOrganization(ctx context.Context, client *auth.AuthenticatedClient) (graphOrganization, error) {
	var resp graphOrganizationResponse
	if err := client.GetJSON(ctx, "organization?$select=id,displayName,tenantId,verifiedDomains&$top=1", &resp); err != nil {
		return graphOrganization{}, err
	}

	if len(resp.Value) == 0 {
		return graphOrganization{}, ErrNoOrganizations
	}

	return resp.Value[0], nil
}
