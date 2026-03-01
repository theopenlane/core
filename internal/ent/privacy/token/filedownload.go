package token

import (
	"context"
)

// DownloadToken that implements the PrivacyToken interface
type DownloadToken struct {
	PrivacyToken
	token string
}

// NewDownloadTokenWithToken creates a new PrivacyToken of type SignUpToken with
// email set
func NewDownloadTokenWithToken(token string) DownloadToken {
	return DownloadToken{
		token: token,
	}
}

// GetToken from verify token
func (token *DownloadToken) GetToken() string {
	return token.token
}

// SetToken on the verify token
func (token *DownloadToken) SetToken(t string) {
	token.token = t
}

// NewContextWithDownloadToken returns a new context with the verify token inside
func NewContextWithDownloadToken(parent context.Context, downloadToken string) context.Context {
	ctx := downloadTokenContextKey.Set(parent, &DownloadToken{
		token: downloadToken,
	})

	return withTokenContextBypassCaller(ctx)
}

// DownloadTokenFromContext parses a context for a verify token and returns the token
func DownloadTokenFromContext(ctx context.Context) *DownloadToken {
	token, ok := downloadTokenContextKey.Get(ctx)
	if !ok {
		return nil
	}

	return token
}
