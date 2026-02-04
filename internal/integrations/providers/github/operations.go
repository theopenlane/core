package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/opsconfig"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	githubOperationHealth types.OperationName = "health.default"
	githubOperationRepos  types.OperationName = "repos.collect_metadata"
	githubOperationVulns  types.OperationName = "vulnerabilities.collect"

	defaultPerPage = 50
	maxPerPage     = 100
	maxSampleSize  = 5

	defaultAlertState = "open"
	githubAPIVersion  = "2022-11-28"
	githubAPIBaseURL  = "https://api.github.com/"
)

// githubOperations returns the GitHub operations supported by this provider.
func githubOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        githubOperationHealth,
			Kind:        types.OperationKindHealth,
			Description: "Validate GitHub OAuth token by calling the /user endpoint.",
			Client:      ClientGitHubAPI,
			Run:         runGitHubHealthOperation,
		},
		{
			Name:        githubOperationRepos,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect repository metadata for the authenticated account.",
			Client:      ClientGitHubAPI,
			Run:         runGitHubRepoOperation,
			ConfigSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"visibility": map[string]any{
						"type":        "string",
						"description": "Optional visibility filter (all, public, private)",
					},
					"per_page": map[string]any{
						"type":        "integer",
						"description": "Override the number of repos fetched per page (max 100).",
					},
				},
			},
		},
		{
			Name:        githubOperationVulns,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect GitHub vulnerability alerts (Dependabot, code scanning, secret scanning) for repositories accessible to the token.",
			Client:      ClientGitHubAPI,
			Run:         runGitHubVulnerabilityOperation,
			ConfigSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"alert_types": map[string]any{
						"type":        "array",
						"description": "Optional alert types to collect (dependabot, code_scanning, secret_scanning). Defaults to all.",
						"items": map[string]any{
							"type": "string",
						},
					},
					"repositories": map[string]any{
						"type":        "array",
						"description": "Optional list of full repo names (owner/repo). If omitted, all accessible repos are scanned.",
						"items": map[string]any{
							"type": "string",
						},
					},
					"visibility": map[string]any{
						"type":        "string",
						"description": "Optional visibility filter (all, public, private) when listing repos.",
					},
					"affiliation": map[string]any{
						"type":        "string",
						"description": "Optional repo affiliation filter (owner, collaborator, organization_member).",
					},
					"per_page": map[string]any{
						"type":        "integer",
						"description": "Override the number of repos/alerts fetched per page (max 100).",
					},
					"max_repos": map[string]any{
						"type":        "integer",
						"description": "Optional cap on the number of repositories to scan.",
					},
					"include_payloads": map[string]any{
						"type":        "boolean",
						"description": "Return raw alert payloads in the response (defaults to false).",
					},
					"alert_state": map[string]any{
						"type":        "string",
						"description": "Dependabot alert state filter (open, dismissed, fixed, all). Defaults to open.",
					},
					"severity": map[string]any{
						"type":        "string",
						"description": "Optional severity filter (low, medium, high, critical).",
					},
					"ecosystem": map[string]any{
						"type":        "string",
						"description": "Optional package ecosystem filter (npm, maven, pip, etc.).",
					},
				},
			},
		},
	}
}

type githubUserResponse struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Name  string `json:"name"`
}

type githubRepoResponse struct {
	Name      string          `json:"name"`
	FullName  string          `json:"full_name"`
	Owner     githubRepoOwner `json:"owner"`
	Private   bool            `json:"private"`
	UpdatedAt time.Time       `json:"updated_at"`
	HTMLURL   string          `json:"html_url"`
}

type githubRepoOwner struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

// runGitHubHealthOperation validates GitHub access for OAuth or App credentials.
func runGitHubHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := helpers.ClientAndOAuthToken(input, input.Provider)
	if err != nil {
		return types.OperationResult{}, err
	}

	if input.Provider == TypeGitHubApp {
		if client != nil {
			var resp githubInstallationRepositoriesResponse
			endpoint := githubAPIBaseURL + "installation/repositories?per_page=1"
			if err := client.GetJSON(ctx, endpoint, &resp); err != nil {
				return helpers.OperationFailure("GitHub App installation lookup failed", err), err
			}

			return types.OperationResult{
				Status:  types.OperationStatusOK,
				Summary: fmt.Sprintf("GitHub App installation token valid (%d repositories accessible)", len(resp.Repositories)),
				Details: map[string]any{"repositories": len(resp.Repositories)},
			}, nil
		}

		repos, err := listGitHubInstallationRepos(ctx, nil, token, githubVulnerabilityConfig{
			Pagination: opsconfig.Pagination{PerPage: 1},
		})
		if err != nil {
			return helpers.OperationFailure("GitHub App installation lookup failed", err), err
		}

		return types.OperationResult{
			Status:  types.OperationStatusOK,
			Summary: fmt.Sprintf("GitHub App installation token valid (%d repositories accessible)", len(repos)),
			Details: map[string]any{"repositories": len(repos)},
		}, nil
	}

	var user githubUserResponse
	if err := fetchGitHubResource(ctx, client, token, "user", nil, &user); err != nil {
		return helpers.OperationFailure("GitHub user lookup failed", err), err
	}

	details := map[string]any{
		"login": user.Login,
		"id":    user.ID,
		"name":  user.Name,
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("GitHub token valid for %s", user.Login),
		Details: details,
	}, nil
}

// runGitHubRepoOperation lists repositories for the authenticated account.
func runGitHubRepoOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := helpers.ClientAndOAuthToken(input, input.Provider)
	if err != nil {
		return types.OperationResult{}, err
	}

	config, err := decodeGitHubVulnerabilityConfig(input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	var repos []githubRepoResponse
	repos, err = listGitHubReposForProvider(ctx, client, token, input.Provider, config)
	if err != nil {
		return helpers.OperationFailure("GitHub repository collection failed", err), err
	}

	samples := make([]map[string]any, 0, minInt(maxSampleSize, len(repos)))
	for _, repo := range repos {
		if len(samples) >= cap(samples) {
			break
		}
		samples = append(samples, map[string]any{
			"name":       repo.Name,
			"private":    repo.Private,
			"updated_at": repo.UpdatedAt,
			"url":        repo.HTMLURL,
		})
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d repositories", len(repos)),
		Details: map[string]any{
			"count":   len(repos),
			"samples": samples,
		},
	}, nil
}

// fetchGitHubResource retrieves GitHub REST API resources with optional pooled client support.
func fetchGitHubResource(ctx context.Context, client *helpers.AuthenticatedClient, token, path string, params url.Values, out any) error {
	endpoint := githubAPIBaseURL + path
	if params != nil {
		if encoded := params.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	headers := map[string]string{
		"Accept":               "application/vnd.github+json",
		"X-GitHub-Api-Version": githubAPIVersion,
	}

	if err := helpers.GetJSONWithClient(ctx, client, endpoint, token, headers, out); err != nil {
		if errors.Is(err, helpers.ErrHTTPRequestFailed) {
			return fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}
		return err
	}

	return nil
}

// clampPerPage bounds the per-page value for GitHub API requests
func clampPerPage(value int) int {
	if value <= 0 {
		return defaultPerPage
	}

	if value > maxPerPage {
		return maxPerPage
	}

	return value
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
