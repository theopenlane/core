package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	gh "github.com/google/go-github/v80/github"
	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	githubAlertTypeDependabot     = "dependabot"
	githubAlertTypeCodeScanning   = "code_scanning"
	githubAlertTypeSecretScanning = "secret_scanning"
)

type githubInstallationRepositoriesResponse struct {
	TotalCount   int                  `json:"total_count"`
	Repositories []githubRepoResponse `json:"repositories"`
}

// runGitHubVulnerabilityOperation collects GitHub alert data and returns envelope payloads.
func runGitHubVulnerabilityOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
	token, err := oauthTokenFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	alertTypes := alertTypesFromConfig(input.Config)

	repoNames := repositoryListFromConfig(input.Config)
	if len(repoNames) == 0 {
		repos, err := listGitHubReposForProvider(ctx, client, token, input.Provider, input.Config)
		if err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "GitHub repository listing failed",
				Details: map[string]any{"error": err.Error()},
			}, err
		}
		repoNames = repoNamesFromResponses(repos, helpers.ConfigString(input.Config, "owner"))
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

	if maxRepos := helpers.ConfigInt(input.Config, "max_repos", 0); maxRepos > 0 && len(repoNames) > maxRepos {
		repoNames = repoNames[:maxRepos]
	}

	var (
		totalAlerts     int
		envelopes       []types.AlertEnvelope
		severityCounts  = map[string]int{}
		alertTypeCounts = map[string]int{}
		samples         []map[string]any
	)

	for _, repo := range repoNames {
		if alertTypeRequested(alertTypes, githubAlertTypeDependabot) {
			batch, err := listDependabotAlerts(ctx, client, token, repo, input.Config)
			if err != nil {
				return types.OperationResult{
					Status:  types.OperationStatusFailed,
					Summary: "GitHub Dependabot alert collection failed",
					Details: map[string]any{
						"repository": repo,
						"error":      err.Error(),
					},
				}, err
			}

			for _, alert := range batch {
				if alert == nil {
					continue
				}
				payload, err := json.Marshal(alert)
				if err != nil {
					return types.OperationResult{
						Status:  types.OperationStatusFailed,
						Summary: "GitHub Dependabot alert serialization failed",
						Details: map[string]any{
							"repository": repo,
							"error":      err.Error(),
						},
					}, err
				}
				envelopes = append(envelopes, types.AlertEnvelope{
					AlertType: githubAlertTypeDependabot,
					Resource:  repo,
					Payload:   payload,
				})
				totalAlerts++
				alertTypeCounts[githubAlertTypeDependabot]++
				if severity := strings.TrimSpace(dependabotAlertSeverity(alert)); severity != "" {
					severityCounts[strings.ToLower(severity)]++
				}
				if len(samples) < maxSampleSize {
					samples = append(samples, map[string]any{
						"external_id": formatDependabotExternalID(repo, alert),
						"severity":    dependabotAlertSeverity(alert),
						"summary":     dependabotAlertSummary(alert),
						"owner":       repo,
					})
				}
			}
		}

		if alertTypeRequested(alertTypes, githubAlertTypeCodeScanning) {
			batch, err := listCodeScanningAlerts(ctx, client, token, repo, input.Config)
			if err != nil {
				return types.OperationResult{
					Status:  types.OperationStatusFailed,
					Summary: "GitHub code scanning alert collection failed",
					Details: map[string]any{
						"repository": repo,
						"error":      err.Error(),
					},
				}, err
			}

			for _, alert := range batch {
				if alert == nil {
					continue
				}
				payload, err := json.Marshal(alert)
				if err != nil {
					return types.OperationResult{
						Status:  types.OperationStatusFailed,
						Summary: "GitHub code scanning alert serialization failed",
						Details: map[string]any{
							"repository": repo,
							"error":      err.Error(),
						},
					}, err
				}
				envelopes = append(envelopes, types.AlertEnvelope{
					AlertType: githubAlertTypeCodeScanning,
					Resource:  repo,
					Payload:   payload,
				})
				totalAlerts++
				alertTypeCounts[githubAlertTypeCodeScanning]++
				if severity := strings.TrimSpace(codeScanningAlertSeverity(alert)); severity != "" {
					severityCounts[strings.ToLower(severity)]++
				}
				if len(samples) < maxSampleSize {
					samples = append(samples, map[string]any{
						"external_id": formatCodeScanningExternalID(repo, alert),
						"severity":    codeScanningAlertSeverity(alert),
						"summary":     codeScanningAlertSummary(alert),
						"owner":       repo,
					})
				}
			}
		}

		if alertTypeRequested(alertTypes, githubAlertTypeSecretScanning) {
			batch, err := listSecretScanningAlerts(ctx, client, token, repo, input.Config)
			if err != nil {
				return types.OperationResult{
					Status:  types.OperationStatusFailed,
					Summary: "GitHub secret scanning alert collection failed",
					Details: map[string]any{
						"repository": repo,
						"error":      err.Error(),
					},
				}, err
			}

			for _, alert := range batch {
				if alert == nil {
					continue
				}
				payload, err := json.Marshal(alert)
				if err != nil {
					return types.OperationResult{
						Status:  types.OperationStatusFailed,
						Summary: "GitHub secret scanning alert serialization failed",
						Details: map[string]any{
							"repository": repo,
							"error":      err.Error(),
						},
					}, err
				}
				envelopes = append(envelopes, types.AlertEnvelope{
					AlertType: githubAlertTypeSecretScanning,
					Resource:  repo,
					Payload:   payload,
				})
				totalAlerts++
				alertTypeCounts[githubAlertTypeSecretScanning]++
				if len(samples) < maxSampleSize {
					samples = append(samples, map[string]any{
						"external_id": formatSecretScanningExternalID(repo, alert),
						"summary":     secretScanningAlertSummary(alert),
						"owner":       repo,
					})
				}
			}
		}
	}

	details := map[string]any{
		"repositories_scanned": len(repoNames),
		"alerts_total":         totalAlerts,
		"severity_counts":      severityCounts,
		"alert_type_counts":    alertTypeCounts,
		"samples":              samples,
	}
	details = helpers.AddPayloadIf(details, helpers.ConfigBool(input.Config, "include_payloads", false), "alerts", envelopes)

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d vulnerability alerts from %d repositories", totalAlerts, len(repoNames)),
		Details: details,
	}, nil
}

