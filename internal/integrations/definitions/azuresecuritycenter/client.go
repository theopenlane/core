package azuresecuritycenter

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// azureSubscriptionScopePrefix is the ARM resource scope prefix for subscription-level requests
	azureSubscriptionScopePrefix = "subscriptions/"
	// defaultAzureScope is the Azure Resource Manager OAuth scope for management API access
	defaultAzureScope = "https://management.azure.com/.default"
	// azureTokenURLTemplate is the Microsoft identity platform token endpoint template, parameterized by tenant ID
	azureTokenURLTemplate = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
)

// Client builds Azure Defender for Cloud clients for one installation
type Client struct{}

// staticAzureCredential implements azcore.TokenCredential for a pre-obtained bearer token
type staticAzureCredential struct {
	// token is the pre-obtained bearer token returned on every GetToken call
	token string
}

// azurePricingsClient wraps a PricingsClient with its subscription scope
type azurePricingsClient struct {
	// client is the underlying Azure Defender for Cloud pricings client
	client *armsecurity.PricingsClient
	// scope is the ARM subscription scope string used for pricings queries
	scope string
}

// GetToken satisfies azcore.TokenCredential for a pre-obtained bearer token
func (s staticAzureCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

// Build constructs an Azure Defender for Cloud client via client credentials flow
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	var cred CredentialSchema
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, ErrCredentialInvalid
	}

	switch {
	case cred.TenantID == "":
		return nil, ErrTenantIDMissing
	case cred.ClientID == "":
		return nil, ErrClientIDMissing
	case cred.ClientSecret == "":
		return nil, ErrClientSecretMissing
	case cred.SubscriptionID == "":
		return nil, ErrSubscriptionIDMissing
	}

	tokenURL := fmt.Sprintf(azureTokenURLTemplate, cred.TenantID)
	cfg := clientcredentials.Config{
		ClientID:     cred.ClientID,
		ClientSecret: cred.ClientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{defaultAzureScope},
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	token, err := cfg.Token(ctx)
	if err != nil {
		return nil, ErrTokenExchangeFailed
	}

	pricingsClient, err := armsecurity.NewPricingsClient(staticAzureCredential{token: token.AccessToken}, nil)
	if err != nil {
		return nil, ErrPricingsClientBuildFailed
	}

	return &azurePricingsClient{
		client: pricingsClient,
		scope:  fmt.Sprintf("%s%s", azureSubscriptionScopePrefix, cred.SubscriptionID),
	}, nil
}

// FromAny casts a registered client instance to the Azure pricing client type
func (Client) FromAny(value any) (*azurePricingsClient, error) {
	c, ok := value.(*azurePricingsClient)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}
