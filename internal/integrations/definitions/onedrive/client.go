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

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// graphScope is the default scope used for Microsoft Graph client requests
const graphScope = "https://graph.microsoft.com/.default"

// Client builds OneDrive Graph clients for one installation
type Client struct {
	// cfg is the operator-level OneDrive configuration
	cfg Config
}

// Build constructs a DriveClient from the installation OAuth credential.
// It wraps an oauth2.TokenSource so that expired access tokens are automatically
// refreshed using the stored refresh token, matching the behavior of the Google Drive client.
func (c Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	cred, _, err := oneDriveCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error decoding onedrive credentials")
		return nil, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	tok := &oauth2.Token{
		AccessToken:  cred.AccessToken,
		RefreshToken: cred.RefreshToken,
		TokenType:    "Bearer",
	}

	if cred.Expiry != nil {
		tok.Expiry = *cred.Expiry
	}

	base := fmt.Sprintf(microsoftAuthBaseURL, "common")

	oauthCfg := &oauth2.Config{
		ClientID:     c.cfg.ClientID,
		ClientSecret: c.cfg.ClientSecret,
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

	// context background used intentionally in this slot
	ts := oauthCfg.TokenSource(context.Background(), tok)

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

	return &DriveClient{Graph: msgraphsdk.NewGraphServiceClient(adapter), TS: ts, Cfg: c.cfg}, nil
}

// oauthTokenCredential wraps an oauth2.TokenSource as an azcore.TokenCredential so that
// the kiota authentication provider can obtain automatically-refreshed access tokens
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
