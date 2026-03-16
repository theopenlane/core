package githubapp

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/shurcooL/githubv4"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100

	githubAlertTypeDependabot   = "dependabot"
	githubAlertTypeCodeScanning = "code_scanning"
	githubAlertTypeSecretScan   = "secret_scanning"
)

// VulnerabilityCollectConfig controls the vulnerability collect operation
type VulnerabilityCollectConfig struct {
	// MaxRepos caps the number of repositories scanned during one run
	MaxRepos int `json:"max_repos,omitempty" jsonschema:"description=Optional cap on the number of repositories to scan."`
	// State filters alerts by GitHub vulnerability alert state
	State string `json:"state,omitempty" jsonschema:"description=Alert state filter (OPEN, DISMISSED, FIXED, AUTO_DISMISSED). Defaults to OPEN.,enum=OPEN,enum=DISMISSED,enum=FIXED,enum=AUTO_DISMISSED"`
	// Severity filters alerts by advisory severity
	Severity string `json:"severity,omitempty" jsonschema:"description=Optional severity filter (LOW, MODERATE, HIGH, CRITICAL).,enum=LOW,enum=MODERATE,enum=HIGH,enum=CRITICAL"`
}

// VulnerabilityAlertPayload is the raw provider payload emitted for ingest
type VulnerabilityAlertPayload struct {
	// Number is the GitHub alert number
	Number int `json:"number,omitempty"`
	// State is the current lifecycle state of the alert
	State string `json:"state,omitempty"`
	// HTMLURL is the GitHub UI URL for the alert
	HTMLURL string `json:"html_url,omitempty"`
	// CreatedAt is when GitHub created the alert
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt is when GitHub last updated the alert
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// SecurityAdvisory contains the advisory details attached to the alert
	SecurityAdvisory SecurityAdvisory `json:"security_advisory,omitempty"`
}

// SecurityAdvisory is the normalized advisory shape used by the GitHub App mapping
type SecurityAdvisory struct {
	// GHSAID is the GitHub Security Advisory identifier
	GHSAID string `json:"ghsa_id,omitempty"`
	// Summary is the short advisory summary
	Summary string `json:"summary,omitempty"`
	// Description is the full advisory description
	Description string `json:"description,omitempty"`
	// Severity is the advisory severity string reported by GitHub
	Severity string `json:"severity,omitempty"`
	// CVEID is the advisory CVE identifier when present
	CVEID string `json:"cve_id,omitempty"`
}

// HealthCheck validates the GitHub App installation token
type HealthCheck struct{}

// RepositorySync lists repositories accessible to the installation
type RepositorySync struct{}

// VulnerabilityCollect collects repository vulnerability alerts for ingest
type VulnerabilityCollect struct{}

type pageInfo struct {
	EndCursor   string
	HasNextPage bool
}

type repositoryNode struct {
	NameWithOwner string
	IsPrivate     bool
	UpdatedAt     time.Time
	URL           string `graphql:"url"`
}

