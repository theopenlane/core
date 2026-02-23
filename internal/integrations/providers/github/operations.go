package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	githubOperationHealth      types.OperationName = "health.default"
	githubOperationRepos       types.OperationName = "repos.collect_metadata"
	githubOperationOrgRepos    types.OperationName = "repos.collect_org_metadata"
	githubOperationVulnCollect types.OperationName = types.OperationVulnerabilitiesCollect

	defaultPerPage = 50
	maxPerPage     = 100

	defaultAlertState = "open"
	githubAPIVersion  = "2022-11-28"
	githubAPIBaseURL  = "https://api.github.com/"
)

type githubRepoOperationConfig struct {
	// Visibility filters repositories by visibility
	Visibility types.LowerString `json:"visibility,omitempty" jsonschema:"description=Optional visibility filter (all, public, private)"`
	// PerPage overrides the number of repositories requested per page
	PerPage int `json:"per_page,omitempty" jsonschema:"description=Override the number of repos fetched per page (max 100)."`
}

type githubVulnerabilityOperationConfig struct {
	// AlertTypes selects which alert types to collect
	AlertTypes []types.LowerString `json:"alert_types,omitempty" jsonschema:"description=Optional alert types to collect (dependabot, code_scanning, secret_scanning). Defaults to all."`
	// Repositories lists explicit repositories to scan
	Repositories []types.TrimmedString `json:"repositories,omitempty" jsonschema:"description=Optional list of full repo names (owner/repo). If omitted, all accessible repos are scanned."`
	// Visibility filters repositories by visibility when listing
	Visibility types.LowerString `json:"visibility,omitempty" jsonschema:"description=Optional visibility filter (all, public, private) when listing repos."`
	// Affiliation filters repositories by affiliation
	Affiliation types.LowerString `json:"affiliation,omitempty" jsonschema:"description=Optional repo affiliation filter (owner, collaborator, organization_member)."`
	// PerPage overrides the number of repos or alerts requested per page
	PerPage int `json:"per_page,omitempty" jsonschema:"description=Override the number of repos/alerts fetched per page (max 100)."`
	// MaxRepos caps the number of repositories to scan
	MaxRepos int `json:"max_repos,omitempty" jsonschema:"description=Optional cap on the number of repositories to scan."`
	// IncludePayloads controls whether raw payloads are returned
	IncludePayloads bool `json:"include_payloads,omitempty" jsonschema:"description=Return raw alert payloads in the response (defaults to false)."`
	// AlertState filters Dependabot alert state
	AlertState types.LowerString `json:"alert_state,omitempty" jsonschema:"description=Dependabot alert state filter (open, dismissed, fixed, all). Defaults to open."`
	// Severity filters alerts by severity
	Severity types.LowerString `json:"severity,omitempty" jsonschema:"description=Optional severity filter (low, medium, high, critical)."`
	// Ecosystem filters alerts by package ecosystem
	Ecosystem types.LowerString `json:"ecosystem,omitempty" jsonschema:"description=Optional package ecosystem filter (npm, maven, pip, etc.)."`
}

var (
	githubRepoConfigSchema          = operations.SchemaFrom[githubRepoOperationConfig]()
	githubOrgRepoConfigSchema       = operations.SchemaFrom[githubOrgRepoOperationConfig]()
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
		githubOrganizationRepoOperationDescriptor(),
		{
			Name:         githubOperationVulnCollect,
			Kind:         types.OperationKindCollectFindings,
			Description:  "Collect GitHub vulnerability alerts (Dependabot, code scanning, secret scanning) for repositories accessible to the token.",
			Client:       ClientGitHubAPI,
			Run:          runGitHubVulnerabilityOperation,
			ConfigSchema: githubVulnerabilityConfigSchema,
		},
	}
}

