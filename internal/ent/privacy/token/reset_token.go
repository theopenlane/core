package token

import "context"

// ResetToken that implements the PrivacyToken interface
type ResetToken struct {
	PrivacyToken
	token string
}

type resetTokenKey struct{}

// GetContextKey from ResetToken
func (ResetToken) GetContextKey() interface{} {
	return resetTokenKey{}
}

// NewResetTokenWithToken creates a new PrivacyToken of type ResetToken with
// token set
func NewResetTokenWithToken(token string) ResetToken {
	return ResetToken{
		token: token,
	}
}

// GetToken from reset token
func (token *ResetToken) GetToken() string {
	return token.token
}

// SetToken on the reset token
func (token *ResetToken) SetToken(t string) {
	token.token = t
}

// NewContextWithResetToken returns a new context with the reset token inside
func NewContextWithResetToken(parent context.Context, resetToken string) context.Context {
	return context.WithValue(parent, resetTokenKey{}, &ResetToken{
		token: resetToken,
	})
}

// ResetTokenFromContext parses a context for a reset token and returns the token
func ResetTokenFromContext(ctx context.Context) *ResetToken {
	token, _ := ctx.Value(resetTokenKey{}).(*ResetToken)
	return token
}
