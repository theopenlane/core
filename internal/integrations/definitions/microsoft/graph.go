package microsoft

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"golang.org/x/oauth2"
)

// NewGraphClient builds a Microsoft Graph service client for the configured endpoints
func (c AppConfig) NewGraphClient(cred azcore.TokenCredential) (*msgraphsdk.GraphServiceClient, error) {
	authProvider, err := kiotaauth.NewAzureIdentityAuthenticationProviderWithScopes(cred, []string{GraphScope})
	if err != nil {
		return nil, err
	}

	if c.endpoints == nil {
		adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
		if err != nil {
			return nil, err
		}

		return msgraphsdk.NewGraphServiceClient(adapter), nil
	}

	adapter, err := msgraphsdk.NewGraphRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(authProvider, nil, nil, c.endpoints.httpClient)
	if err != nil {
		return nil, err
	}

	adapter.SetBaseUrl(c.endpoints.graphURL)

	return msgraphsdk.NewGraphServiceClient(adapter), nil
}

// NewDelegatedGraphClient builds a Graph client from stored delegated OAuth material, returning the refreshing token source for plain HTTP calls
func (c AppConfig) NewDelegatedGraphClient(ctx context.Context, tenant string, scopes []string, token *oauth2.Token) (*msgraphsdk.GraphServiceClient, oauth2.TokenSource, error) {
	ts := c.delegatedTokenSource(ctx, tenant, scopes, token)

	graph, err := c.NewGraphClient(NewOAuthTokenCredential(ts))
	if err != nil {
		return nil, nil, err
	}

	return graph, ts, nil
}
