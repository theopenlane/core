package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

// githubAppOperationHealth is the default health operation name.
const githubAppOperationHealth types.OperationName = "health.default"

// githubAppInstallationReposResponse models the installation repositories response.
type githubAppInstallationReposResponse struct {
	// TotalCount is the total number of repositories.
	TotalCount int `json:"total_count"`
	// Repositories lists repositories visible to the installation.
	Repositories []githubRepoResponse `json:"repositories"`
}

// githubAppOperations returns GitHub App operation descriptors.
func githubAppOperations(baseURL string) []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        githubAppOperationHealth,
			Kind:        types.OperationKindHealth,
			Description: "Validate GitHub App installation token by calling the installation repositories endpoint.",
			Run: func(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
				token, err := auth.OAuthTokenFromPayload(input.Credential)
				if err != nil {
					return types.OperationResult{}, err
				}

				var resp githubAppInstallationReposResponse
				if err := fetchGitHubAppResource(ctx, baseURL, token, "installation/repositories", nil, &resp); err != nil {
					return operations.OperationFailure("GitHub App installation lookup failed", err, nil)
				}

				count := resp.TotalCount
				if count == 0 && len(resp.Repositories) > 0 {
					count = len(resp.Repositories)
				}

				return types.OperationResult{
					Status:  types.OperationStatusOK,
					Summary: fmt.Sprintf("GitHub App token valid for %d repositories", count),
					Details: map[string]any{"count": count},
				}, nil
			},
		},
	}
}

// fetchGitHubAppResource performs a GitHub App API GET request and decodes JSON responses.
func fetchGitHubAppResource(ctx context.Context, baseURL, token, path string, params url.Values, out any) error {
	base := strings.TrimRight(baseURL, "/")
	endpoint := base + "/" + strings.TrimLeft(path, "/")
	if params != nil {
		if encoded := params.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	return auth.HTTPGetJSON(ctx, nil, endpoint, token, githubClientHeaders, out)
}
