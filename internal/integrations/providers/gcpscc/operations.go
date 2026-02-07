package gcpscc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
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
	// MaxFindings caps the number of findings returned
	MaxFindings int `json:"max_findings"`
}

type securityCenterFindingsSchema struct {
	// SourceID overrides the SCC source identifier
	SourceID types.TrimmedString `json:"sourceId,omitempty" jsonschema:"description=Optional SCC source override (full resource name or bare source ID)."`
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

	parent, err := resolveSecurityCenterParent(meta)
	if err != nil {
		return types.OperationResult{}, err
	}

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
		details := map[string]any{
			"parent": parent,
			"error":  err.Error(),
		}
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Security Command Center list sources failed",
			Details: details,
		}, err
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Security Command Center reachable for %s", parent),
		Details: map[string]any{
			"parent": parent,
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

	sourceName, err := resolveSecurityCenterSource(meta, config)
	if err != nil {
		return types.OperationResult{}, err
	}

	filter := string(config.Filter)
	if filter == "" {
		filter = string(meta.FindingFilter)
	}

	pageSize := config.EffectivePageSize(findingsPageSize)
	if pageSize <= 0 {
		pageSize = findingsPageSize
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	maxFindings := config.MaxFindings

	req := &securitycenterpb.ListFindingsRequest{
		Parent:   sourceName,
		Filter:   filter,
		PageSize: int32(pageSize),
	}

	it := client.ListFindings(ctx, req)
	total := 0
	samples := make([]map[string]any, 0, maxSampleSize)
	envelopes := make([]types.AlertEnvelope, 0)
	severityCounts := map[string]int{}
	stateCounts := map[string]int{}
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	for {
		result, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "Security Command Center list findings failed",
				Details: map[string]any{
					"source": sourceName,
					"filter": filter,
					"error":  err.Error(),
				},
			}, err
		}

		finding := result.GetFinding()
		if finding == nil {
			continue
		}

		if maxFindings > 0 && total >= maxFindings {
			break
		}

		payload, err := marshaler.Marshal(finding)
		if err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "Security Command Center finding serialization failed",
				Details: map[string]any{
					"source": sourceName,
					"error":  err.Error(),
				},
			}, err
		}

		resourceName := finding.GetResourceName()
		envelopes = append(envelopes, types.AlertEnvelope{
			AlertType: sccAlertTypeFinding,
			Resource:  resourceName,
			Payload:   payload,
		})
		total++

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
			})
		}
	}

	details := map[string]any{
		"source":          sourceName,
		"filter":          filter,
		"totalFindings":   total,
		"severity_counts": severityCounts,
		"state_counts":    stateCounts,
		"samples":         samples,
	}
	details = operations.AddPayloadIf(details, config.IncludePayloads, "alerts", envelopes)

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d findings from %s", total, sourceName),
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

	parent, err := resolveSecurityCenterParent(meta)
	if err != nil {
		return types.OperationResult{}, err
	}

	req := &securitycenterpb.ListNotificationConfigsRequest{
		Parent:   parent,
		PageSize: settingsPageSize,
	}

	it := client.ListNotificationConfigs(ctx, req)
	configs := make([]map[string]any, 0, sampleConfigsCapacity)
	count := 0

	for {
		cfg, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "Security Command Center notification config scan failed",
				Details: map[string]any{
					"parent": parent,
					"error":  err.Error(),
				},
			}, err
		}

		count++
		if len(configs) < cap(configs) {
			configs = append(configs, map[string]any{
				"name":        cfg.GetName(),
				"description": cfg.GetDescription(),
				"pubsubTopic": cfg.GetPubsubTopic(),
			})
		}
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Discovered %d notification configs under %s", count, parent),
		Details: map[string]any{
			"parent":                    parent,
			"notificationConfigCount":   count,
			"sampleNotificationConfigs": configs,
		},
	}, nil
}

// resolveSecurityCenterParent chooses an org or project parent resource
func resolveSecurityCenterParent(meta credentialMetadata) (string, error) {
	if meta.OrganizationID != "" {
		return fmt.Sprintf("organizations/%s", meta.OrganizationID), nil
	}

	if meta.ProjectID != "" {
		return fmt.Sprintf("projects/%s", meta.ProjectID), nil
	}

	return "", errProjectIDRequired
}

// resolveSecurityCenterSource resolves the source name from config or metadata
func resolveSecurityCenterSource(meta credentialMetadata, config securityCenterFindingsConfig) (string, error) {
	if config.SourceID != "" {
		return normalizeSourceName(string(config.SourceID), meta)
	}

	if meta.SourceID != "" {
		return normalizeSourceName(string(meta.SourceID), meta)
	}

	return "", errSourceIDRequired
}

// normalizeSourceName expands a source short name to a full resource path
func normalizeSourceName(source string, meta credentialMetadata) (string, error) {
	if strings.HasPrefix(source, "organizations/") || strings.HasPrefix(source, "projects/") {
		return source, nil
	}

	parent, err := resolveSecurityCenterParent(meta)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/sources/%s", parent, source), nil
}

// decodeSecurityCenterFindingsConfig decodes operation config into a typed struct
func decodeSecurityCenterFindingsConfig(config map[string]any) (securityCenterFindingsConfig, error) {
	var decoded securityCenterFindingsConfig
	if err := operations.DecodeConfig(config, &decoded); err != nil {
		return decoded, err
	}
	return decoded, nil
}
