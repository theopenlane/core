package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v83/github"
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	githubAlertTypeDependabot     = "dependabot"
	githubAlertTypeCodeScanning   = "code_scanning"
	githubAlertTypeSecretScanning = "secret_scanning"
)

type githubVulnerabilityConfig struct {
	// RepositorySelector controls which repositories to scan
	operations.RepositorySelector
	// Pagination controls page sizing for list calls
	operations.Pagination
	// PayloadOptions controls payload inclusion
	operations.PayloadOptions

	// AlertTypes selects which alert types to collect
	AlertTypes []types.LowerString `json:"alert_types"`
	// MaxRepos caps the number of repositories to scan
	MaxRepos int `json:"max_repos"`
	// Visibility filters repositories by visibility
	Visibility types.LowerString `json:"visibility"`
	// Affiliation filters repositories by affiliation
	Affiliation types.LowerString `json:"affiliation"`
	// AlertState filters alert state for Dependabot alerts
	AlertState types.LowerString `json:"alert_state"`
	// Severity filters alerts by severity
	Severity types.LowerString `json:"severity"`
	// Ecosystem filters alerts by package ecosystem
	Ecosystem types.LowerString `json:"ecosystem"`
}

type githubVulnerabilityDetails struct {
	RepositoriesScanned int                   `json:"repositories_scanned"`
	AlertsTotal         int                   `json:"alerts_total"`
	AlertTypeCounts     map[string]int        `json:"alert_type_counts,omitempty"`
	Alerts              []types.AlertEnvelope `json:"alerts,omitempty"`
}

type githubRepositoryFailureDetails struct {
	Repository string `json:"repository,omitempty"`
}

