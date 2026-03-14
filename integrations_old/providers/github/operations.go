package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	gh "github.com/google/go-github/v83/github"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	githubOperationHealth      types.OperationName = types.OperationHealthDefault
	githubOperationRepos       types.OperationName = "repos.collect_metadata"
	githubOperationOrgRepos    types.OperationName = "repos.collect_org_metadata"
	githubOperationVulnCollect types.OperationName = types.OperationVulnerabilitiesCollect

	defaultPerPage = 50
	maxPerPage     = 100

	defaultAlertState = "open"
	githubAPIVersion  = "2022-11-28"
	githubAPIBaseURL  = "https://api.github.com"
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
	githubRepoConfigSchema          = providerkit.SchemaFrom[githubRepoOperationConfig]()
	githubOrgRepoConfigSchema       = providerkit.SchemaFrom[githubOrgRepoOperationConfig]()
	githubVulnerabilityConfigSchema = providerkit.SchemaFrom[githubVulnerabilityOperationConfig]()
)

// githubOperations returns the GitHub operations supported by this provider.
func githubOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(githubOperationHealth, "Validate GitHub OAuth token by calling the /user endpoint.", ClientGitHubAPI, runGitHubHealthOperation),
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
			Ingest: []types.IngestContract{
				{
					Schema:         mappingSchemaVulnerability,
					EnsurePayloads: true,
				},
			},
		},
	}
}

// githubAppOperations returns the GitHub App operations supported by this provider.
func githubAppOperations(baseURL string) []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(
			githubOperationHealth,
			"Validate GitHub App installation token by calling the installation repositories endpoint.",
			ClientGitHubAPI,
			runGitHubAppHealthOperation(baseURL),
		),
		githubOrganizationRepoOperationDescriptor(),
		{
			Name:         githubOperationVulnCollect,
			Kind:         types.OperationKindCollectFindings,
			Description:  "Collect GitHub vulnerability alerts (Dependabot, code scanning, secret scanning) for repositories accessible to the app installation.",
			Client:       ClientGitHubAPI,
			Run:          runGitHubVulnerabilityOperation,
			ConfigSchema: githubVulnerabilityConfigSchema,
			Ingest: []types.IngestContract{
				{
					Schema:         mappingSchemaVulnerability,
					EnsurePayloads: true,
				},
			},
		},
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

type githubHealthDetails struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Name  string `json:"name"`
}

type githubAppHealthDetails struct {
	Count int `json:"count"`
}

type githubRepoSample struct {
	Name      string    `json:"name"`
	Private   bool      `json:"private"`
	UpdatedAt time.Time `json:"updated_at"`
	URL       string    `json:"url"`
}

type githubRepoCollectionDetails struct {
	Count   int                `json:"count"`
	Samples []githubRepoSample `json:"samples"`
}

// runGitHubHealthOperation validates GitHub OAuth credentials.
func runGitHubHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := githubRESTClientForOperation(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return providerkit.OperationFailure("GitHub user lookup failed", normalizeGitHubAPIError(err), nil)
	}

	login := user.GetLogin()
	details := githubHealthDetails{
		Login: login,
		ID:    user.GetID(),
		Name:  user.GetName(),
	}

	return providerkit.OperationSuccess(fmt.Sprintf("GitHub token valid for %s", login), details), nil
}

// runGitHubAppHealthOperation validates GitHub App installation tokens.
func runGitHubAppHealthOperation(baseURL string) types.OperationFunc {
	return func(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
		client, err := githubRESTClientForOperationWithBaseURL(ctx, input, baseURL)
		if err != nil {
			return types.OperationResult{}, err
		}

		repositories, _, err := client.Apps.ListRepos(ctx, &gh.ListOptions{Page: 1, PerPage: 1})
		if err != nil {
			return providerkit.OperationFailure("GitHub App installation lookup failed", normalizeGitHubAPIError(err), nil)
		}

		count := 0
		if repositories != nil {
			count = repositories.GetTotalCount()
			if count == 0 {
				count = len(repositories.Repositories)
			}
		}

		return providerkit.OperationSuccess(fmt.Sprintf("GitHub App token valid for %d repositories", count), githubAppHealthDetails{Count: count}), nil
	}
}

// runGitHubRepoOperation lists repositories for the authenticated account.
func runGitHubRepoOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := githubRESTClientForOperation(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	var repoConfig githubRepoOperationConfig
	if err := jsonx.UnmarshalIfPresent(input.Config, &repoConfig); err != nil {
		return types.OperationResult{}, err
	}

	config := githubVulnerabilityConfig{
		Pagination: providerkit.Pagination{PerPage: repoConfig.PerPage},
		Visibility: repoConfig.Visibility,
	}

	repos, err := listGitHubReposForProvider(ctx, client, input.Provider, config)
	if err != nil {
		return providerkit.OperationFailure("GitHub repository collection failed", err, nil)
	}

	sampleSize := min(len(repos), providerkit.DefaultSampleSize)
	samples := lo.Map(repos[:sampleSize], func(repo *gh.Repository, _ int) githubRepoSample {
		if repo == nil {
			return githubRepoSample{}
		}

		return githubRepoSample{
			Name:      repo.GetName(),
			Private:   repo.GetPrivate(),
			UpdatedAt: repo.GetUpdatedAt().Time,
			URL:       repo.GetHTMLURL(),
		}
	})

	return providerkit.OperationSuccess(fmt.Sprintf("Collected %d repositories", len(repos)), githubRepoCollectionDetails{
		Count:   len(repos),
		Samples: samples,
	}), nil
}

// normalizeGitHubAPIError maps GitHub SDK HTTP errors to provider-level API errors.
func normalizeGitHubAPIError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *gh.ErrorResponse
	if errors.As(err, &apiErr) {
		return ErrAPIRequest
	}

	var rateErr *gh.RateLimitError
	if errors.As(err, &rateErr) {
		return ErrAPIRequest
	}

	var abuseErr *gh.AbuseRateLimitError
	if errors.As(err, &abuseErr) {
		return ErrAPIRequest
	}

	return err
}

// clampPerPage bounds the per-page value for GitHub API requests.
func clampPerPage(value int) int {
	switch {
	case value <= 0:
		return defaultPerPage
	case value > maxPerPage:
		return maxPerPage
	default:
		return value
	}
}