// listGitHubReposForProvider enumerates repositories using either OAuth or app installation tokens.
func listGitHubReposForProvider(ctx context.Context, client *helpers.AuthenticatedClient, token string, provider types.ProviderType, config map[string]any) ([]githubRepoResponse, error) {
	if provider == TypeGitHubApp {
		return listGitHubInstallationRepos(ctx, client, token, config)
	}
	return listGitHubRepos(ctx, client, token, config)
}

// listGitHubInstallationRepos lists repositories visible to a GitHub App installation.
func listGitHubInstallationRepos(ctx context.Context, client *helpers.AuthenticatedClient, token string, config map[string]any) ([]githubRepoResponse, error) {
	perPage := clampPerPage(helpers.ConfigInt(config, "per_page", defaultPerPage))
	page := 1
	out := make([]githubRepoResponse, 0)

	for {
		params := url.Values{}
		params.Set("per_page", fmt.Sprintf("%d", perPage))
		params.Set("page", fmt.Sprintf("%d", page))

		var batch githubInstallationRepositoriesResponse
		if err := fetchGitHubResource(ctx, client, token, "installation/repositories", params, &batch); err != nil {
			return nil, err
		}

		out = append(out, batch.Repositories...)
		if len(batch.Repositories) < perPage {
			break
		}
		page++
	}

	return out, nil
}

// listGitHubRepos lists repositories accessible to the OAuth token.
func listGitHubRepos(ctx context.Context, client *helpers.AuthenticatedClient, token string, config map[string]any) ([]githubRepoResponse, error) {
	perPage := clampPerPage(helpers.ConfigInt(config, "per_page", defaultPerPage))
	page := 1
	out := make([]githubRepoResponse, 0)

	for {
		params := url.Values{}
		params.Set("per_page", fmt.Sprintf("%d", perPage))
		params.Set("page", fmt.Sprintf("%d", page))
		if visibility := helpers.ConfigString(config, "visibility"); visibility != "" {
			params.Set("visibility", visibility)
		}
		if affiliation := helpers.ConfigString(config, "affiliation"); affiliation != "" {
			params.Set("affiliation", affiliation)
		}

		var batch []githubRepoResponse
		if err := fetchGitHubResource(ctx, client, token, "user/repos", params, &batch); err != nil {
			return nil, err
		}

		out = append(out, batch...)
		if len(batch) < perPage {
			break
		}
		page++
	}

	return out, nil
}

