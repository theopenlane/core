package types

import (
	"encoding/json"
	"time"

	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

// CredentialSet holds the credential fields for an integration;
// auth-type-specific fields (access keys, service account keys, etc.) are
// serialized into ProviderData by each provider's Mint implementation
type CredentialSet struct {
	// OAuthAccessToken holds the OAuth access token when applicable
	OAuthAccessToken string `json:"oauthAccessToken,omitempty"`
	// OAuthRefreshToken holds the OAuth refresh token when applicable
	OAuthRefreshToken string `json:"oauthRefreshToken,omitempty"`
	// OAuthTokenType stores the OAuth token type (e.g. Bearer)
	OAuthTokenType string `json:"oauthTokenType,omitempty"`
	// OAuthExpiry stores the token expiry timestamp
	OAuthExpiry *time.Time `json:"oauthExpiry,omitempty"`
	// ProviderData holds provider-specific credential data serialized by the provider's Mint implementation
	ProviderData json.RawMessage `json:"providerData,omitempty"`
	// Claims stores serialized ID token claims if available
	Claims map[string]any `json:"claims,omitempty"`
}

// CloneCredentialSet returns a deep copy of a CredentialSet
func CloneCredentialSet(set CredentialSet) CredentialSet {
	cloned := set
	cloned.ProviderData = jsonx.CloneRawMessage(set.ProviderData)
	cloned.Claims = mapx.DeepCloneMapAny(set.Claims)

	if set.OAuthExpiry != nil {
		expiry := *set.OAuthExpiry
		cloned.OAuthExpiry = &expiry
	}

	return cloned
}

