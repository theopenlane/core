package azureentraid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const azureEntraGraphBaseURL = "https://graph.microsoft.com/v1.0/"

type azureEntraHealthDetails struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	DisplayName string `json:"displayName"`
}

type azureEntraTenantDetails struct {
	ID              string `json:"id"`
	DisplayName     string `json:"displayName"`
	VerifiedDomains any    `json:"verifiedDomains"`
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

// buildGraphClient builds the Microsoft Graph API client for one installation
func buildGraphClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	token := req.Credential.OAuthAccessToken
	if token == "" {
		return nil, ErrOAuthTokenMissing
	}

	return providerkit.NewAuthenticatedClient(azureEntraGraphBaseURL, token, nil), nil
}

// runHealthOperation calls /organization to verify tenant access
func runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	org, err := fetchOrganization(ctx, c)
	if err != nil {
		return nil, err
	}

	return jsonx.ToRawMessage(azureEntraHealthDetails{
		ID:          org.ID,
		TenantID:    org.TenantID,
		DisplayName: org.DisplayName,
	})
}

// runDirectoryInspectOperation collects tenant metadata via Microsoft Graph
func runDirectoryInspectOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	org, err := fetchOrganization(ctx, c)
	if err != nil {
		return nil, err
	}

	return jsonx.ToRawMessage(azureEntraTenantDetails{
		ID:              org.ID,
		DisplayName:     org.DisplayName,
		VerifiedDomains: org.VerifiedDomains,
	})
}

// fetchOrganization retrieves the first organization entry from Microsoft Graph
func fetchOrganization(ctx context.Context, client *providerkit.AuthenticatedClient) (graphOrganization, error) {
	var resp graphOrganizationResponse
	if err := client.GetJSON(ctx, "organization?$select=id,displayName,tenantId,verifiedDomains&$top=1", &resp); err != nil {
		return graphOrganization{}, fmt.Errorf("azureentraid: organization lookup failed: %w", err)
	}

	if len(resp.Value) == 0 {
		return graphOrganization{}, errors.New("azureentraid: no organizations returned")
	}

	return resp.Value[0], nil
}
