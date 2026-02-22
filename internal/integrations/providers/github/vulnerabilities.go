package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	gh "github.com/google/go-github/v83/github"
	"github.com/samber/lo"
	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	githubAlertTypeDependabot     = "dependabot"
	githubAlertTypeCodeScanning   = "code_scanning"
	githubAlertTypeSecretScanning = "secret_scanning"
)

type githubInstallationRepositoriesResponse struct {
	// TotalCount is the number of repositories returned by GitHub
	TotalCount int `json:"total_count"`
	// Repositories lists the repositories in the response
	Repositories []githubRepoResponse `json:"repositories"`
}

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

// runGitHubVulnerabilityOperation collects GitHub alert data and returns envelope payloads
func runGitHubVulnerabilityOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndOAuthToken(input)
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
		repos, err := listGitHubReposForProvider(ctx, client, token, input.Provider, config)
		if err != nil {
			return operations.OperationFailure("GitHub repository listing failed", err, nil)
		}
		repoNames = repoNamesFromResponses(repos, config.Owner.String())
	}

	if len(repoNames) == 0 {
		return types.OperationResult{
			Status:  types.OperationStatusOK,
			Summary: "No repositories available for vulnerability alerts",
			Details: map[string]any{
				"repositories": 0,
				"alerts":       0,
			},
		}, nil
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
			batch, err := listDependabotAlerts(ctx, client, token, repo, config)
			if err != nil {
				return operations.OperationFailure("GitHub Dependabot alert collection failed", err, map[string]any{
					"repository": repo,
				})
			}
			envelopes = appendAlertEnvelopes(envelopes, githubAlertTypeDependabot, repo, batch)
			totalAlerts += len(batch)
			alertTypeCounts[githubAlertTypeDependabot] += len(batch)
		}

		if alertTypeRequested(alertTypes, githubAlertTypeCodeScanning) {
			batch, err := listCodeScanningAlerts(ctx, client, token, repo, config)
			if err != nil {
				return operations.OperationFailure("GitHub code scanning alert collection failed", err, map[string]any{
					"repository": repo,
				})
			}
			envelopes = appendAlertEnvelopes(envelopes, githubAlertTypeCodeScanning, repo, batch)
			totalAlerts += len(batch)
			alertTypeCounts[githubAlertTypeCodeScanning] += len(batch)
		}

		if alertTypeRequested(alertTypes, githubAlertTypeSecretScanning) {
			batch, err := listSecretScanningAlerts(ctx, client, token, repo, config)
			if err != nil {
				return operations.OperationFailure("GitHub secret scanning alert collection failed", err, map[string]any{
					"repository": repo,
				})
			}
			envelopes = appendAlertEnvelopes(envelopes, githubAlertTypeSecretScanning, repo, batch)
			totalAlerts += len(batch)
			alertTypeCounts[githubAlertTypeSecretScanning] += len(batch)
		}
	}

	details := map[string]any{
		"repositories_scanned": len(repoNames),
		"alerts_total":         totalAlerts,
		"alert_type_counts":    alertTypeCounts,
	}
	details = operations.AddPayloadIf(details, config.IncludePayloads, "alerts", envelopes)

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d vulnerability alerts from %d repositories", totalAlerts, len(repoNames)),
		Details: details,
	}, nil
}

// listGitHubReposForProvider enumerates repositories using either OAuth or app installation tokens
func listGitHubReposForProvider(ctx context.Context, client *auth.AuthenticatedClient, token string, provider types.ProviderType, config githubVulnerabilityConfig) ([]githubRepoResponse, error) {
	if provider == TypeGitHubApp {
		return listGitHubInstallationRepos(ctx, client, token, config)
	}

	return listGitHubRepos(ctx, client, token, config)
}

