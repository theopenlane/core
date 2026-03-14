package providerkit

import (
	"strings"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/jsonx"
)

// OAuthTokenFromCredential extracts a usable access token from the credential set.
// Returns ErrOAuthTokenMissing if all OAuth fields are empty, or ErrAccessTokenEmpty
// if the access token specifically is missing.
func OAuthTokenFromCredential(credential models.CredentialSet) (string, error) {
	if credential.OAuthAccessToken == "" &&
		credential.OAuthRefreshToken == "" &&
		credential.OAuthTokenType == "" &&
		credential.OAuthExpiry == nil {
		return "", ErrOAuthTokenMissing
	}

	if credential.OAuthAccessToken == "" {
		return "", ErrAccessTokenEmpty
	}

	return credential.OAuthAccessToken, nil
}

// APITokenFromCredential extracts a raw API token from the credential set's ProviderData.
// The token is expected to be stored under the "apiToken" key in the JSON object.
// Returns ErrAPITokenMissing if the key is absent or the value is empty.
func APITokenFromCredential(credential models.CredentialSet) (string, error) {
	metadata, err := jsonx.ToRawMap(credential.ProviderData)
	if err != nil {
		return "", ErrAPITokenMissing
	}

	raw, ok := metadata["apiToken"]
	if !ok || len(raw) == 0 {
		return "", ErrAPITokenMissing
	}

	var token string
	if err := jsonx.RoundTrip(raw, &token); err != nil {
		return "", ErrAPITokenMissing
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrAPITokenMissing
	}

	return token, nil
}
