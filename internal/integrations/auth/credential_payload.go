package auth

import (
	"encoding/json"

	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/jsonx"
)

// BuildOAuthCredentialSet builds OAuth/OIDC credential fields from upstream token/claims.
func BuildOAuthCredentialSet(token *oauth2.Token, claims *oidc.IDTokenClaims) (models.CredentialSet, error) {
	credential := models.CredentialSet{}

	if token != nil {
		credential.OAuthAccessToken = token.AccessToken
		credential.OAuthRefreshToken = token.RefreshToken
		credential.OAuthTokenType = token.TokenType
		if !token.Expiry.IsZero() {
			exp := token.Expiry.UTC()
			credential.OAuthExpiry = &exp
		}
	}

	if claims != nil {
		claimsMap, err := jsonx.ToMap(claims)
		if err != nil {
			return models.CredentialSet{}, err
		}
		credential.Claims = claimsMap
	}

	return credential, nil
}

// BuildAPITokenCredentialSet builds API token credentials with optional provider metadata.
func BuildAPITokenCredentialSet(token string, providerData json.RawMessage) models.CredentialSet {
	credential := models.CredentialSet{APIToken: token}
	if len(providerData) > 0 {
		credential.ProviderData = providerData
	}
	return credential
}
