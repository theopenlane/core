package token

import "context"

// OrgInviteToken that implements the PrivacyToken interface
type OrgInviteToken struct {
	PrivacyToken
	token string
}

type orgInviteTokenKey struct{}

// GetContextKey from OrgInviteToken
func (OrgInviteToken) GetContextKey() interface{} {
	return orgInviteTokenKey{}
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
	return context.WithValue(parent, orgInviteTokenKey{}, &OrgInviteToken{
		token: orgInviteToken,
	})
}

// OrgInviteTokenFromContext parses a context for a reset token and returns the token
func OrgInviteTokenFromContext(ctx context.Context) *OrgInviteToken {
	token, _ := ctx.Value(orgInviteTokenKey{}).(*OrgInviteToken)
	return token
}