// runGitHubVulnerabilityOperation collects GitHub alert data and returns envelope payloads
func runGitHubVulnerabilityOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := githubRESTClientForOperation(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	config, err := operations.Decode[githubVulnerabilityConfig](input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	alertTypes := alertTypesFromConfig(config.AlertTypes)

	repoNames := config.List()
	if len(repoNames) == 0 {
		repos, err := listGitHubReposForProvider(ctx, client, input.Provider, config)
		if err != nil {
			return operations.OperationFailure("GitHub repository listing failed", err, nil)
		}
		repoNames = repoNamesFromResponses(repos, config.Owner.String())
	}

	if len(repoNames) == 0 {
		return operations.OperationSuccess("No repositories available for vulnerability alerts", githubVulnerabilityDetails{
			RepositoriesScanned: 0,
			AlertsTotal:         0,
		}), nil
	}

	if maxRepos := config.MaxRepos; maxRepos > 0 && len(repoNames) > maxRepos {
		repoNames = repoNames[:maxRepos]
	}

	var (
		totalAlerts     int
		envelopes       []types.AlertEnvelope
		alertTypeCounts = map[string]int{}
	)

	for _, repo := range repoNames {
		if alertTypeRequested(alertTypes, githubAlertTypeDependabot) {
			batch, err := listDependabotAlerts(ctx, client, repo, config)
			if err != nil {
				return operations.OperationFailure("GitHub Dependabot alert collection failed", err, githubRepositoryFailureDetails{
					Repository: repo,
				})
			}
			envelopes = appendAlertEnvelopes(envelopes, githubAlertTypeDependabot, repo, batch)
			totalAlerts += len(batch)
			alertTypeCounts[githubAlertTypeDependabot] += len(batch)
		}

		if alertTypeRequested(alertTypes, githubAlertTypeCodeScanning) {
			batch, err := listCodeScanningAlerts(ctx, client, repo, config)
			if err != nil {
				return operations.OperationFailure("GitHub code scanning alert collection failed", err, githubRepositoryFailureDetails{
					Repository: repo,
				})
			}
			envelopes = appendAlertEnvelopes(envelopes, githubAlertTypeCodeScanning, repo, batch)
			totalAlerts += len(batch)
			alertTypeCounts[githubAlertTypeCodeScanning] += len(batch)
		}

		if alertTypeRequested(alertTypes, githubAlertTypeSecretScanning) {
			batch, err := listSecretScanningAlerts(ctx, client, repo, config)
			if err != nil {
				return operations.OperationFailure("GitHub secret scanning alert collection failed", err, githubRepositoryFailureDetails{
					Repository: repo,
				})
			}
			envelopes = appendAlertEnvelopes(envelopes, githubAlertTypeSecretScanning, repo, batch)
			totalAlerts += len(batch)
			alertTypeCounts[githubAlertTypeSecretScanning] += len(batch)
		}
	}

	details := githubVulnerabilityDetails{
		RepositoriesScanned: len(repoNames),
		AlertsTotal:         totalAlerts,
		AlertTypeCounts:     alertTypeCounts,
	}
	if config.IncludePayloads {
		details.Alerts = envelopes
	}

	return operations.OperationSuccess(fmt.Sprintf("Collected %d vulnerability alerts from %d repositories", totalAlerts, len(repoNames)), details), nil
}

// listGitHubReposForProvider enumerates repositories using either OAuth or app installation tokens
func listGitHubReposForProvider(ctx context.Context, client *gh.Client, provider types.ProviderType, config githubVulnerabilityConfig) ([]*gh.Repository, error) {
	if provider == TypeGitHubApp {
		return listGitHubInstallationRepos(ctx, client, config)
	}

	return listGitHubRepos(ctx, client, config)
}

// listGitHubInstallationRepos lists repositories visible to a GitHub App installation
func listGitHubInstallationRepos(ctx context.Context, client *gh.Client, config githubVulnerabilityConfig) ([]*gh.Repository, error) {
	perPage := clampPerPage(config.EffectivePageSize(defaultPerPage))
	out := make([]*gh.Repository, 0)

	err := collectGitHubPaged(ctx, perPage, func(page, perPage int) ([]*gh.Repository, *gh.Response, error) {
		response, httpResponse, err := client.Apps.ListRepos(ctx, &gh.ListOptions{Page: page, PerPage: perPage})
		if err != nil {
			return nil, nil, normalizeGitHubAPIError(err)
		}
		if response == nil {
			return nil, httpResponse, nil
		}

		return response.Repositories, httpResponse, nil
	}, func(batch []*gh.Repository) error {
		out = append(out, batch...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// listGitHubRepos lists repositories accessible to the OAuth token
func listGitHubRepos(ctx context.Context, client *gh.Client, config githubVulnerabilityConfig) ([]*gh.Repository, error) {
	perPage := clampPerPage(config.EffectivePageSize(defaultPerPage))
	out := make([]*gh.Repository, 0)

	err := collectGitHubPaged(ctx, perPage, func(page, perPage int) ([]*gh.Repository, *gh.Response, error) {
		opts := &gh.RepositoryListByAuthenticatedUserOptions{
			ListOptions: gh.ListOptions{Page: page, PerPage: perPage},
		}
		if visibility := config.Visibility.String(); visibility != "" {
			opts.Visibility = visibility
		}
		if affiliation := config.Affiliation.String(); affiliation != "" {
			opts.Affiliation = affiliation
		}

		batch, response, err := client.Repositories.ListByAuthenticatedUser(ctx, opts)
		if err != nil {
			return nil, nil, normalizeGitHubAPIError(err)
		}

		return batch, response, nil
	}, func(batch []*gh.Repository) error {
		out = append(out, batch...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// listDependabotAlerts fetches Dependabot alerts for a repository
func listDependabotAlerts(ctx context.Context, client *gh.Client, repo string, config githubVulnerabilityConfig) ([]json.RawMessage, error) {
	owner, repository, err := splitGitHubRepository(repo)
	if err != nil {
		return nil, err
	}

	perPage := clampPerPage(config.EffectivePageSize(defaultPerPage))
	state := lo.CoalesceOrEmpty(config.AlertState.String(), defaultAlertState)
	out := make([]json.RawMessage, 0)

	err = collectGitHubPaged(ctx, perPage, func(page, perPage int) ([]*gh.DependabotAlert, *gh.Response, error) {
		opts := &gh.ListAlertsOptions{
			ListOptions: gh.ListOptions{Page: page, PerPage: perPage},
		}
		if state != "" {
			opts.State = lo.ToPtr(state)
		}
		if severity := config.Severity.String(); severity != "" {
			opts.Severity = lo.ToPtr(severity)
		}
		if ecosystem := config.Ecosystem.String(); ecosystem != "" {
			opts.Ecosystem = lo.ToPtr(ecosystem)
		}

		batch, response, err := client.Dependabot.ListRepoAlerts(ctx, owner, repository, opts)
		if err != nil {
			return nil, nil, normalizeGitHubAPIError(err)
		}

		return batch, response, nil
	}, func(batch []*gh.DependabotAlert) error {
		for _, alert := range batch {
			if alert == nil {
				continue
			}

			payload, err := json.Marshal(alert)
			if err != nil {
				return err
			}

			out = append(out, payload)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// listCodeScanningAlerts fetches code scanning alerts for a repository
func listCodeScanningAlerts(ctx context.Context, client *gh.Client, repo string, config githubVulnerabilityConfig) ([]json.RawMessage, error) {
	owner, repository, err := splitGitHubRepository(repo)
	if err != nil {
		return nil, err
	}

	perPage := clampPerPage(config.EffectivePageSize(defaultPerPage))
	state := lo.CoalesceOrEmpty(config.AlertState.String(), defaultAlertState)
	out := make([]json.RawMessage, 0)

	err = collectGitHubPaged(ctx, perPage, func(page, perPage int) ([]*gh.Alert, *gh.Response, error) {
		opts := &gh.AlertListOptions{
			State:       state,
			ListOptions: gh.ListOptions{Page: page, PerPage: perPage},
		}

		batch, response, err := client.CodeScanning.ListAlertsForRepo(ctx, owner, repository, opts)
		if err != nil {
			return nil, nil, normalizeGitHubAPIError(err)
		}

		return batch, response, nil
	}, func(batch []*gh.Alert) error {
		for _, alert := range batch {
			if alert == nil {
				continue
			}

			payload, err := json.Marshal(alert)
			if err != nil {
				return err
			}

			out = append(out, payload)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// listSecretScanningAlerts fetches secret scanning alerts for a repository
func listSecretScanningAlerts(ctx context.Context, client *gh.Client, repo string, config githubVulnerabilityConfig) ([]json.RawMessage, error) {
	owner, repository, err := splitGitHubRepository(repo)
	if err != nil {
		return nil, err
	}

	perPage := clampPerPage(config.EffectivePageSize(defaultPerPage))
	state := lo.CoalesceOrEmpty(config.AlertState.String(), defaultAlertState)
	out := make([]json.RawMessage, 0)

	err = collectGitHubPaged(ctx, perPage, func(page, perPage int) ([]*gh.SecretScanningAlert, *gh.Response, error) {
		opts := &gh.SecretScanningAlertListOptions{
			State:       state,
			ListOptions: gh.ListOptions{Page: page, PerPage: perPage},
		}

		batch, response, err := client.SecretScanning.ListAlertsForRepo(ctx, owner, repository, opts)
		if err != nil {
			return nil, nil, normalizeGitHubAPIError(err)
		}

		return batch, response, nil
	}, func(batch []*gh.SecretScanningAlert) error {
		for _, alert := range batch {
			if alert == nil {
				continue
			}

			payload, err := json.Marshal(alert)
			if err != nil {
				return err
			}

			out = append(out, payload)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// splitGitHubRepository parses owner/repo repository names.
func splitGitHubRepository(value string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(value), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%w: %q", ErrRepositoryInvalid, value)
	}

	return parts[0], parts[1], nil
}

// collectGitHubPaged iterates through paged GitHub API responses.
func collectGitHubPaged[T any](ctx context.Context, perPage int, fetch func(page, perPage int) ([]T, *gh.Response, error), handle func([]T) error) error {
	page := 1
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		batch, response, err := fetch(page, perPage)
		if err != nil {
			return err
		}

		if err := handle(batch); err != nil {
			return err
		}

		if response == nil || response.NextPage == 0 {
			return nil
		}

		page = response.NextPage
	}
}

// appendAlertEnvelopes wraps payloads into alert envelopes
func appendAlertEnvelopes(envelopes []types.AlertEnvelope, alertType, resource string, payloads []json.RawMessage) []types.AlertEnvelope {
	return append(envelopes, lo.Map(payloads, func(p json.RawMessage, _ int) types.AlertEnvelope {
		return types.AlertEnvelope{AlertType: alertType, Resource: resource, Payload: p}
	})...)
}

// repoNamesFromResponses builds full repo names from API responses
func repoNamesFromResponses(repos []*gh.Repository, ownerFilter string) []string {
	return lo.FilterMap(repos, func(repo *gh.Repository, _ int) (string, bool) {
		if repo == nil {
			return "", false
		}

		full := strings.TrimSpace(repo.GetFullName())
		if full == "" {
			owner := ""
			if ownerUser := repo.GetOwner(); ownerUser != nil {
				owner = strings.TrimSpace(ownerUser.GetLogin())
			}

			name := strings.TrimSpace(repo.GetName())
			if owner != "" && name != "" {
				full = owner + "/" + name
			}
		}

		if full == "" {
			return "", false
		}
		if ownerFilter == "" {
			return full, true
		}

		repoOwner := ""
		if ownerUser := repo.GetOwner(); ownerUser != nil {
			repoOwner = ownerUser.GetLogin()
		}

		return full, strings.HasPrefix(full, ownerFilter+"/") || strings.EqualFold(repoOwner, ownerFilter)
	})
}

// defaultAlertTypes lists every alert category collected when none is specified.
var defaultAlertTypes = []string{githubAlertTypeDependabot, githubAlertTypeCodeScanning, githubAlertTypeSecretScanning}

// alertTypesFromConfig normalizes and defaults the requested alert types
func alertTypesFromConfig(values []types.LowerString) []string {
	out := lo.Uniq(lo.FilterMap(values, func(value types.LowerString, _ int) (string, bool) {
		normalized := normalizeAlertType(value.String())
		return normalized, normalized != ""
	}))

	if len(out) == 0 {
		return defaultAlertTypes
	}

	return out
}

// alertTypeRequested checks whether a specific alert type should be fetched
func alertTypeRequested(alertTypes []string, target string) bool {
	if len(alertTypes) == 0 {
		return true
	}

	needle := normalizeAlertType(target)
	if needle == "" {
		return false
	}

	return lo.ContainsBy(alertTypes, func(value string) bool {
		return normalizeAlertType(value) == needle
	})
}

// normalizeAlertType standardizes alert type identifiers
func normalizeAlertType(value string) string {
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	switch value {
	case "dependabot_alerts":
		return githubAlertTypeDependabot
	case "code_scanning_alerts":
		return githubAlertTypeCodeScanning
	case "secret_scanning_alerts":
		return githubAlertTypeSecretScanning
	}

	return value
}
