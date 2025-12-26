package models

import (
	"encoding/json"
	"io"
	"time"

	"github.com/rs/zerolog/log"
)

// CredentialSet is a custom type for pricing data
type CredentialSet struct {
	// AccessKeyID for cloud providers
	AccessKeyID string `json:"accessKeyID"`
	// SecretAccessKey for cloud providers
	SecretAccessKey string `json:"secretAccessKey"`
	// ProjectID for GCS
	ProjectID string `json:"projectID"`
	// AccountID for Cloudflare R2
	AccountID string `json:"accountID"`
	// APIToken for Cloudflare R2
	APIToken string `json:"apiToken"`
	// ProviderData stores provider-specific metadata or attributes
	ProviderData map[string]any `json:"providerData,omitempty"`
	// OAuthAccessToken holds the OAuth access token when applicable
	OAuthAccessToken string `json:"oauthAccessToken,omitempty"`
	// OAuthRefreshToken holds the OAuth refresh token when applicable
	OAuthRefreshToken string `json:"oauthRefreshToken,omitempty"`
	// OAuthTokenType stores the OAuth token type (e.g., Bearer)
	OAuthTokenType string `json:"oauthTokenType,omitempty"`
	// OAuthExpiry stores the token expiry timestamp
	OAuthExpiry *time.Time `json:"oauthExpiry,omitempty"`
	// Claims stores serialized ID token claims if available
	Claims map[string]any `json:"claims,omitempty"`
}

// String returns a string representation of the CredentialSet with sensitive fields masked for logging
func (c CredentialSet) String() string {
	// Mask sensitive information for logging/debugging only
	masked := make(map[string]string)
	if c.AccessKeyID != "" {
		masked["accessKeyID"] = "***"
	}

	if c.SecretAccessKey != "" {
		masked["secretAccessKey"] = "***"
	}

	if c.ProjectID != "" {
		masked["projectID"] = c.ProjectID
	}

	if c.AccountID != "" {
		masked["accountID"] = c.AccountID
	}

	if c.APIToken != "" {
		masked["apiToken"] = "***"
	}
	if c.OAuthAccessToken != "" {
		masked["oauthAccessToken"] = "***"
	}
	if c.OAuthRefreshToken != "" {
		masked["oauthRefreshToken"] = "***"
	}

	data, _ := json.Marshal(masked)

	return string(data)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (c CredentialSet) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(c)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling json object")
	}

	_, err = w.Write(byteData)
	if err != nil {
		log.Fatal().Err(err).Msg("error writing json object")
	}
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (c *CredentialSet) UnmarshalGQL(v interface{}) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteData, c)
	if err != nil {
		return err
	}

	return nil
}
