package gcpscc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"github.com/samber/lo"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Operation names published by the GCP SCC provider.
const (
	OperationHealthDefault   types.OperationName = types.OperationHealthDefault
	OperationCollectFindings types.OperationName = "findings.collect"
	OperationScanSettings    types.OperationName = "settings.scan"

	findingsPageSize      = 100
	findingsMaxPageSize   = 1000
	settingsPageSize      = 10
	sampleConfigsCapacity = 5
	sccAlertTypeFinding   = "finding"
)

var (
	_ types.OperationProvider = (*Provider)(nil)
)

type securityCenterFindingsConfig struct {
	// Pagination controls page sizing for SCC findings
	operations.Pagination
	// PayloadOptions controls payload inclusion for findings
	operations.PayloadOptions

	// Filter overrides the stored findings filter
	Filter types.TrimmedString `json:"filter"`
	// SourceID overrides the stored SCC source identifier
	SourceID types.TrimmedString `json:"sourceId"`
	// SourceIDs overrides stored SCC source identifiers for fan-out collection
	SourceIDs []string `json:"sourceIds"`
	// MaxFindings caps the number of findings returned
	MaxFindings int `json:"max_findings"`
}

type securityCenterFindingsSchema struct {
	// SourceID overrides the SCC source identifier
	SourceID types.TrimmedString `json:"sourceId,omitempty" jsonschema:"description=Optional SCC source override (full resource name or bare source ID)."`
	// SourceIDs overrides SCC source identifiers for fan-out collection
	SourceIDs []string `json:"sourceIds,omitempty" jsonschema:"description=Optional SCC source overrides for fan-out collection. Bare source IDs expand against selected parents."`
	// Filter overrides the stored SCC findings filter
	Filter types.TrimmedString `json:"filter,omitempty" jsonschema:"description=Optional SCC findings filter overriding stored metadata."`
	// PageSize overrides the findings page size
	PageSize int `json:"page_size,omitempty" jsonschema:"description=Optional page size override (max 1000)."`
	// MaxFindings caps the total findings returned
	MaxFindings int `json:"max_findings,omitempty" jsonschema:"description=Optional cap on total findings returned."`
	// IncludePayloads controls whether raw payloads are returned
	IncludePayloads bool `json:"include_payloads,omitempty" jsonschema:"description=Return raw finding payloads in the response (defaults to false)."`
}

type securityCenterHealthDetails struct {
	Parents []string `json:"parents"`
}

