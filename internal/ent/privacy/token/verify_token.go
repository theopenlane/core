package token

import "context"

// VerifyToken that implements the PrivacyToken interface
type VerifyToken struct {
	PrivacyToken
	token string
}

type verifyTokenKey struct{}

// GetContextKey from VerifyToken
func (VerifyToken) GetContextKey() interface{} {
	return verifyTokenKey{}
}

// NewVerifyTokenWithToken creates a new PrivacyToken of type SignUpToken with
// email set
func NewVerifyTokenWithToken(token string) VerifyToken {
	return VerifyToken{
		token: token,
	}
}

// GetToken from verify token
func (token *VerifyToken) GetToken() string {
	return token.token
}

// SetToken on the verify token
func (token *VerifyToken) SetToken(t string) {
	token.token = t
}

// NewContextWithVerifyToken returns a new context with the verify token inside
func NewContextWithVerifyToken(parent context.Context, verifyToken string) context.Context {
	return context.WithValue(parent, verifyTokenKey{}, &VerifyToken{
		token: verifyToken,
	})
}

// VerifyTokenFromContext parses a context for a verify token and returns the token
func VerifyTokenFromContext(ctx context.Context) *VerifyToken {
	token, _ := ctx.Value(verifyTokenKey{}).(*VerifyToken)
	return token
}