// listDependabotAlerts fetches Dependabot alerts for a repository.
func listDependabotAlerts(ctx context.Context, client *helpers.AuthenticatedClient, token, repo string, config map[string]any) ([]*gh.DependabotAlert, error) {
	perPage := clampPerPage(helpers.ConfigInt(config, "per_page", defaultPerPage))
	page := 1
	out := make([]*gh.DependabotAlert, 0)

	state := helpers.ConfigString(config, "alert_state")
	if state == "" {
		state = defaultAlertState
	}

	for {
		params := url.Values{}
		params.Set("per_page", fmt.Sprintf("%d", perPage))
		params.Set("page", fmt.Sprintf("%d", page))
		if state != "" {
			params.Set("state", state)
		}
		if severity := helpers.ConfigString(config, "severity"); severity != "" {
			params.Set("severity", severity)
		}
		if ecosystem := helpers.ConfigString(config, "ecosystem"); ecosystem != "" {
			params.Set("ecosystem", ecosystem)
		}

		var batch []*gh.DependabotAlert
		path := fmt.Sprintf("repos/%s/dependabot/alerts", strings.TrimSpace(repo))
		if err := fetchGitHubResource(ctx, client, token, path, params, &batch); err != nil {
			return nil, err
		}

		out = append(out, batch...)
		if len(batch) < perPage {
			break
		}
		page++
	}

	return out, nil
}

// listCodeScanningAlerts fetches code scanning alerts for a repository.
func listCodeScanningAlerts(ctx context.Context, client *helpers.AuthenticatedClient, token, repo string, config map[string]any) ([]*gh.Alert, error) {
	perPage := clampPerPage(helpers.ConfigInt(config, "per_page", defaultPerPage))
	page := 1
	out := make([]*gh.Alert, 0)

	state := helpers.ConfigString(config, "alert_state")
	if state == "" {
		state = defaultAlertState
	}

	for {
		params := url.Values{}
		params.Set("per_page", fmt.Sprintf("%d", perPage))
		params.Set("page", fmt.Sprintf("%d", page))
		if state != "" {
			params.Set("state", state)
		}

		var batch []*gh.Alert
		path := fmt.Sprintf("repos/%s/code-scanning/alerts", strings.TrimSpace(repo))
		if err := fetchGitHubResource(ctx, client, token, path, params, &batch); err != nil {
			return nil, err
		}

		out = append(out, batch...)
		if len(batch) < perPage {
			break
		}
		page++
	}

	return out, nil
}

// listSecretScanningAlerts fetches secret scanning alerts for a repository.
func listSecretScanningAlerts(ctx context.Context, client *helpers.AuthenticatedClient, token, repo string, config map[string]any) ([]*gh.SecretScanningAlert, error) {
	perPage := clampPerPage(helpers.ConfigInt(config, "per_page", defaultPerPage))
	page := 1
	out := make([]*gh.SecretScanningAlert, 0)

	state := helpers.ConfigString(config, "alert_state")
	if state == "" {
		state = defaultAlertState
	}

	for {
		params := url.Values{}
		params.Set("per_page", fmt.Sprintf("%d", perPage))
		params.Set("page", fmt.Sprintf("%d", page))
		if state != "" {
			params.Set("state", state)
		}

		var batch []*gh.SecretScanningAlert
		path := fmt.Sprintf("repos/%s/secret-scanning/alerts", strings.TrimSpace(repo))
		if err := fetchGitHubResource(ctx, client, token, path, params, &batch); err != nil {
			return nil, err
		}

		out = append(out, batch...)
		if len(batch) < perPage {
			break
		}
		page++
	}

	return out, nil
}

// dependabotAlertSeverity extracts the severity label from a Dependabot alert.
func dependabotAlertSeverity(alert *gh.DependabotAlert) string {
	if alert == nil {
		return ""
	}
	advisory := alert.GetSecurityAdvisory()
	vulnerability := alert.GetSecurityVulnerability()
	severity := strings.TrimSpace(advisory.GetSeverity())
	if severity == "" {
		severity = strings.TrimSpace(vulnerability.GetSeverity())
	}
	return severity
}

// dependabotAlertSummary builds a concise summary for a Dependabot alert.
func dependabotAlertSummary(alert *gh.DependabotAlert) string {
	if alert == nil {
		return ""
	}
	advisory := alert.GetSecurityAdvisory()
	summary := strings.TrimSpace(advisory.GetSummary())
	if summary == "" {
		summary = strings.TrimSpace(advisory.GetGHSAID())
	}
	return summary
}

// codeScanningAlertSeverity extracts the severity label from a code scanning alert.
func codeScanningAlertSeverity(alert *gh.Alert) string {
	if alert == nil {
		return ""
	}
	rule := alert.GetRule()
	severity := strings.TrimSpace(rule.GetSecuritySeverityLevel())
	if severity == "" {
		severity = strings.TrimSpace(rule.GetSeverity())
	}
	return severity
}

