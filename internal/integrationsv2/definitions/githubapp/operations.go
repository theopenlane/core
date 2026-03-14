package githubapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	gh "github.com/google/go-github/v83/github"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultPerPage = 50
	maxPerPage     = 100
	defaultAlertState = "open"

	githubAlertTypeDependabot     = "dependabot"
	githubAlertTypeCodeScanning   = "code_scanning"
	githubAlertTypeSecretScanning = "secret_scanning"
	githubRepositoryNameParts     = 2
)

// healthDetails is the JSON payload returned by the health operation
type healthDetails struct {
	Count int `json:"count"`
}

// repoSyncDetails is the JSON payload returned by the repository sync operation
type repoSyncDetails struct {
	Count   int          `json:"count"`
	Samples []repoSample `json:"samples"`
}

type repoSample struct {
	Name      string    `json:"name"`
	Private   bool      `json:"private"`
	UpdatedAt time.Time `json:"updated_at"`
	URL       string    `json:"url"`
}

// vulnerabilityDetails is the JSON payload returned by the vulnerability collect operation
type vulnerabilityDetails struct {
	RepositoriesScanned int            `json:"repositories_scanned"`
	AlertsTotal         int            `json:"alerts_total"`
	AlertTypeCounts     map[string]int `json:"alert_type_counts,omitempty"`
}

// vulnerabilityOperationConfig controls the vulnerability collect operation
type vulnerabilityOperationConfig struct {
	// AlertTypes selects which alert types to collect
	AlertTypes []string `json:"alert_types,omitempty" jsonschema:"description=Alert types to collect (dependabot, code_scanning, secret_scanning). Defaults to all."`
	// MaxRepos caps the number of repositories to scan
	MaxRepos int `json:"max_repos,omitempty" jsonschema:"description=Optional cap on the number of repositories to scan."`
	// AlertState filters Dependabot alert state
	AlertState string `json:"alert_state,omitempty" jsonschema:"description=Dependabot alert state filter (open, dismissed, fixed, all). Defaults to open."`
	// Severity filters alerts by severity
	Severity string `json:"severity,omitempty" jsonschema:"description=Optional severity filter (low, medium, high, critical)."`
}

// runHealthOperation validates the GitHub App installation token by calling Apps.ListRepos
func (d *def) runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	ghClient, err := restClientFromAny(client)
	if err != nil {
		return nil, err
	}

	repositories, _, err := ghClient.Apps.ListRepos(ctx, &gh.ListOptions{Page: 1, PerPage: 1})
	if err != nil {
		return nil, normalizeGitHubAPIError(err)
	}

	count := 0
	if repositories != nil {
		count = repositories.GetTotalCount()
		if count == 0 {
			count = len(repositories.Repositories)
		}
	}

	return jsonx.ToRawMessage(healthDetails{Count: count})
}

// runRepositorySyncOperation lists all repositories accessible to the GitHub App installation
func (d *def) runRepositorySyncOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	ghClient, err := restClientFromAny(client)
	if err != nil {
		return nil, err
	}

	repos, err := listInstallationRepos(ctx, ghClient, defaultPerPage)
	if err != nil {
		return nil, fmt.Errorf("github app repository sync failed: %w", err)
	}

	sampleSize := min(len(repos), 10)
	samples := lo.Map(repos[:sampleSize], func(repo *gh.Repository, _ int) repoSample {
		if repo == nil {
			return repoSample{}
		}

		return repoSample{
			Name:      repo.GetFullName(),
			Private:   repo.GetPrivate(),
			UpdatedAt: repo.GetUpdatedAt().Time,
			URL:       repo.GetHTMLURL(),
		}
	})

	return jsonx.ToRawMessage(repoSyncDetails{
		Count:   len(repos),
		Samples: samples,
	})
}

