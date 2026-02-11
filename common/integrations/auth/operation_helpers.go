package auth

import "github.com/theopenlane/core/common/integrations/types"

// ClientAndOAuthToken returns the optional authenticated client and OAuth token.
func ClientAndOAuthToken(input types.OperationInput) (*AuthenticatedClient, string, error) {
	client := AuthenticatedClientFromAny(input.Client)
	token, err := OAuthTokenFromPayload(input.Credential)
	if err != nil {
		return client, "", err
	}

	return client, token, nil
}

// ClientAndAPIToken returns the optional authenticated client and API token.
func ClientAndAPIToken(input types.OperationInput) (*AuthenticatedClient, string, error) {
	client := AuthenticatedClientFromAny(input.Client)
	token, err := APITokenFromPayload(input.Credential)
	if err != nil {
		return client, "", err
	}

	return client, token, nil
}
