package helpers

import (
	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/oidc"
)

// CloneOAuthToken returns a shallow copy of the token to prevent callers from modifying shared state
func CloneOAuthToken(token *oauth2.Token) *oauth2.Token {
	if token == nil {
		return nil
	}

	out := new(oauth2.Token)
	*out = *token

	return out
}

// CloneOIDCClaims returns a shallow copy of the claims struct
func CloneOIDCClaims(claims *oidc.IDTokenClaims) *oidc.IDTokenClaims {
	if claims == nil {
		return nil
	}

	out := new(oidc.IDTokenClaims)
	*out = *claims

	return out
}