// githubAppOperations returns the GitHub App operations supported by this provider.
func githubAppOperations(baseURL string) []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(
			githubOperationHealth,
			"Validate GitHub App installation token by calling the installation repositories endpoint.",
			ClientGitHubAPI,
			runGitHubAppHealthOperation(baseURL),
		),
		githubOrganizationRepoOperationDescriptor(),
	}
}

// githubOrganizationRepoOperationDescriptor builds the shared org repository GraphQL descriptor.
func githubOrganizationRepoOperationDescriptor() types.OperationDescriptor {
	return types.OperationDescriptor{
		Name:         githubOperationOrgRepos,
		Kind:         types.OperationKindCollectFindings,
		Description:  "Collect repository metadata for a GitHub organization using GraphQL.",
		Client:       ClientGitHubGraphQL,
		Run:          runGitHubOrganizationReposOperation,
		ConfigSchema: githubOrgRepoConfigSchema,
	}
}

type githubUserResponse struct {
	// Login is the GitHub username
	Login string `json:"login"`
	// ID is the GitHub user identifier
	ID int64 `json:"id"`
	// Name is the display name for the user
	Name string `json:"name"`
}

type githubRepoResponse struct {
	// Name is the repository name
	Name string `json:"name"`
	// FullName is the owner/name identifier
	FullName string `json:"full_name"`
	// Owner describes the repository owner
	Owner githubRepoOwner `json:"owner"`
	// Private reports whether the repository is private
	Private bool `json:"private"`
	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time `json:"updated_at"`
	// HTMLURL is the web URL for the repository
	HTMLURL string `json:"html_url"`
}

type githubRepoOwner struct {
	// Login is the owner login name
	Login string `json:"login"`
	// ID is the owner identifier
	ID int64 `json:"id"`
}

// githubAppInstallationReposResponse models the installation repositories response.
type githubAppInstallationReposResponse struct {
	// TotalCount is the total number of repositories.
	TotalCount int `json:"total_count"`
	// Repositories lists repositories visible to the installation.
	Repositories []githubRepoResponse `json:"repositories"`
}

// runGitHubHealthOperation validates GitHub OAuth credentials.
func runGitHubHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndOAuthToken(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	var user githubUserResponse
	if err := fetchGitHubResource(ctx, client, token, "user", nil, &user); err != nil {
		return operations.OperationFailure("GitHub user lookup failed", err, nil)
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

// runGitHubAppHealthOperation validates GitHub App installation tokens.
func runGitHubAppHealthOperation(baseURL string) types.OperationFunc {
	return func(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
		client, token, err := auth.ClientAndOAuthToken(input)
		if err != nil {
			return types.OperationResult{}, err
		}

		var resp githubAppInstallationReposResponse
		if err := fetchGitHubResourceWithBaseURL(ctx, client, token, baseURL, "installation/repositories", nil, &resp); err != nil {
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
	}
}

// runGitHubRepoOperation lists repositories for the authenticated account.
func runGitHubRepoOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndOAuthToken(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	repoConfig, err := operations.Decode[githubRepoOperationConfig](input.Config)
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
		return operations.OperationFailure("GitHub repository collection failed", err, nil)
	}

	samples := lo.Map(repos[:min(len(repos), operations.DefaultSampleSize)], func(repo githubRepoResponse, _ int) map[string]any {
		return map[string]any{
			"name":       repo.Name,
			"private":    repo.Private,
			"updated_at": repo.UpdatedAt,
			"url":        repo.HTMLURL,
		}
	})

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
	return fetchGitHubResourceWithBaseURL(ctx, client, token, githubAPIBaseURL, path, params, out)
}

// fetchGitHubResourceWithBaseURL retrieves GitHub REST API resources using a custom base URL.
func fetchGitHubResourceWithBaseURL(ctx context.Context, client *auth.AuthenticatedClient, token, baseURL, path string, params url.Values, out any) error {
	endpoint := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
	if params != nil {
		if encoded := params.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, githubClientHeaders, out); err != nil {
		if errors.Is(err, auth.ErrHTTPRequestFailed) {
			return ErrAPIRequest
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
