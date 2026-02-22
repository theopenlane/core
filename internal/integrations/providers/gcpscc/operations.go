package gcpscc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"github.com/samber/lo"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

// Operation names published by the GCP SCC provider.
const (
	OperationHealthDefault   types.OperationName = "health.default"
	OperationCollectFindings types.OperationName = "findings.collect"
	OperationScanSettings    types.OperationName = "settings.scan"

	findingsPageSize      = 100
	findingsMaxPageSize   = 1000
	maxSampleSize         = 5
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

	client, ok := input.Client.(*cloudscc.Client)
	if !ok || client == nil {
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
			return operations.OperationFailure("Security Command Center list sources failed", err, map[string]any{
				"parent":  parent,
				"parents": parents,
			})
		}
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Security Command Center reachable for %d parent(s)", len(parents)),
		Details: map[string]any{
			"parents": parents,
		},
	}, nil
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

	client, ok := input.Client.(*cloudscc.Client)
	if !ok || client == nil {
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
	samples := make([]map[string]any, 0, maxSampleSize)
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
				return operations.OperationFailure("Security Command Center list findings failed", err, map[string]any{
					"sources": sources,
					"filter":  filter,
					"source":  sourceName,
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
				return operations.OperationFailure("Security Command Center finding serialization failed", err, map[string]any{
					"sources": sources,
					"source":  sourceName,
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
				samples = append(samples, map[string]any{
					"name":     finding.GetName(),
					"category": finding.GetCategory(),
					"state":    finding.GetState().String(),
					"severity": finding.GetSeverity().String(),
					"source":   sourceName,
				})
			}
		}
	}

	details := map[string]any{
		"sources":          sources,
		"sourceCount":      len(sources),
		"filter":           filter,
		"totalFindings":    total,
		"findingsBySource": sourceCounts,
		"severity_counts":  severityCounts,
		"state_counts":     stateCounts,
		"samples":          samples,
	}
	details = operations.AddPayloadIf(details, config.IncludePayloads, "alerts", envelopes)

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d findings from %d source(s)", total, len(sources)),
		Details: details,
	}, nil
}

// runSecurityCenterSettingsOperation lists SCC notification configs
func runSecurityCenterSettingsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	meta, err := metadataFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	client, ok := input.Client.(*cloudscc.Client)
	if !ok || client == nil {
		return types.OperationResult{}, ErrSecurityCenterClientRequired
	}

	parents, err := resolveSecurityCenterParents(meta)
	if err != nil {
		return types.OperationResult{}, err
	}

	configs := make([]map[string]any, 0, sampleConfigsCapacity)
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
				return operations.OperationFailure("Security Command Center notification config scan failed", err, map[string]any{
					"parents": parents,
					"parent":  parent,
				})
			}

			count++
			if len(configs) < cap(configs) {
				configs = append(configs, map[string]any{
					"name":        cfg.GetName(),
					"description": cfg.GetDescription(),
					"pubsubTopic": cfg.GetPubsubTopic(),
					"parent":      parent,
				})
			}
		}
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Discovered %d notification configs across %d parent(s)", count, len(parents)),
		Details: map[string]any{
			"parents":                   parents,
			"notificationConfigCount":   count,
			"sampleNotificationConfigs": configs,
		},
	}, nil
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
func decodeSecurityCenterFindingsConfig(config map[string]any) (securityCenterFindingsConfig, error) {
	var decoded securityCenterFindingsConfig
	if err := operations.DecodeConfig(config, &decoded); err != nil {
		return decoded, err
	}
	decoded.SourceIDs = types.NormalizeStringSlice(decoded.SourceIDs)
	return decoded, nil
}
