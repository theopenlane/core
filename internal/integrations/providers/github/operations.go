package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	githubOperationHealth types.OperationName = "health.default"
	githubOperationRepos  types.OperationName = "repos.collect_metadata"

	defaultPerPage = 50
	maxPerPage     = 100
	maxSampleSize  = 5

	defaultAlertState = "open"
	githubAPIVersion  = "2022-11-28"
	githubAPIBaseURL  = "https://api.github.com/"
)

type githubRepoOperationConfig struct {
	Visibility types.LowerString `json:"visibility,omitempty" jsonschema:"description=Optional visibility filter (all, public, private)"`
	PerPage    int               `json:"per_page,omitempty" jsonschema:"description=Override the number of repos fetched per page (max 100)."`
}

type githubVulnerabilityOperationConfig struct {
	AlertTypes      []types.LowerString   `json:"alert_types,omitempty" jsonschema:"description=Optional alert types to collect (dependabot, code_scanning, secret_scanning). Defaults to all."`
	Repositories    []types.TrimmedString `json:"repositories,omitempty" jsonschema:"description=Optional list of full repo names (owner/repo). If omitted, all accessible repos are scanned."`
	Visibility      types.LowerString     `json:"visibility,omitempty" jsonschema:"description=Optional visibility filter (all, public, private) when listing repos."`
	Affiliation     types.LowerString     `json:"affiliation,omitempty" jsonschema:"description=Optional repo affiliation filter (owner, collaborator, organization_member)."`
	PerPage         int                   `json:"per_page,omitempty" jsonschema:"description=Override the number of repos/alerts fetched per page (max 100)."`
	MaxRepos        int                   `json:"max_repos,omitempty" jsonschema:"description=Optional cap on the number of repositories to scan."`
	IncludePayloads bool                  `json:"include_payloads,omitempty" jsonschema:"description=Return raw alert payloads in the response (defaults to false)."`
	AlertState      types.LowerString     `json:"alert_state,omitempty" jsonschema:"description=Dependabot alert state filter (open, dismissed, fixed, all). Defaults to open."`
	Severity        types.LowerString     `json:"severity,omitempty" jsonschema:"description=Optional severity filter (low, medium, high, critical)."`
	Ecosystem       types.LowerString     `json:"ecosystem,omitempty" jsonschema:"description=Optional package ecosystem filter (npm, maven, pip, etc.)."`
}

var (
	githubRepoConfigSchema          = operations.SchemaFrom[githubRepoOperationConfig]()
	githubVulnerabilityConfigSchema = operations.SchemaFrom[githubVulnerabilityOperationConfig]()
)

// githubOperations returns the GitHub operations supported by this provider.
func githubOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(githubOperationHealth, "Validate GitHub OAuth token by calling the /user endpoint.", ClientGitHubAPI, runGitHubHealthOperation),
		{
			Name:         githubOperationRepos,
			Kind:         types.OperationKindCollectFindings,
			Description:  "Collect repository metadata for the authenticated account.",
			Client:       ClientGitHubAPI,
			Run:          runGitHubRepoOperation,
			ConfigSchema: githubRepoConfigSchema,
		},
		{
			Name:         types.OperationVulnerabilitiesCollect,
			Kind:         types.OperationKindCollectFindings,
			Description:  "Collect GitHub vulnerability alerts (Dependabot, code scanning, secret scanning) for repositories accessible to the token.",
			Client:       ClientGitHubAPI,
			Run:          runGitHubVulnerabilityOperation,
			ConfigSchema: githubVulnerabilityConfigSchema,
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
	client, token, err := auth.ClientAndOAuthToken(input, input.Provider)
	if err != nil {
		return types.OperationResult{}, err
	}

	if input.Provider == TypeGitHubApp {
		if client != nil {
			var resp githubInstallationRepositoriesResponse
			endpoint := githubAPIBaseURL + "installation/repositories?per_page=1"
			if err := client.GetJSON(ctx, endpoint, &resp); err != nil {
				return operations.OperationFailure("GitHub App installation lookup failed", err), err
			}

			return types.OperationResult{
				Status:  types.OperationStatusOK,
				Summary: fmt.Sprintf("GitHub App installation token valid (%d repositories accessible)", len(resp.Repositories)),
				Details: map[string]any{"repositories": len(resp.Repositories)},
			}, nil
		}

		repos, err := listGitHubInstallationRepos(ctx, nil, token, githubVulnerabilityConfig{
			Pagination: operations.Pagination{PerPage: 1},
		})
		if err != nil {
			return operations.OperationFailure("GitHub App installation lookup failed", err), err
		}

		return types.OperationResult{
			Status:  types.OperationStatusOK,
			Summary: fmt.Sprintf("GitHub App installation token valid (%d repositories accessible)", len(repos)),
			Details: map[string]any{"repositories": len(repos)},
		}, nil
	}

	var user githubUserResponse
	if err := fetchGitHubResource(ctx, client, token, "user", nil, &user); err != nil {
		return operations.OperationFailure("GitHub user lookup failed", err), err
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
	client, token, err := auth.ClientAndOAuthToken(input, input.Provider)
	if err != nil {
		return types.OperationResult{}, err
	}

	repoConfig, err := decodeGitHubRepoConfig(input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	config := githubVulnerabilityConfig{
		Pagination: operations.Pagination{PerPage: repoConfig.PerPage},
		Visibility: repoConfig.Visibility,
	}

	var repos []githubRepoResponse
	repos, err = listGitHubReposForProvider(ctx, client, token, input.Provider, config)
	if err != nil {
		return operations.OperationFailure("GitHub repository collection failed", err), err
	}

	samples := make([]map[string]any, 0, min(maxSampleSize, len(repos)))
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
func fetchGitHubResource(ctx context.Context, client *auth.AuthenticatedClient, token, path string, params url.Values, out any) error {
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

	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, headers, out); err != nil {
		if errors.Is(err, auth.ErrHTTPRequestFailed) {
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

// decodeGitHubRepoConfig decodes repo collection config into a typed struct.
func decodeGitHubRepoConfig(config map[string]any) (githubRepoOperationConfig, error) {
	var decoded githubRepoOperationConfig
	if err := operations.DecodeConfig(config, &decoded); err != nil {
		return decoded, err
	}

	return decoded, nil
}