// listGitHubInstallationRepos lists repositories visible to a GitHub App installation
func listGitHubInstallationRepos(ctx context.Context, client *auth.AuthenticatedClient, token string, config githubVulnerabilityConfig) ([]githubRepoResponse, error) {
	perPage := clampPerPage(config.EffectivePageSize(defaultPerPage))
	out := make([]githubRepoResponse, 0)
	err := collectGitHubPaged(ctx, perPage, func(page, perPage int) ([]githubRepoResponse, error) {
		params := gitHubPageParams(page, perPage)
		var batch githubInstallationRepositoriesResponse
		if err := fetchGitHubResource(ctx, client, token, "installation/repositories", params, &batch); err != nil {
			return nil, err
		}
		return batch.Repositories, nil
	}, func(batch []githubRepoResponse) error {
		out = append(out, batch...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// listGitHubRepos lists repositories accessible to the OAuth token
func listGitHubRepos(ctx context.Context, client *auth.AuthenticatedClient, token string, config githubVulnerabilityConfig) ([]githubRepoResponse, error) {
	perPage := clampPerPage(config.EffectivePageSize(defaultPerPage))
	out := make([]githubRepoResponse, 0)
	err := collectGitHubPaged(ctx, perPage, func(page, perPage int) ([]githubRepoResponse, error) {
		params := gitHubPageParams(page, perPage)
		if visibility := config.Visibility.String(); visibility != "" {
			params.Set("visibility", visibility)
		}
		if affiliation := config.Affiliation.String(); affiliation != "" {
			params.Set("affiliation", affiliation)
		}

		var batch []githubRepoResponse
		if err := fetchGitHubResource(ctx, client, token, "user/repos", params, &batch); err != nil {
			return nil, err
		}

		return batch, nil
	}, func(batch []githubRepoResponse) error {
		out = append(out, batch...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// listDependabotAlerts fetches Dependabot alerts for a repository
func listDependabotAlerts(ctx context.Context, client *auth.AuthenticatedClient, token, repo string, config githubVulnerabilityConfig) ([]json.RawMessage, error) {
	path := fmt.Sprintf("repos/%s/dependabot/alerts", repo)
	return listGitHubAlerts[*gh.DependabotAlert](ctx, client, token, path, config, func(params url.Values) {
		if severity := config.Severity.String(); severity != "" {
			params.Set("severity", severity)
		}

		if ecosystem := config.Ecosystem.String(); ecosystem != "" {
			params.Set("ecosystem", ecosystem)
		}
	})
}

// listCodeScanningAlerts fetches code scanning alerts for a repository
func listCodeScanningAlerts(ctx context.Context, client *auth.AuthenticatedClient, token, repo string, config githubVulnerabilityConfig) ([]json.RawMessage, error) {
	path := fmt.Sprintf("repos/%s/code-scanning/alerts", repo)

	return listGitHubAlerts[*gh.Alert](ctx, client, token, path, config, nil)
}

// listSecretScanningAlerts fetches secret scanning alerts for a repository
func listSecretScanningAlerts(ctx context.Context, client *auth.AuthenticatedClient, token, repo string, config githubVulnerabilityConfig) ([]json.RawMessage, error) {
	path := fmt.Sprintf("repos/%s/secret-scanning/alerts", repo)

	return listGitHubAlerts[*gh.SecretScanningAlert](ctx, client, token, path, config, nil)
}

// listGitHubAlerts fetches and marshals GitHub alert payloads with pagination
func listGitHubAlerts[T any](ctx context.Context, client *auth.AuthenticatedClient, token, path string, config githubVulnerabilityConfig, decorate func(url.Values)) ([]json.RawMessage, error) {
	perPage := clampPerPage(config.EffectivePageSize(defaultPerPage))
	state := lo.CoalesceOrEmpty(config.AlertState.String(), defaultAlertState)
	out := make([]json.RawMessage, 0)

	err := collectGitHubPaged(ctx, perPage, func(page, perPage int) ([]T, error) {
		params := gitHubPageParams(page, perPage)
		if state != "" {
			params.Set("state", state)
		}

		if decorate != nil {
			decorate(params)
		}

		var batch []T
		if err := fetchGitHubResource(ctx, client, token, path, params, &batch); err != nil {
			return nil, err
		}

		return batch, nil
	}, func(batch []T) error {
		for _, alert := range batch {
			if lo.IsNil(alert) {
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

// collectGitHubPaged iterates through paged GitHub API responses
func collectGitHubPaged[T any](ctx context.Context, perPage int, fetch func(page, perPage int) ([]T, error), handle func([]T) error) error {
	page := 1
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		batch, err := fetch(page, perPage)
		if err != nil {
			return err
		}

		if err := handle(batch); err != nil {
			return err
		}

		if len(batch) < perPage {
			return nil
		}
		page++
	}
}

// gitHubPageParams builds query parameters for paged GitHub API requests
func gitHubPageParams(page, perPage int) url.Values {
	params := url.Values{}
	params.Set("per_page", fmt.Sprintf("%d", perPage))
	params.Set("page", fmt.Sprintf("%d", page))

	return params
}

// appendAlertEnvelopes wraps payloads into alert envelopes
func appendAlertEnvelopes(envelopes []types.AlertEnvelope, alertType, resource string, payloads []json.RawMessage) []types.AlertEnvelope {
	return append(envelopes, lo.Map(payloads, func(p json.RawMessage, _ int) types.AlertEnvelope {
		return types.AlertEnvelope{AlertType: alertType, Resource: resource, Payload: p}
	})...)
}

// repoNamesFromResponses builds full repo names from API responses
func repoNamesFromResponses(repos []githubRepoResponse, ownerFilter string) []string {
	return lo.FilterMap(repos, func(repo githubRepoResponse, _ int) (string, bool) {
		full := lo.CoalesceOrEmpty(repo.FullName, lo.Ternary(repo.Owner.Login != "", repo.Owner.Login+"/"+repo.Name, ""))
		if full == "" {
			return "", false
		}
		if ownerFilter == "" {
			return full, true
		}
		return full, strings.HasPrefix(full, ownerFilter+"/") || strings.EqualFold(repo.Owner.Login, ownerFilter)
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
