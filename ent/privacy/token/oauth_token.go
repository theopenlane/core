package token

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// OauthTooToken that implements the PrivacyToken interface
type OauthTooToken struct {
	PrivacyToken
	email string
}

// NewOauthTooWithEmail creates a new PrivacyToken of type OauthTooToken with
// email set
func NewOauthTooWithEmail(email string) OauthTooToken {
	return OauthTooToken{
		email: email,
	}
}

// GetEmail from oauth2 token
func (token *OauthTooToken) GetEmail() string {
	return token.email
}

// SetEmail on the oauth2 token
func (token *OauthTooToken) SetEmail(email string) {
	token.email = email
}

// NewContextWithOauthTooToken creates a new context with a oauth2 token. It takes a
// parent context and a oauth2 token as parameters and returns a new context with
// the oauth2 token added
func NewContextWithOauthTooToken(parent context.Context, email string) context.Context {
	return contextx.With(parent, &OauthTooToken{
		email: email,
	})
}

// OauthTooTokenFromContext retrieves the value associated with the
// oauthTooTokenKey key from the context.
// It then type asserts the value to an OauthTooToken and returns it. If the
// value is not of type OauthTooToken, it returns nil
func OauthTooTokenFromContext(ctx context.Context) *OauthTooToken {
	token, ok := contextx.From[*OauthTooToken](ctx)
	if !ok {
		return nil
	}

	return token
}
