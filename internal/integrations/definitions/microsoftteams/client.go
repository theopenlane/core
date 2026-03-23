package microsoftteams

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// teamsGraphScope is the Microsoft Graph scope used for Teams operations
const teamsGraphScope = "https://graph.microsoft.com/.default"

// Client builds Microsoft Graph service clients for one Teams installation
type Client struct{}

// Build constructs the Microsoft Graph service client from the installation OAuth access token
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	var tc teamsCred
	if err := jsonx.UnmarshalIfPresent(req.Credential.Data, &tc); err != nil {
		return nil, ErrCredentialDecode
	}

	if tc.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	cred := &staticTokenCredential{
		token:  tc.AccessToken,
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
	// token is the pre-obtained bearer token used for Microsoft Graph requests
	token string
	// expiry is the timestamp at which the bearer token expires
	expiry time.Time
}

// GetToken returns the static bearer token
func (s *staticTokenCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token, ExpiresOn: s.expiry}, nil
}