type securityCenterFindingSample struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	State    string `json:"state"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
}

type securityCenterFindingsDetails struct {
	Sources          []string                      `json:"sources"`
	SourceCount      int                           `json:"sourceCount"`
	Filter           string                        `json:"filter"`
	TotalFindings    int                           `json:"totalFindings"`
	FindingsBySource map[string]int                `json:"findingsBySource"`
	SeverityCounts   map[string]int                `json:"severity_counts"`
	StateCounts      map[string]int                `json:"state_counts"`
	Samples          []securityCenterFindingSample `json:"samples"`
	Alerts           []types.AlertEnvelope         `json:"alerts,omitempty"`
}

type securityCenterNotificationConfigSample struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PubSubTopic string `json:"pubsubTopic"`
	Parent      string `json:"parent"`
}

type securityCenterSettingsDetails struct {
	Parents                   []string                                 `json:"parents"`
	NotificationConfigCount   int                                      `json:"notificationConfigCount"`
	SampleNotificationConfigs []securityCenterNotificationConfigSample `json:"sampleNotificationConfigs"`
}

type securityCenterFailureDetails struct {
	Parent  string   `json:"parent,omitempty"`
	Parents []string `json:"parents,omitempty"`
	Source  string   `json:"source,omitempty"`
	Sources []string `json:"sources,omitempty"`
	Filter  string   `json:"filter,omitempty"`
}

var securityCenterFindingsConfigSchema = operations.SchemaFrom[securityCenterFindingsSchema]()

// Operations returns the provider operations published by GCP SCC.
func (p *Provider) Operations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Provider:    TypeGCPSCC,
			Name:        OperationHealthDefault,
			Kind:        types.OperationKindHealth,
			Description: "Validate Security Command Center access by listing sources.",
			Client:      ClientSecurityCenter,
			Run:         runSecurityCenterHealthOperation,
		},
		{
			Provider:     TypeGCPSCC,
			Name:         OperationCollectFindings,
			Kind:         types.OperationKindCollectFindings,
			Description:  "Collect Security Command Center findings using the configured source/filter.",
			Client:       ClientSecurityCenter,
			ConfigSchema: securityCenterFindingsConfigSchema,
			Run:          runSecurityCenterFindingsOperation,
		},
		{
			Provider:    TypeGCPSCC,
			Name:        OperationScanSettings,
			Kind:        types.OperationKindScanSettings,
			Description: "Inspect Security Command Center organization settings.",
			Client:      ClientSecurityCenter,
			Run:         runSecurityCenterSettingsOperation,
		},
	}
}

// runSecurityCenterHealthOperation checks SCC reachability for the org or project
func runSecurityCenterHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	meta, err := metadataFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	client, ok := types.ClientInstanceAs[*cloudscc.Client](input.Client)
	if !ok {
		return types.OperationResult{}, ErrSecurityCenterClientRequired
	}

	parents, err := resolveSecurityCenterParents(meta)
	if err != nil {
		return types.OperationResult{}, err
	}

	for _, parent := range parents {
		req := &securitycenterpb.ListSourcesRequest{
			Parent:   parent,
			PageSize: 1,
		}

		it := client.ListSources(ctx, req)
		_, err = it.Next()
		if errors.Is(err, iterator.Done) {
			err = nil
		}
		if err != nil {
			return operations.OperationFailure("Security Command Center list sources failed", err, securityCenterFailureDetails{
				Parent:  parent,
				Parents: parents,
			})
		}
	}

	return operations.OperationSuccess(fmt.Sprintf("Security Command Center reachable for %d parent(s)", len(parents)), securityCenterHealthDetails{
		Parents: parents,
	}), nil
}

// runSecurityCenterFindingsOperation collects findings from SCC
func runSecurityCenterFindingsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	meta, err := metadataFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	config, err := decodeSecurityCenterFindingsConfig(input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	client, ok := types.ClientInstanceAs[*cloudscc.Client](input.Client)
	if !ok {
		return types.OperationResult{}, ErrSecurityCenterClientRequired
	}

	sources, err := resolveSecurityCenterSources(meta, config)
	if err != nil {
		return types.OperationResult{}, err
	}

	filter := lo.CoalesceOrEmpty(config.Filter, meta.FindingFilter).String()

	pageSize := config.EffectivePageSize(findingsPageSize)
	if pageSize <= 0 {
		pageSize = findingsPageSize
	}
	if pageSize > findingsMaxPageSize {
		pageSize = findingsMaxPageSize
	}

	maxFindings := config.MaxFindings

	total := 0
	samples := make([]securityCenterFindingSample, 0, operations.DefaultSampleSize)
	envelopes := make([]types.AlertEnvelope, 0)
	severityCounts := map[string]int{}
	stateCounts := map[string]int{}
	sourceCounts := map[string]int{}
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

collectLoop:
	for _, sourceName := range sources {
		req := &securitycenterpb.ListFindingsRequest{
			Parent:   sourceName,
			Filter:   filter,
			PageSize: int32(min(pageSize, math.MaxInt32)), //nolint:gosec // bounds checked via min
		}

		it := client.ListFindings(ctx, req)

		for {
			result, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				return operations.OperationFailure("Security Command Center list findings failed", err, securityCenterFailureDetails{
					Sources: sources,
					Filter:  filter,
					Source:  sourceName,
				})
			}

			finding := result.GetFinding()
			if finding == nil {
				continue
			}

			if maxFindings > 0 && total >= maxFindings {
				break collectLoop
			}

			payload, err := marshaler.Marshal(finding)
			if err != nil {
				return operations.OperationFailure("Security Command Center finding serialization failed", err, securityCenterFailureDetails{
					Sources: sources,
					Source:  sourceName,
				})
			}

			resourceName := finding.GetResourceName()
			envelopes = append(envelopes, types.AlertEnvelope{
				AlertType: sccAlertTypeFinding,
				Resource:  resourceName,
				Payload:   payload,
			})
			total++
			sourceCounts[sourceName]++

			if severity := finding.GetSeverity().String(); severity != "" {
				key := strings.ToLower(severity)
				if key != "severity_unspecified" {
					severityCounts[key]++
				}
			}
			if state := finding.GetState().String(); state != "" {
				key := strings.ToLower(state)
				if key != "state_unspecified" {
					stateCounts[key]++
				}
			}

			if len(samples) < cap(samples) {
				samples = append(samples, securityCenterFindingSample{
					Name:     finding.GetName(),
					Category: finding.GetCategory(),
					State:    finding.GetState().String(),
					Severity: finding.GetSeverity().String(),
					Source:   sourceName,
				})
			}
		}
	}

	details := securityCenterFindingsDetails{
		Sources:          sources,
		SourceCount:      len(sources),
		Filter:           filter,
		TotalFindings:    total,
		FindingsBySource: sourceCounts,
		SeverityCounts:   severityCounts,
		StateCounts:      stateCounts,
		Samples:          samples,
	}
	if config.IncludePayloads {
		details.Alerts = envelopes
	}

	return operations.OperationSuccess(fmt.Sprintf("Collected %d findings from %d source(s)", total, len(sources)), details), nil
}

// runSecurityCenterSettingsOperation lists SCC notification configs
func runSecurityCenterSettingsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	meta, err := metadataFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	client, ok := types.ClientInstanceAs[*cloudscc.Client](input.Client)
	if !ok {
		return types.OperationResult{}, ErrSecurityCenterClientRequired
	}

	parents, err := resolveSecurityCenterParents(meta)
	if err != nil {
		return types.OperationResult{}, err
	}

	configs := make([]securityCenterNotificationConfigSample, 0, sampleConfigsCapacity)
	count := 0

	for _, parent := range parents {
		req := &securitycenterpb.ListNotificationConfigsRequest{
			Parent:   parent,
			PageSize: settingsPageSize,
		}

		it := client.ListNotificationConfigs(ctx, req)
		for {
			cfg, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				return operations.OperationFailure("Security Command Center notification config scan failed", err, securityCenterFailureDetails{
					Parents: parents,
					Parent:  parent,
				})
			}

			count++
			if len(configs) < cap(configs) {
				configs = append(configs, securityCenterNotificationConfigSample{
					Name:        cfg.GetName(),
					Description: cfg.GetDescription(),
					PubSubTopic: cfg.GetPubsubTopic(),
					Parent:      parent,
				})
			}
		}
	}

	return operations.OperationSuccess(fmt.Sprintf("Discovered %d notification configs across %d parent(s)", count, len(parents)), securityCenterSettingsDetails{
		Parents:                   parents,
		NotificationConfigCount:   count,
		SampleNotificationConfigs: configs,
	}), nil
}

// resolveSecurityCenterParents chooses the SCC parent resources used for health/settings checks.
func resolveSecurityCenterParents(meta credentialMetadata) ([]string, error) {
	if meta.OrganizationID != "" && meta.ProjectScope != projectScopeSpecific {
		return []string{fmt.Sprintf("organizations/%s", meta.OrganizationID)}, nil
	}

	if meta.ProjectScope == projectScopeSpecific {
		parents := lo.FilterMap(meta.ProjectIDs, func(projectID string, _ int) (string, bool) {
			value := strings.TrimSpace(projectID)
			if value == "" {
				return "", false
			}
			return fmt.Sprintf("projects/%s", value), true
		})
		parents = lo.Uniq(parents)
		if len(parents) == 0 {
			return nil, ErrProjectIDRequired
		}
		return parents, nil
	}

	if meta.ProjectID != "" {
		return []string{fmt.Sprintf("projects/%s", meta.ProjectID)}, nil
	}

	if meta.OrganizationID != "" {
		return []string{fmt.Sprintf("organizations/%s", meta.OrganizationID)}, nil
	}

	return nil, ErrProjectIDRequired
}

// resolveSecurityCenterSources resolves source resource names from config and metadata.
func resolveSecurityCenterSources(meta credentialMetadata, config securityCenterFindingsConfig) ([]string, error) {
	raw := make([]string, 0, 1+len(meta.SourceIDs))
	if config.SourceID != "" {
		raw = append(raw, config.SourceID.String())
	}
	raw = append(raw, config.SourceIDs...)
	if len(raw) == 0 {
		raw = append(raw, meta.SourceIDs...)
		if len(raw) == 0 && meta.SourceID != "" {
			raw = append(raw, meta.SourceID.String())
		}
	}
	if len(raw) == 0 {
		return nil, ErrSourceIDRequired
	}

	parents, err := resolveSecurityCenterParents(meta)
	if err != nil {
		return nil, err
	}

	out := lo.Uniq(lo.FlatMap(raw, func(source string, _ int) []string {
		source = strings.TrimSpace(source)
		if source == "" {
			return nil
		}
		if strings.HasPrefix(source, "organizations/") || strings.HasPrefix(source, "projects/") {
			return []string{source}
		}
		return lo.Map(parents, func(parent string, _ int) string {
			return fmt.Sprintf("%s/sources/%s", parent, source)
		})
	}))
	if len(out) == 0 {
		return nil, ErrSourceIDRequired
	}

	return out, nil
}

// decodeSecurityCenterFindingsConfig decodes operation config into a typed struct
func decodeSecurityCenterFindingsConfig(config json.RawMessage) (securityCenterFindingsConfig, error) {
	var decoded securityCenterFindingsConfig
	if err := jsonx.UnmarshalIfPresent(config, &decoded); err != nil {
		return decoded, err
	}
	decoded.SourceIDs = types.NormalizeStringSlice(decoded.SourceIDs)
	return decoded, nil
}