// runVulnerabilityCollectionOperation collects GitHub vulnerability alerts from all installation repos
func (d *def) runVulnerabilityCollectionOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, config json.RawMessage) (json.RawMessage, error) {
	ghClient, err := restClientFromAny(client)
	if err != nil {
		return nil, err
	}

	var cfg vulnerabilityOperationConfig
	if err := jsonx.UnmarshalIfPresent(config, &cfg); err != nil {
		return nil, err
	}

	repos, err := listInstallationRepos(ctx, ghClient, defaultPerPage)
	if err != nil {
		return nil, fmt.Errorf("github app repository listing failed: %w", err)
	}

	repoNames := lo.FilterMap(repos, func(repo *gh.Repository, _ int) (string, bool) {
		if repo == nil {
			return "", false
		}

		name := strings.TrimSpace(repo.GetFullName())
		return name, name != ""
	})

	if maxRepos := cfg.MaxRepos; maxRepos > 0 && len(repoNames) > maxRepos {
		repoNames = repoNames[:maxRepos]
	}

	alertTypes := resolveAlertTypes(cfg.AlertTypes)

	var (
		totalAlerts     int
		alertTypeCounts = map[string]int{}
	)

	for _, repo := range repoNames {
		if alertTypeRequested(alertTypes, githubAlertTypeDependabot) {
			batch, err := listDependabotAlerts(ctx, ghClient, repo, cfg)
			if err != nil {
				return nil, fmt.Errorf("github app dependabot alert collection failed for %s: %w", repo, err)
			}

			totalAlerts += len(batch)
			alertTypeCounts[githubAlertTypeDependabot] += len(batch)
		}

		if alertTypeRequested(alertTypes, githubAlertTypeCodeScanning) {
			batch, err := listCodeScanningAlerts(ctx, ghClient, repo)
			if err != nil {
				return nil, fmt.Errorf("github app code scanning alert collection failed for %s: %w", repo, err)
			}

			totalAlerts += len(batch)
			alertTypeCounts[githubAlertTypeCodeScanning] += len(batch)
		}

		if alertTypeRequested(alertTypes, githubAlertTypeSecretScanning) {
			batch, err := listSecretScanningAlerts(ctx, ghClient, repo, cfg)
			if err != nil {
				return nil, fmt.Errorf("github app secret scanning alert collection failed for %s: %w", repo, err)
			}

			totalAlerts += len(batch)
			alertTypeCounts[githubAlertTypeSecretScanning] += len(batch)
		}
	}

	return jsonx.ToRawMessage(vulnerabilityDetails{
		RepositoriesScanned: len(repoNames),
		AlertsTotal:         totalAlerts,
		AlertTypeCounts:     alertTypeCounts,
	})
}

// listInstallationRepos paginates all repositories accessible to the GitHub App installation
func listInstallationRepos(ctx context.Context, client *gh.Client, perPage int) ([]*gh.Repository, error) {
	perPage = clampPerPage(perPage)
	out := make([]*gh.Repository, 0)

	page := 1

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		response, httpResponse, err := client.Apps.ListRepos(ctx, &gh.ListOptions{Page: page, PerPage: perPage})
		if err != nil {
			return nil, normalizeGitHubAPIError(err)
		}

		if response != nil {
			out = append(out, response.Repositories...)
		}

		if httpResponse == nil || httpResponse.NextPage == 0 {
			break
		}

		page = httpResponse.NextPage
	}

	return out, nil
}

// listDependabotAlerts fetches Dependabot alerts for a repository
func listDependabotAlerts(ctx context.Context, client *gh.Client, repo string, cfg vulnerabilityOperationConfig) ([]*gh.DependabotAlert, error) {
	owner, repository, err := splitRepository(repo)
	if err != nil {
		return nil, err
	}

	state := lo.CoalesceOrEmpty(cfg.AlertState, defaultAlertState)
	perPage := defaultPerPage
	out := make([]*gh.DependabotAlert, 0)
	page := 1

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		opts := &gh.ListAlertsOptions{
			ListOptions: gh.ListOptions{Page: page, PerPage: perPage},
		}

		if state != "" {
			opts.State = lo.ToPtr(state)
		}

		if cfg.Severity != "" {
			opts.Severity = lo.ToPtr(cfg.Severity)
		}

		batch, response, err := client.Dependabot.ListRepoAlerts(ctx, owner, repository, opts)
		if err != nil {
			return nil, normalizeGitHubAPIError(err)
		}

		out = append(out, batch...)

		if response == nil || response.NextPage == 0 {
			break
		}

		page = response.NextPage
	}

	return out, nil
}

