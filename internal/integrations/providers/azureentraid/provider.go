package azureentraid

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeAzureEntraID identifies the Azure Entra ID provider
const TypeAzureEntraID = types.ProviderType("azureentraid")

const (
	// ClientAzureEntraAPI identifies the Microsoft Graph API client for Entra ID
	ClientAzureEntraAPI types.ClientName = "api"
)

const azureEntraGraphBaseURL = "https://graph.microsoft.com/v1.0/"

const (
	azureEntraHealthOp types.OperationName = types.OperationHealthDefault
	azureEntraTenantOp types.OperationName = "directory.inspect"
)

// azureEntraCredentialsSchema is the JSON Schema for Azure Entra ID tenant credentials.
var azureEntraCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["tenantId"],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly identifier for this Azure tenant."},"tenantId":{"type":"string","title":"Tenant ID","description":"Azure Active Directory tenant ID used for Graph API calls."},"appId":{"type":"string","title":"Application (Client) ID","description":"Optional override if you provide a tenant-specific Azure app registration."},"appSecret":{"type":"string","title":"Application Secret","description":"Optional secret tied to the provided Application ID."}}}`)

// Builder returns the Azure Entra ID provider builder with the supplied operator config applied.
func Builder(cfg Config) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeAzureEntraID,
		SpecFunc:     azureEntraIDSpec,
		BuildFunc: func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
			if s.OAuth != nil && cfg.ClientID != "" {
				s.OAuth.ClientID = cfg.ClientID
				s.OAuth.ClientSecret = cfg.ClientSecret
			}

			return oauth.New(s, oauth.WithOperations(azureOperations()), oauth.WithClientDescriptors(azureEntraClientDescriptors()))
		},
	}
}

// azureEntraIDSpec returns the static provider specification for the Azure Entra ID provider.
func azureEntraIDSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:             "azureentraid",
		DisplayName:      "Azure Entra ID",
		Category:         "identity",
		AuthType:         types.AuthKindOAuth2,
		AuthStartPath:    "/v1/integrations/oauth/start",
		AuthCallbackPath: "/v1/integrations/oauth/callback",
		Active:           lo.ToPtr(false),
		Visible:          lo.ToPtr(true),
		LogoURL:          "",
		DocsURL:          "https://docs.theopenlane.io/docs/platform/integrations/azure_entra_id/overview",
		OAuth: &spec.OAuthSpec{
			AuthURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:    "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			Scopes:      []string{"https://graph.microsoft.com/.default", "offline_access"},
			RedirectURI: "https://api.theopenlane.io/v1/integrations/oauth/callback",
		},
		UserInfo: &spec.UserInfoSpec{
			URL:       "https://graph.microsoft.com/v1.0/me",
			Method:    "GET",
			AuthStyle: "Bearer",
			IDPath:    "id",
			EmailPath: "mail",
			LoginPath: "displayName",
		},
		Persistence: &spec.PersistenceSpec{
			StoreRefreshToken: true,
		},
		Labels: map[string]string{
			"vendor":  "microsoft",
			"product": "entra-id",
		},
		CredentialsSchema: azureEntraCredentialsSchema,
		Description:       "Connect to Microsoft Graph to validate tenant access and inspect Azure Entra ID organization metadata.",
	}
}

// azureEntraClientDescriptors returns the client descriptors published by Azure Entra ID
func azureEntraClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeAzureEntraID, ClientAzureEntraAPI, "Microsoft Graph API client", providerkit.TokenClientBuilder(providerkit.OAuthTokenFromCredential, nil))
}

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

// azureOperations returns the Azure Entra ID operations supported by this provider
func azureOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(azureEntraHealthOp, "Call Microsoft Graph /organization to verify tenant access.", ClientAzureEntraAPI, runAzureEntraHealth),
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
	client, err := providerkit.ResolveAuthenticatedClient(input, providerkit.OAuthTokenFromCredential, azureEntraGraphBaseURL, nil)
	if err != nil {
		return types.OperationResult{}, err
	}

	org, err := fetchOrganization(ctx, client)
	if err != nil {
		return providerkit.OperationFailure("Graph organization lookup failed", err, nil)
	}

	summary := fmt.Sprintf("Tenant %s reachable", org.DisplayName)
	return providerkit.OperationSuccess(summary, azureEntraHealthDetails{
		ID:          org.ID,
		TenantID:    org.TenantID,
		DisplayName: org.DisplayName,
	}), nil
}

// runAzureEntraTenantInspect collects tenant metadata from Microsoft Graph
func runAzureEntraTenantInspect(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := providerkit.ResolveAuthenticatedClient(input, providerkit.OAuthTokenFromCredential, azureEntraGraphBaseURL, nil)
	if err != nil {
		return types.OperationResult{}, err
	}

	org, err := fetchOrganization(ctx, client)
	if err != nil {
		return providerkit.OperationFailure("Graph organization lookup failed", err, nil)
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Collected metadata for tenant %s", org.DisplayName), azureEntraTenantDetails{
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

// fetchOrganization retrieves the first organization entry from Microsoft Graph
func fetchOrganization(ctx context.Context, client *providerkit.AuthenticatedClient) (graphOrganization, error) {
	var resp graphOrganizationResponse
	if err := client.GetJSON(ctx, "organization?$select=id,displayName,tenantId,verifiedDomains&$top=1", &resp); err != nil {
		return graphOrganization{}, err
	}

	if len(resp.Value) == 0 {
		return graphOrganization{}, ErrNoOrganizations
	}

	return resp.Value[0], nil
}
