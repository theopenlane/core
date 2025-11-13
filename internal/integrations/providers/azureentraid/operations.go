package azureentraid

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/integrations/providers/helpers"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	azureEntraHealthOp types.OperationName = "health.default"
	azureEntraTenantOp types.OperationName = "directory.inspect"
)

func azureOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        azureEntraHealthOp,
			Kind:        types.OperationKindHealth,
			Description: "Call Microsoft Graph /organization to verify tenant access.",
			Run:         runAzureEntraHealth,
		},
		{
			Name:        azureEntraTenantOp,
			Kind:        types.OperationKindScanSettings,
			Description: "Collect basic tenant metadata via Microsoft Graph.",
			Run:         runAzureEntraTenantInspect,
		},
	}
}

func runAzureEntraHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeAzureEntraID))
	if err != nil {
		return types.OperationResult{}, err
	}

	org, err := fetchOrganization(ctx, token)
	if err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Graph organization lookup failed",
			Details: map[string]any{"error": err.Error()},
		}, err
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

func runAzureEntraTenantInspect(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeAzureEntraID))
	if err != nil {
		return types.OperationResult{}, err
	}

	org, err := fetchOrganization(ctx, token)
	if err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Graph organization lookup failed",
			Details: map[string]any{"error": err.Error()},
		}, err
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
	ID              string        `json:"id"`
	DisplayName     string        `json:"displayName"`
	TenantID        string        `json:"tenantId"`
	VerifiedDomains []interface{} `json:"verifiedDomains"`
}

type graphOrganizationResponse struct {
	Value []graphOrganization `json:"value"`
}

func fetchOrganization(ctx context.Context, token string) (graphOrganization, error) {
	endpoint := "https://graph.microsoft.com/v1.0/organization?$select=id,displayName,tenantId,verifiedDomains&$top=1"
	var resp graphOrganizationResponse
	if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &resp); err != nil {
		return graphOrganization{}, err
	}

	if len(resp.Value) == 0 {
		return graphOrganization{}, fmt.Errorf("graph returned zero organizations")
	}

	return resp.Value[0], nil
}
