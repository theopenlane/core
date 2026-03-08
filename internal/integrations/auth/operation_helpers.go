package auth

import (
	"maps"
	"strings"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TokenExtractor extracts a token string from a credential payload.
type TokenExtractor func(models.CredentialSet) (string, error)

// ResolveAuthenticatedClient returns a resolved HTTP client for operation execution.
// If a pooled client is present, it is cloned and reused; otherwise a new one is built
// from the credential token extractor and supplied base URL/headers.
func ResolveAuthenticatedClient(input types.OperationInput, extract TokenExtractor, baseURL string, headers map[string]string) (*AuthenticatedClient, error) {
	if pooled := AuthenticatedClientFromClient(input.Client); pooled != nil {
		client := pooled.clone()
		if client == nil {
			return nil, ErrClientResolutionFailed
		}
		if strings.TrimSpace(client.BaseURL) == "" && strings.TrimSpace(baseURL) != "" {
			client.BaseURL = strings.TrimSpace(baseURL)
		}
		if len(client.Headers) == 0 && len(headers) > 0 {
			client.Headers = maps.Clone(headers)
		}

		return client, nil
	}

	token, err := extract(input.Credential)
	if err != nil {
		return nil, err
	}

	return NewAuthenticatedClient(baseURL, token, headers), nil
}
