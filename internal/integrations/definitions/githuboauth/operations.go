package githuboauth

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
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultPerPage        = 50
	maxPerPage            = 100
	defaultAlertState     = "open"
	githubRepositoryParts = 2
)

type healthDetails struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Name  string `json:"name"`
}

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

type vulnerabilityDetails struct {
	RepositoriesScanned int            `json:"repositories_scanned"`
	AlertsTotal         int            `json:"alerts_total"`
	AlertTypeCounts     map[string]int `json:"alert_type_counts,omitempty"`
}

type vulnerabilityOperationConfig struct {
	AlertTypes []string `json:"alert_types,omitempty"`
	MaxRepos   int      `json:"max_repos,omitempty"`
	AlertState string   `json:"alert_state,omitempty"`
	Severity   string   `json:"severity,omitempty"`
}

const (
	alertTypeDependabot     = "dependabot"
	alertTypeCodeScanning   = "code_scanning"
	alertTypeSecretScanning = "secret_scanning"
)

// runHealthOperation validates the GitHub OAuth token via /user
func runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, err := restClientFromAny(client)
	if err != nil {
		return nil, err
	}

	user, _, err := c.Users.Get(ctx, "")
	if err != nil {
		return nil, normalizeGitHubAPIError(err)
	}

	return jsonx.ToRawMessage(healthDetails{
		Login: user.GetLogin(),
		ID:    user.GetID(),
		Name:  user.GetName(),
	})
}

// runRepositorySyncOperation lists repositories accessible to the OAuth token
func runRepositorySyncOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, err := restClientFromAny(client)
	if err != nil {
		return nil, err
	}

	repos, err := listUserRepos(ctx, c, defaultPerPage)
	if err != nil {
		return nil, fmt.Errorf("githuboauth: repository sync failed: %w", err)
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

	return jsonx.ToRawMessage(repoSyncDetails{Count: len(repos), Samples: samples})
}

// runVulnerabilityCollectOperation collects vulnerability alerts from accessible repos
func runVulnerabilityCollectOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, config json.RawMessage) (json.RawMessage, error) {
	c, err := restClientFromAny(client)
	if err != nil {
		return nil, err
	}

	var cfg vulnerabilityOperationConfig
	if err := jsonx.UnmarshalIfPresent(config, &cfg); err != nil {
		return nil, err
	}

	repos, err := listUserRepos(ctx, c, defaultPerPage)
	if err != nil {
		return nil, fmt.Errorf("githuboauth: repository listing failed: %w", err)
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
	var totalAlerts int
	alertTypeCounts := map[string]int{}

	for _, repo := range repoNames {
		owner, repository, err := splitRepository(repo)
		if err != nil {
			continue
		}

		if alertTypeRequested(alertTypes, alertTypeDependabot) {
			state := lo.CoalesceOrEmpty(cfg.AlertState, defaultAlertState)
			page := 1
			for {
				if err := ctx.Err(); err != nil {
					return nil, err
				}

				opts := &gh.ListAlertsOptions{
					ListOptions: gh.ListOptions{Page: page, PerPage: defaultPerPage},
				}
				if state != "" {
					opts.State = lo.ToPtr(state)
				}
				if cfg.Severity != "" {
					opts.Severity = lo.ToPtr(cfg.Severity)
				}

				batch, resp, err := c.Dependabot.ListRepoAlerts(ctx, owner, repository, opts)
				if err != nil {
					return nil, fmt.Errorf("githuboauth: dependabot alerts failed for %s: %w", repo, normalizeGitHubAPIError(err))
				}

				totalAlerts += len(batch)
				alertTypeCounts[alertTypeDependabot] += len(batch)

				if resp == nil || resp.NextPage == 0 {
					break
				}
				page = resp.NextPage
			}
		}

		if alertTypeRequested(alertTypes, alertTypeCodeScanning) {
			page := 1
			for {
				if err := ctx.Err(); err != nil {
					return nil, err
				}

				opts := &gh.AlertListOptions{
					State:       defaultAlertState,
					ListOptions: gh.ListOptions{Page: page, PerPage: defaultPerPage},
				}

				batch, resp, err := c.CodeScanning.ListAlertsForRepo(ctx, owner, repository, opts)
				if err != nil {
					return nil, fmt.Errorf("githuboauth: code scanning alerts failed for %s: %w", repo, normalizeGitHubAPIError(err))
				}

				totalAlerts += len(batch)
				alertTypeCounts[alertTypeCodeScanning] += len(batch)

				if resp == nil || resp.NextPage == 0 {
					break
				}
				page = resp.NextPage
			}
		}

		if alertTypeRequested(alertTypes, alertTypeSecretScanning) {
			state := lo.CoalesceOrEmpty(cfg.AlertState, defaultAlertState)
			page := 1
			for {
				if err := ctx.Err(); err != nil {
					return nil, err
				}

				opts := &gh.SecretScanningAlertListOptions{
					State:       state,
					ListOptions: gh.ListOptions{Page: page, PerPage: defaultPerPage},
				}

				batch, resp, err := c.SecretScanning.ListAlertsForRepo(ctx, owner, repository, opts)
				if err != nil {
					return nil, fmt.Errorf("githuboauth: secret scanning alerts failed for %s: %w", repo, normalizeGitHubAPIError(err))
				}

				totalAlerts += len(batch)
				alertTypeCounts[alertTypeSecretScanning] += len(batch)

				if resp == nil || resp.NextPage == 0 {
					break
				}
				page = resp.NextPage
			}
		}
	}

	return jsonx.ToRawMessage(vulnerabilityDetails{
		RepositoriesScanned: len(repoNames),
		AlertsTotal:         totalAlerts,
		AlertTypeCounts:     alertTypeCounts,
	})
}

// listUserRepos paginates all repositories accessible to the OAuth token
func listUserRepos(ctx context.Context, client *gh.Client, perPage int) ([]*gh.Repository, error) {
	perPage = clampPerPage(perPage)
	out := make([]*gh.Repository, 0)
	page := 1

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		batch, resp, err := client.Repositories.List(ctx, "", &gh.RepositoryListOptions{
			ListOptions: gh.ListOptions{Page: page, PerPage: perPage},
		})
		if err != nil {
			return nil, normalizeGitHubAPIError(err)
		}

		out = append(out, batch...)

		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return out, nil
}

var defaultAlertTypes = []string{alertTypeDependabot, alertTypeCodeScanning, alertTypeSecretScanning}

func resolveAlertTypes(values []string) []string {
	normalized := lo.Uniq(lo.Filter(values, func(v string, _ int) bool {
		return strings.TrimSpace(v) != ""
	}))
	if len(normalized) == 0 {
		return defaultAlertTypes
	}
	return normalized
}

func alertTypeRequested(alertTypes []string, target string) bool {
	if len(alertTypes) == 0 {
		return true
	}
	return lo.Contains(alertTypes, target)
}

func splitRepository(value string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(value), "/", githubRepositoryParts)
	if len(parts) != githubRepositoryParts || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%w: %q", ErrRepositoryInvalid, value)
	}
	return parts[0], parts[1], nil
}

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