type vulnerabilityAlertNode struct {
	Number                int
	State                 string
	URL                   string `graphql:"url"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	SecurityVulnerability struct {
		Severity string
		Advisory struct {
			GHSAID      string `graphql:"ghsaId"`
			Summary     string
			Description string
			Identifiers []struct {
				Type  string
				Value string
			}
		}
	}
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		githubClient, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, githubClient)
	}
}

// Run executes the health check using the GitHub GraphQL client
func (HealthCheck) Run(ctx context.Context, client *githubv4.Client) (json.RawMessage, error) {
	_, err := queryRepositories(ctx, client, 1)
	if err != nil {
		return nil, err
	}

	return jsonx.ToRawMessage(map[string]any{})
}

// Handle adapts repository sync to the generic operation registration boundary
func (r RepositorySync) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		githubClient, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return r.Run(ctx, githubClient)
	}
}

// Run enumerates repositories accessible to the installation
func (RepositorySync) Run(ctx context.Context, client *githubv4.Client) (json.RawMessage, error) {
	repositories, err := queryRepositories(ctx, client, defaultPageSize)
	if err != nil {
		return nil, err
	}

	sampleSize := min(len(repositories), 10)
	samples := lo.Map(repositories[:sampleSize], func(repository repositoryNode, _ int) map[string]any {
		return map[string]any{
			"name":       repository.NameWithOwner,
			"private":    repository.IsPrivate,
			"updated_at": repository.UpdatedAt,
			"url":        repository.URL,
		}
	})

	return jsonx.ToRawMessage(map[string]any{
		"count":   len(repositories),
		"samples": samples,
	})
}

// Handle adapts vulnerability collection to the generic operation registration boundary
func (v VulnerabilityCollect) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		githubClient, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		var collectConfig VulnerabilityCollectConfig
		if err := jsonx.UnmarshalIfPresent(request.Config, &collectConfig); err != nil {
			return nil, err
		}

		return v.Run(ctx, githubClient, collectConfig)
	}
}

// Run collects repository vulnerability alerts and emits ingest payloads
func (VulnerabilityCollect) Run(ctx context.Context, client *githubv4.Client, config VulnerabilityCollectConfig) (json.RawMessage, error) {
	repositories, err := queryRepositories(ctx, client, defaultPageSize)
	if err != nil {
		return nil, err
	}

	if config.MaxRepos > 0 && len(repositories) > config.MaxRepos {
		repositories = repositories[:config.MaxRepos]
	}

	envelopes := make([]types.MappingEnvelope, 0)

	for _, repository := range repositories {
		alerts, err := queryRepositoryVulnerabilityAlerts(ctx, client, repository.NameWithOwner, config)
		if err != nil {
			return nil, err
		}

		mapped, err := buildMappingEnvelopes(repository.NameWithOwner, alerts)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, mapped...)
	}

	return jsonx.ToRawMessage([]types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: envelopes,
		},
	})
}

// queryRepositories lists repositories accessible to the installation
func queryRepositories(ctx context.Context, client *githubv4.Client, pageSize int) ([]repositoryNode, error) {
	repositories := make([]repositoryNode, 0)
	pageSize = clampPageSize(pageSize)
	var after *githubv4.String

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var query struct {
			Viewer struct {
				Repositories struct {
					Nodes    []repositoryNode
					PageInfo pageInfo
				} `graphql:"repositories(first: $first, after: $after, orderBy: {field: UPDATED_AT, direction: DESC})"`
			}
		}

		variables := map[string]any{
			"first": githubv4.Int(pageSize),
			"after": after,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, normalizeGitHubAPIError(err)
		}

		repositories = append(repositories, query.Viewer.Repositories.Nodes...)
		if !query.Viewer.Repositories.PageInfo.HasNextPage || query.Viewer.Repositories.PageInfo.EndCursor == "" {
			break
		}

		after = githubv4.NewString(githubv4.String(query.Viewer.Repositories.PageInfo.EndCursor))
	}

	return repositories, nil
}

// queryRepositoryVulnerabilityAlerts lists vulnerability alerts for one repository
func queryRepositoryVulnerabilityAlerts(ctx context.Context, client *githubv4.Client, repository string, config VulnerabilityCollectConfig) ([]VulnerabilityAlertPayload, error) {
	repositoryOwner, repositoryName, ok := strings.Cut(repository, "/")
	if !ok || repositoryOwner == "" || repositoryName == "" {
		return nil, ErrAPIRequest
	}

	alerts := make([]VulnerabilityAlertPayload, 0)
	var after *githubv4.String

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var query struct {
			Repository struct {
				VulnerabilityAlerts struct {
					Nodes    []vulnerabilityAlertNode
					PageInfo pageInfo
				} `graphql:"vulnerabilityAlerts(first: $first, after: $after, states: $states)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables := map[string]interface{}{
			"owner":  githubv4.String(repositoryOwner),
			"name":   githubv4.String(repositoryName),
			"first":  githubv4.Int(defaultPageSize),
			"after":  after,
			"states": config.states(),
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, normalizeGitHubAPIError(err)
		}

		for _, alert := range query.Repository.VulnerabilityAlerts.Nodes {
			payload := mapVulnerabilityAlert(alert)
			if config.Severity != "" && payload.SecurityAdvisory.Severity != config.Severity {
				continue
			}

			alerts = append(alerts, payload)
		}

		if !query.Repository.VulnerabilityAlerts.PageInfo.HasNextPage || query.Repository.VulnerabilityAlerts.PageInfo.EndCursor == "" {
			break
		}

		after = githubv4.NewString(githubv4.String(query.Repository.VulnerabilityAlerts.PageInfo.EndCursor))
	}

	return alerts, nil
}

// mapVulnerabilityAlert converts a GraphQL alert node into the ingest payload shape
func mapVulnerabilityAlert(alert vulnerabilityAlertNode) VulnerabilityAlertPayload {
	return VulnerabilityAlertPayload{
		Number:    alert.Number,
		State:     alert.State,
		HTMLURL:   alert.URL,
		CreatedAt: alert.CreatedAt,
		UpdatedAt: alert.UpdatedAt,
		SecurityAdvisory: SecurityAdvisory{
			GHSAID:      alert.SecurityVulnerability.Advisory.GHSAID,
			Summary:     alert.SecurityVulnerability.Advisory.Summary,
			Description: alert.SecurityVulnerability.Advisory.Description,
			Severity:    alert.SecurityVulnerability.Severity,
			CVEID:       advisoryIdentifier(alert.SecurityVulnerability.Advisory.Identifiers, "CVE"),
		},
	}
}

// advisoryIdentifier returns the first advisory identifier matching the requested type
func advisoryIdentifier(identifiers []struct {
	Type  string
	Value string
}, identifierType string) string {
	for _, identifier := range identifiers {
		if identifier.Type == identifierType {
			return identifier.Value
		}
	}

	return ""
}

// states converts the configured alert state into GitHub GraphQL enum values
func (c VulnerabilityCollectConfig) states() []githubv4.RepositoryVulnerabilityAlertState {
	if c.State == "" {
		return []githubv4.RepositoryVulnerabilityAlertState{githubv4.RepositoryVulnerabilityAlertStateOpen}
	}

	return []githubv4.RepositoryVulnerabilityAlertState{
		githubv4.RepositoryVulnerabilityAlertState(c.State),
	}
}

// buildMappingEnvelopes wraps vulnerability payloads for mapping and ingest
func buildMappingEnvelopes(resource string, payloads []VulnerabilityAlertPayload) ([]types.MappingEnvelope, error) {
	envelopes := make([]types.MappingEnvelope, 0, len(payloads))

	for _, payload := range payloads {
		rawPayload, err := jsonx.ToRawMessage(payload)
		if err != nil {
			return nil, ErrIngestPayloadEncode
		}

		envelopes = append(envelopes, types.MappingEnvelope{
			Variant:  githubAlertTypeDependabot,
			Resource: resource,
			Payload:  rawPayload,
		})
	}

	return envelopes, nil
}

// clampPageSize constrains page sizes to the supported GitHub API range
func clampPageSize(value int) int {
	switch {
	case value <= 0:
		return defaultPageSize
	case value > maxPageSize:
		return maxPageSize
	default:
		return value
	}
}

// normalizeGitHubAPIError collapses provider-specific errors into integration errors
func normalizeGitHubAPIError(err error) error {
	if err == nil {
		return nil
	}

	return ErrAPIRequest
}
