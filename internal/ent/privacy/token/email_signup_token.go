package token

import (
	"context"
)

// PrivacyToken interface
type PrivacyToken interface {
	GetContextKey() interface{}
}

// SignUpToken that implements the PrivacyToken interface
type SignUpToken struct {
	PrivacyToken
	email string
}

type signUpTokenKey struct{}

// NewSignUpTokenWithEmail creates a new PrivacyToken of type SignUpToken with
// email set
func NewSignUpTokenWithEmail(email string) SignUpToken {
	return SignUpToken{
		email: email,
	}
}

// GetEmail from sign-up token
func (token *SignUpToken) GetEmail() string {
	return token.email
}

// SetEmail on the sign-up token
func (token *SignUpToken) SetEmail(email string) {
	token.email = email
}

// GetContextKey from SignUpToken
func (SignUpToken) GetContextKey() interface{} {
	return signUpTokenKey{}
}

// NewContextWithSignUpToken creates a new context with a sign-up token. It takes a
// parent context and a sign-up token as parameters and returns a new context with
// the sign-up token added
func NewContextWithSignUpToken(parent context.Context, email string) context.Context {
	return context.WithValue(parent, signUpTokenKey{}, &SignUpToken{
		email: email,
	})
}

// EmailSignUpTokenFromContext retrieves the value associated with the
// signUpTokenKey key from the context.
// It then type asserts the value to an EmailSignUpToken and returns it. If the
// value is not of type EmailSignUpToken, it returns nil
func EmailSignUpTokenFromContext(ctx context.Context) *SignUpToken {
	token, _ := ctx.Value(signUpTokenKey{}).(*SignUpToken)
	return token
}
