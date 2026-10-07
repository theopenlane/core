package onedrive

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// graphScope is the default scope used for Microsoft Graph client requests
const graphScope = "https://graph.microsoft.com/.default"

// clientBuilder returns the OneDrive Graph client builder bound to the operator config
func clientBuilder(cfg Config) func(context.Context, types.ConnectionRequest[oneDriveCred]) (*DriveClient, error) {
	return func(ctx context.Context, req types.ConnectionRequest[oneDriveCred]) (*DriveClient, error) {
		cred := req.Credential

		if cred.AccessToken == "" {
			return nil, ErrOAuthTokenMissing
		}

		base := fmt.Sprintf(microsoftAuthBaseURL, "common")

		oauthCfg := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  base + "/authorize",
				TokenURL: base + "/token",
			},
			Scopes: []string{
				"https://graph.microsoft.com/Files.Read",
				"https://graph.microsoft.com/User.Read",
				"offline_access",
			},
		}

		ts := oauthCfg.TokenSource(context.Background(), providerkit.OAuthToken(cred.AccessToken, cred.RefreshToken, cred.Expiry))

		tokenCred := &oauthTokenCredential{ts: ts}

		authProvider, err := kiotaauth.NewAzureIdentityAuthenticationProviderWithScopes(tokenCred, []string{graphScope})
		if err != nil {
			return nil, ErrClientBuildFailed
		}

		adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error building onedrive client")
			return nil, ErrClientBuildFailed
		}

		return &DriveClient{Graph: msgraphsdk.NewGraphServiceClient(adapter), TS: ts, Cfg: cfg}, nil
	}
}

// oauthTokenCredential adapts an oauth2.TokenSource to azcore.TokenCredential
type oauthTokenCredential struct {
	ts oauth2.TokenSource
}

// GetToken obtains the current (or refreshed) access token from the underlying oauth2.TokenSource
func (c *oauthTokenCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	tok, err := c.ts.Token()
	if err != nil {
		return azcore.AccessToken{}, err
	}

	expiry := tok.Expiry
	if expiry.IsZero() {
		expiry = time.Now().Add(time.Hour)
	}

	return azcore.AccessToken{Token: tok.AccessToken, ExpiresOn: expiry}, nil
}