// listCodeScanningAlerts fetches code scanning alerts for a repository
func listCodeScanningAlerts(ctx context.Context, client *gh.Client, repo string) ([]*gh.Alert, error) {
	owner, repository, err := splitRepository(repo)
	if err != nil {
		return nil, err
	}

	out := make([]*gh.Alert, 0)
	page := 1

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		opts := &gh.AlertListOptions{
			State:       defaultAlertState,
			ListOptions: gh.ListOptions{Page: page, PerPage: defaultPerPage},
		}

		batch, response, err := client.CodeScanning.ListAlertsForRepo(ctx, owner, repository, opts)
		if err != nil {
			return nil, normalizeGitHubAPIError(err)
		}

		out = append(out, batch...)

		if response == nil || response.NextPage == 0 {
			break
		}

		page = response.NextPage
	}

	return out, nil
}

// listSecretScanningAlerts fetches secret scanning alerts for a repository
func listSecretScanningAlerts(ctx context.Context, client *gh.Client, repo string, cfg vulnerabilityOperationConfig) ([]*gh.SecretScanningAlert, error) {
	owner, repository, err := splitRepository(repo)
	if err != nil {
		return nil, err
	}

	state := lo.CoalesceOrEmpty(cfg.AlertState, defaultAlertState)
	out := make([]*gh.SecretScanningAlert, 0)
	page := 1

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		opts := &gh.SecretScanningAlertListOptions{
			State:       state,
			ListOptions: gh.ListOptions{Page: page, PerPage: defaultPerPage},
		}

		batch, response, err := client.SecretScanning.ListAlertsForRepo(ctx, owner, repository, opts)
		if err != nil {
			return nil, normalizeGitHubAPIError(err)
		}

		out = append(out, batch...)

		if response == nil || response.NextPage == 0 {
			break
		}

		page = response.NextPage
	}

	return out, nil
}

// splitRepository parses owner/repo repository names
func splitRepository(value string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(value), "/", githubRepositoryNameParts)
	if len(parts) != githubRepositoryNameParts || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%w: %q", ErrRepositoryInvalid, value)
	}

	return parts[0], parts[1], nil
}

// clampPerPage bounds the per-page value for GitHub API requests
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

// defaultAlertTypes lists every alert category collected when none is specified
var defaultAlertTypes = []string{githubAlertTypeDependabot, githubAlertTypeCodeScanning, githubAlertTypeSecretScanning}

// resolveAlertTypes normalizes and defaults the requested alert types
func resolveAlertTypes(values []string) []string {
	normalized := lo.Uniq(lo.Filter(values, func(v string, _ int) bool {
		return strings.TrimSpace(v) != ""
	}))

	if len(normalized) == 0 {
		return defaultAlertTypes
	}

	return normalized
}

// alertTypeRequested checks whether a specific alert type should be fetched
func alertTypeRequested(alertTypes []string, target string) bool {
	if len(alertTypes) == 0 {
		return true
	}

	return lo.Contains(alertTypes, target)
}

// normalizeGitHubAPIError maps GitHub SDK HTTP errors to package-level errors
func normalizeGitHubAPIError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *gh.ErrorResponse
	if errors.As(err, &apiErr) {
		return fmt.Errorf("%w: %s", ErrAPIRequest, apiErr.Message)
	}

	var rateErr *gh.RateLimitError
	if errors.As(err, &rateErr) {
		return fmt.Errorf("%w: rate limited", ErrAPIRequest)
	}

	var abuseErr *gh.AbuseRateLimitError
	if errors.As(err, &abuseErr) {
		return fmt.Errorf("%w: abuse rate limit", ErrAPIRequest)
	}

	return err
}
