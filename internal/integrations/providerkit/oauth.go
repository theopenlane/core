package providerkit

import (
	"context"
	"time"

	"github.com/samber/lo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// bearerTokenType is the OAuth2 token type attached to stored access tokens
const bearerTokenType = "Bearer"

// OAuthToken builds a refreshable bearer token from stored OAuth credential material
func OAuthToken(accessToken, refreshToken string, expiry *time.Time) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    bearerTokenType,
		Expiry:       lo.FromPtr(expiry),
	}
}

// GoogleTokenSource builds a token source refreshing against the Google endpoint
func GoogleTokenSource(ctx context.Context, clientID, clientSecret string, token *oauth2.Token) oauth2.TokenSource {
	return (&oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
	}).TokenSource(ctx, token)
}
