package auth

import "github.com/theopenlane/core/common/integrations/types"

// TokenExtractor extracts a token string from a credential payload.
type TokenExtractor func(types.CredentialPayload) (string, error)

// ClientAndToken returns the optional authenticated client and extracted token.
func ClientAndToken(input types.OperationInput, extract TokenExtractor) (*AuthenticatedClient, string, error) {
	client := AuthenticatedClientFromAny(input.Client)
	token, err := extract(input.Credential)

	return client, token, err
}
