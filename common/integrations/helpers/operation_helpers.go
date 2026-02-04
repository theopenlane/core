package helpers

import "github.com/theopenlane/core/common/integrations/types"

// ClientAndOAuthToken returns the optional authenticated client and OAuth token
func ClientAndOAuthToken(input types.OperationInput, provider types.ProviderType) (*AuthenticatedClient, string, error) {
	client := AuthenticatedClientFromAny(input.Client)
	token, err := OAuthTokenFromPayload(input.Credential, string(provider))
	if err != nil {
		return client, "", err
	}
	return client, token, nil
}

// ClientAndAPIToken returns the optional authenticated client and API token
func ClientAndAPIToken(input types.OperationInput, provider types.ProviderType) (*AuthenticatedClient, string, error) {
	client := AuthenticatedClientFromAny(input.Client)
	token, err := APITokenFromPayload(input.Credential, string(provider))
	if err != nil {
		return client, "", err
	}
	return client, token, nil
}

// OperationFailure builds a failed operation result with an error detail
func OperationFailure(summary string, err error) types.OperationResult {
	if err == nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: summary,
		}
	}
	return types.OperationResult{
		Status:  types.OperationStatusFailed,
		Summary: summary,
		Details: map[string]any{"error": err.Error()},
	}
}