// codeScanningAlertSummary builds a concise summary for a code scanning alert.
func codeScanningAlertSummary(alert *gh.Alert) string {
	if alert == nil {
		return ""
	}
	rule := alert.GetRule()
	instance := alert.GetMostRecentInstance()
	summary := strings.TrimSpace(rule.GetDescription())
	if summary == "" {
		summary = strings.TrimSpace(rule.GetName())
	}
	if summary == "" {
		summary = strings.TrimSpace(instance.GetMessage().GetText())
	}
	return summary
}

// secretScanningAlertSummary builds a concise summary for a secret scanning alert.
func secretScanningAlertSummary(alert *gh.SecretScanningAlert) string {
	if alert == nil {
		return ""
	}
	summary := strings.TrimSpace(alert.GetSecretTypeDisplayName())
	if summary == "" {
		summary = strings.TrimSpace(alert.GetSecretType())
	}
	return summary
}

// formatDependabotExternalID builds a stable external ID for a Dependabot alert.
func formatDependabotExternalID(repo string, alert *gh.DependabotAlert) string {
	if number := alert.GetNumber(); number != 0 {
		return fmt.Sprintf("github:%s:dependabot:%d", repo, number)
	}
	if ghsa := strings.TrimSpace(alert.GetSecurityAdvisory().GetGHSAID()); ghsa != "" {
		return fmt.Sprintf("github:%s:dependabot:%s", repo, ghsa)
	}
	return fmt.Sprintf("github:%s:dependabot:unknown", repo)
}

// formatCodeScanningExternalID builds a stable external ID for a code scanning alert.
func formatCodeScanningExternalID(repo string, alert *gh.Alert) string {
	if id := alert.ID(); id != 0 {
		return fmt.Sprintf("github:%s:code_scanning:%d", repo, id)
	}
	if number := alert.GetNumber(); number != 0 {
		return fmt.Sprintf("github:%s:code_scanning:%d", repo, number)
	}
	if ruleID := strings.TrimSpace(alert.GetRule().GetID()); ruleID != "" {
		return fmt.Sprintf("github:%s:code_scanning:%s", repo, ruleID)
	}
	return fmt.Sprintf("github:%s:code_scanning:unknown", repo)
}

// formatSecretScanningExternalID builds a stable external ID for a secret scanning alert.
func formatSecretScanningExternalID(repo string, alert *gh.SecretScanningAlert) string {
	if number := alert.GetNumber(); number != 0 {
		return fmt.Sprintf("github:%s:secret_scanning:%d", repo, number)
	}
	return fmt.Sprintf("github:%s:secret_scanning:unknown", repo)
}

func repositoryListFromConfig(config map[string]any) []string {
	values := helpers.ConfigStringSlice(config, "repositories")
	if len(values) == 0 {
		values = helpers.ConfigStringSlice(config, "repos")
	}
	if len(values) == 0 {
		if repo := helpers.ConfigString(config, "repository"); repo != "" {
			values = []string{repo}
		}
	}

	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func repoNamesFromResponses(repos []githubRepoResponse, ownerFilter string) []string {
	filter := strings.TrimSpace(ownerFilter)
	names := make([]string, 0, len(repos))

	for _, repo := range repos {
		full := strings.TrimSpace(repo.FullName)
		if full == "" && repo.Owner.Login != "" {
			full = strings.TrimSpace(repo.Owner.Login + "/" + repo.Name)
		}
		if full == "" {
			continue
		}
		if filter != "" {
			if strings.HasPrefix(full, filter+"/") || strings.EqualFold(repo.Owner.Login, filter) {
				names = append(names, full)
			}
			continue
		}
		names = append(names, full)
	}

	return names
}

func alertTypesFromConfig(config map[string]any) []string {
	values := helpers.ConfigStringSlice(config, "alert_types")
	if len(values) == 0 {
		values = helpers.ConfigStringSlice(config, "alertTypes")
	}
	if len(values) == 0 {
		return []string{githubAlertTypeDependabot, githubAlertTypeCodeScanning, githubAlertTypeSecretScanning}
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := normalizeAlertType(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return []string{githubAlertTypeDependabot, githubAlertTypeCodeScanning, githubAlertTypeSecretScanning}
	}
	return out
}

func alertTypeRequested(alertTypes []string, target string) bool {
	if len(alertTypes) == 0 {
		return true
	}
	needle := normalizeAlertType(target)
	if needle == "" {
		return false
	}
	for _, value := range alertTypes {
		if normalizeAlertType(value) == needle {
			return true
		}
	}
	return false
}

func normalizeAlertType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
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
