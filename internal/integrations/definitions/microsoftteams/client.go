package microsoftteams

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/internal/integrations/types"
)

// teamsGraphScope is the Microsoft Graph scope used for Teams operations
const teamsGraphScope = "https://graph.microsoft.com/.default"

// Client builds Microsoft Graph service clients for one Teams installation
type Client struct{}

// Build constructs the Microsoft Graph service client from the installation OAuth access token
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	if req.Credential.OAuthAccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	cred := &staticTokenCredential{
		token:  req.Credential.OAuthAccessToken,
		expiry: time.Now().Add(time.Hour),
	}

	authProvider, err := kiotaauth.NewAzureIdentityAuthenticationProviderWithScopes(cred, []string{teamsGraphScope})
	if err != nil {
		return nil, ErrOAuthTokenMissing
	}

	adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		return nil, ErrOAuthTokenMissing
	}

	return msgraphsdk.NewGraphServiceClient(adapter), nil
}

// staticTokenCredential wraps a pre-obtained bearer token as an azcore.TokenCredential
type staticTokenCredential struct {
	token  string
	expiry time.Time
}

// GetToken returns the static bearer token
func (s *staticTokenCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token, ExpiresOn: s.expiry}, nil
}
