package token

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// AssesmentResponseToken that implements the PrivacyToken interface
type AssesmentResponseToken struct {
	PrivacyToken
	token string
	email string
}

// NewAssesmentResponseToken creates a new PrivacyToken of type TemplateResponderToken with
// token and email set
func NewAssesmentResponseToken(token, email string) AssesmentResponseToken {
	return AssesmentResponseToken{
		token: token,
		email: email,
	}
}

// GetToken from template responder token
func (t *AssesmentResponseToken) GetToken() string {
	return t.token
}

// SetToken on the template responder token
func (t *AssesmentResponseToken) SetToken(token string) {
	t.token = token
}

// GetEmail from template responder token
func (t *AssesmentResponseToken) GetEmail() string {
	return t.email
}

// SetEmail on the template responder token
func (t *AssesmentResponseToken) SetEmail(email string) {
	t.email = email
}

// NewContextWithTemplateResponderToken returns a new context with the template responder token inside
func NewContextWithTemplateResponderToken(parent context.Context, token, email string) context.Context {
	return contextx.With(parent, &AssesmentResponseToken{
		token: token,
		email: email,
	})
}

// TemplateResponderTokenFromContext parses a context for a template responder token and returns the token
func TemplateResponderTokenFromContext(ctx context.Context) *AssesmentResponseToken {
	token, ok := contextx.From[*AssesmentResponseToken](ctx)
	if !ok {
		return nil
	}

	return token
}
