package models

import (
	"encoding/json"
	"io"
	"time"

	"github.com/rs/zerolog/log"
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

// MarshalGQL implements the graphql.Marshaler interface for gqlgen scalar serialization
func (c CredentialSet) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(c)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling credential set")
	}

	if _, err = w.Write(byteData); err != nil {
		log.Fatal().Err(err).Msg("error writing credential set")
	}
}

// UnmarshalGQL implements the graphql.Unmarshaler interface for gqlgen scalar deserialization
func (c *CredentialSet) UnmarshalGQL(v any) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, c)
}
