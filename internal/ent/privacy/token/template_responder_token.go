package token

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// TemplateResponderToken that implements the PrivacyToken interface
type TemplateResponderToken struct {
	PrivacyToken
	token string
	email string
}

// NewTemplateResponderToken creates a new PrivacyToken of type TemplateResponderToken with
// token and email set
func NewTemplateResponderToken(token, email string) TemplateResponderToken {
	return TemplateResponderToken{
		token: token,
		email: email,
	}
}

// GetToken from template responder token
func (t *TemplateResponderToken) GetToken() string {
	return t.token
}

// SetToken on the template responder token
func (t *TemplateResponderToken) SetToken(token string) {
	t.token = token
}

// GetEmail from template responder token
func (t *TemplateResponderToken) GetEmail() string {
	return t.email
}

// SetEmail on the template responder token
func (t *TemplateResponderToken) SetEmail(email string) {
	t.email = email
}

// NewContextWithTemplateResponderToken returns a new context with the template responder token inside
func NewContextWithTemplateResponderToken(parent context.Context, token, email string) context.Context {
	return contextx.With(parent, &TemplateResponderToken{
		token: token,
		email: email,
	})
}

// TemplateResponderTokenFromContext parses a context for a template responder token and returns the token
func TemplateResponderTokenFromContext(ctx context.Context) *TemplateResponderToken {
	token, ok := contextx.From[*TemplateResponderToken](ctx)
	if !ok {
		return nil
	}

	return token
}
