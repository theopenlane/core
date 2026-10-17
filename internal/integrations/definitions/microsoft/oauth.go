package microsoft

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"golang.org/x/oauth2"
)

// delegatedTokenSource builds a token source that refreshes stored delegated OAuth material against the given tenant segment
func (c AppConfig) delegatedTokenSource(ctx context.Context, tenant string, scopes []string, token *oauth2.Token) oauth2.TokenSource {
	oauthCfg := &oauth2.Config{
		ClientID:     c.ApplicationID,
		ClientSecret: c.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthorizeURL(tenant),
			TokenURL: TokenURL(tenant),
		},
		Scopes: scopes,
	}

	return oauthCfg.TokenSource(ctx, token)
}

var _ azcore.TokenCredential = (*OAuthTokenCredential)(nil)

// OAuthTokenCredential wraps an oauth2.TokenSource as an azcore.TokenCredential so that
// the kiota authentication provider can obtain automatically-refreshed access tokens
type OAuthTokenCredential struct {
	// ts is the refreshing token source
	ts oauth2.TokenSource
}

// NewOAuthTokenCredential wraps a token source as an azcore.TokenCredential
func NewOAuthTokenCredential(ts oauth2.TokenSource) *OAuthTokenCredential {
	return &OAuthTokenCredential{ts: ts}
}

// GetToken obtains the current (or refreshed) access token from the underlying oauth2.TokenSource
func (c *OAuthTokenCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
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
