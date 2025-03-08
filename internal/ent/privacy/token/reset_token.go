package token

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// ResetToken that implements the PrivacyToken interface
type ResetToken struct {
	PrivacyToken
	token string
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
	return contextx.With(parent, &ResetToken{
		token: resetToken,
	})
}

// ResetTokenFromContext parses a context for a reset token and returns the token
func ResetTokenFromContext(ctx context.Context) *ResetToken {
	token, ok := contextx.From[*ResetToken](ctx)
	if !ok {
		return nil
	}

	return token
}
