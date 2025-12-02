package token

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// OrgInviteToken that implements the PrivacyToken interface
type OrgInviteToken struct {
	PrivacyToken
	token string
}

// NewOrgInviteTokenWithToken creates a new PrivacyToken of type OrgInviteToken with
// token set
func NewOrgInviteTokenWithToken(token string) OrgInviteToken {
	return OrgInviteToken{
		token: token,
	}
}

// GetToken from invite token
func (token *OrgInviteToken) GetToken() string {
	return token.token
}

// SetToken on the invite token
func (token *OrgInviteToken) SetToken(t string) {
	token.token = t
}

// NewContextWithOrgInviteToken returns a new context with the reset token inside
func NewContextWithOrgInviteToken(parent context.Context, orgInviteToken string) context.Context {
	return contextx.With(parent, &OrgInviteToken{
		token: orgInviteToken,
	})
}

// OrgInviteTokenFromContext parses a context for a reset token and returns the token
func OrgInviteTokenFromContext(ctx context.Context) *OrgInviteToken {
	token, ok := contextx.From[*OrgInviteToken](ctx)
	if !ok {
		return nil
	}

	return token
}
