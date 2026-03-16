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

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultScope          = "https://www.googleapis.com/auth/cloud-platform"
	projectScopeAll       = "all"
	projectScopeSpecific  = "specific"
	findingsPageSize      = 100
	findingsMaxPageSize   = 1000
	settingsPageSize      = 10
	sampleConfigsCapacity = 5
)

// credentialMetadata captures the persisted SCC credential metadata supplied during activation
type credentialMetadata struct {
	ProjectID                string   `json:"projectId,omitempty"`
	OrganizationID           string   `json:"organizationId,omitempty"`
	ProjectScope             string   `json:"projectScope,omitempty"`
	ProjectIDs               []string `json:"projectIds,omitempty"`
	WorkloadIdentityProvider string   `json:"workloadIdentityProvider,omitempty"`
	Audience                 string   `json:"audience,omitempty"`
	ServiceAccountEmail      string   `json:"serviceAccountEmail,omitempty"`
	SourceID                 string   `json:"sourceId,omitempty"`
	SourceIDs                []string `json:"sourceIds,omitempty"`
	Scopes                   []string `json:"scopes,omitempty"`
	TokenLifetime            string   `json:"tokenLifetime,omitempty"`
	FindingFilter            string   `json:"findingFilter,omitempty"`
	ServiceAccountKey        string   `json:"serviceAccountKey,omitempty"`
}

// applyDefaults fills in fallback values for missing optional fields
func (m credentialMetadata) applyDefaults() credentialMetadata {
	normalized := m
	if normalized.ProjectScope == "" {
		normalized.ProjectScope = projectScopeAll
	}

	normalized.ServiceAccountKey = normalizeServiceAccountKey(normalized.ServiceAccountKey)

	return normalized
}

// FindingsConfig holds per-invocation parameters for the findings.collect operation
type FindingsConfig struct {
	// Filter is a server-side CEL filter for findings
	Filter string `json:"filter,omitempty"`
	// SourceID scopes collection to a single SCC source
	SourceID string `json:"sourceId,omitempty"`
	// SourceIDs scopes collection to multiple SCC sources
	SourceIDs []string `json:"sourceIds,omitempty"`
	// PageSize controls the number of findings per API page
	PageSize int `json:"page_size,omitempty"`
	// MaxFindings caps the total number of findings returned
	MaxFindings int `json:"max_findings,omitempty"`
}

// HealthCheck holds the result of a GCP SCC health check
type HealthCheck struct {
	// Parents is the list of SCC parent resources that were probed
	Parents []string `json:"parents"`
}

// NotificationConfigSample holds a single SCC notification config entry
type NotificationConfigSample struct {
	// Name is the notification config resource name
	Name string `json:"name"`
	// Description is the notification config description
	Description string `json:"description"`
	// PubSubTopic is the Pub/Sub topic for the notification config
	PubSubTopic string `json:"pubsubTopic"`
	// Parent is the parent resource for the notification config
	Parent string `json:"parent"`
}

// SettingsScan scans GCP SCC notification settings
type SettingsScan struct {
	// Parents is the list of SCC parent resources that were scanned
	Parents []string `json:"parents"`
	// NotificationConfigCount is the total count of notification configs found
	NotificationConfigCount int `json:"notificationConfigCount"`
	// SampleNotificationConfigs holds a representative subset of notification configs
	SampleNotificationConfigs []NotificationConfigSample `json:"sampleNotificationConfigs"`
}

// FindingsCollect collects GCP SCC findings for ingest
type FindingsCollect struct{}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, request.Credential, c)
	}
}

// Run executes the GCP SCC health check
func (HealthCheck) Run(ctx context.Context, credential types.CredentialSet, c *cloudscc.Client) (json.RawMessage, error) {
	meta, err := metadataFromCredential(credential)
	if err != nil {
		return nil, err
	}

	parents, err := resolveParents(meta)
	if err != nil {
		return nil, err
	}

	for _, parent := range parents {
		req := &securitycenterpb.ListSourcesRequest{
			Parent:   parent,
			PageSize: 1,
		}

		it := c.ListSources(ctx, req)
		_, err = it.Next()

		if errors.Is(err, iterator.Done) {
			err = nil
		}

		if err != nil {
			return nil, fmt.Errorf("gcpscc: list sources failed for %s: %w", parent, err)
		}
	}

	return jsonx.ToRawMessage(HealthCheck{Parents: parents})
}

// Handle adapts findings collection to the generic operation registration boundary
func (f FindingsCollect) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		var cfg FindingsConfig
		if err := jsonx.UnmarshalIfPresent(request.Config, &cfg); err != nil {
			return nil, err
		}

		return f.Run(ctx, request.Credential, c, cfg)
	}
}

// Run collects GCP SCC findings from configured sources
func (FindingsCollect) Run(ctx context.Context, credential types.CredentialSet, c *cloudscc.Client, cfg FindingsConfig) (json.RawMessage, error) {
	meta, err := metadataFromCredential(credential)
	if err != nil {
		return nil, err
	}

	sources, err := resolveSources(meta, cfg)
	if err != nil {
		return nil, err
	}

	filter := lo.CoalesceOrEmpty(cfg.Filter, meta.FindingFilter)

	pageSize := cfg.PageSize
	if pageSize <= 0 {
		pageSize = findingsPageSize
	}

	if pageSize > findingsMaxPageSize {
		pageSize = findingsMaxPageSize
	}

	if maxFindings := cfg.MaxFindings; maxFindings > 0 && maxFindings < pageSize {
		pageSize = maxFindings
	}

	maxFindings := cfg.MaxFindings
	marshaler := protojson.MarshalOptions{UseProtoNames: true}
	envelopes := make([]types.MappingEnvelope, 0)

	if maxFindings > 0 {
		envelopes = make([]types.MappingEnvelope, 0, maxFindings)
	}

	collected := 0

collectLoop:
	for _, sourceName := range sources {
		req := &securitycenterpb.ListFindingsRequest{
			Parent:   sourceName,
			Filter:   filter,
			PageSize: int32(min(pageSize, math.MaxInt32)), //nolint:gosec // bounds checked via min
		}

		it := c.ListFindings(ctx, req)

		for {
			result, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("gcpscc: list findings failed for %s: %w", sourceName, err)
			}

			finding := result.GetFinding()
			if finding == nil {
				continue
			}

			if maxFindings > 0 && collected >= maxFindings {
				break collectLoop
			}

			envelope, err := buildFindingEnvelope(sourceName, finding, marshaler)
			if err != nil {
				return nil, err
			}

			envelopes = append(envelopes, envelope)
			collected++
		}
	}

	return jsonx.ToRawMessage([]types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: envelopes,
		},
	})
}

// Handle adapts settings scan to the generic operation registration boundary
func (s SettingsScan) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return s.Run(ctx, request.Credential, c)
	}
}

// Run scans GCP SCC notification configs
func (SettingsScan) Run(ctx context.Context, credential types.CredentialSet, c *cloudscc.Client) (json.RawMessage, error) {
	meta, err := metadataFromCredential(credential)
	if err != nil {
		return nil, err
	}

	parents, err := resolveParents(meta)
	if err != nil {
		return nil, err
	}

	configs := make([]NotificationConfigSample, 0, sampleConfigsCapacity)
	count := 0

	for _, parent := range parents {
		req := &securitycenterpb.ListNotificationConfigsRequest{
			Parent:   parent,
			PageSize: settingsPageSize,
		}

		it := c.ListNotificationConfigs(ctx, req)

		for {
			cfg, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("gcpscc: notification config scan failed for %s: %w", parent, err)
			}

			count++

			if len(configs) < cap(configs) {
				configs = append(configs, NotificationConfigSample{
					Name:        cfg.GetName(),
					Description: cfg.GetDescription(),
					PubSubTopic: cfg.GetPubsubTopic(),
					Parent:      parent,
				})
			}
		}
	}

	return jsonx.ToRawMessage(SettingsScan{
		Parents:                   parents,
		NotificationConfigCount:   count,
		SampleNotificationConfigs: configs,
	})
}

// buildFindingEnvelope serializes a finding into a mapping envelope
func buildFindingEnvelope(sourceName string, finding *securitycenterpb.Finding, marshaler protojson.MarshalOptions) (types.MappingEnvelope, error) {
	rawPayload, err := marshaler.Marshal(finding)
	if err != nil {
		return types.MappingEnvelope{}, fmt.Errorf("gcpscc: finding serialization failed for %s: %w", sourceName, err)
	}

	return types.MappingEnvelope{
		Resource: resolveFindingResource(sourceName, finding),
		Payload:  rawPayload,
	}, nil
}

// resolveFindingResource chooses the resource identifier used for ingest
func resolveFindingResource(sourceName string, finding *securitycenterpb.Finding) string {
	if finding != nil {
		resource := strings.TrimSpace(finding.GetResourceName())
		if resource != "" {
			return resource
		}
	}

	return strings.TrimSpace(sourceName)
}

// resolveParents chooses the SCC parent resources used for health/settings checks
func resolveParents(meta credentialMetadata) ([]string, error) {
	if meta.OrganizationID != "" && meta.ProjectScope != projectScopeSpecific {
		return []string{fmt.Sprintf("organizations/%s", meta.OrganizationID)}, nil
	}

	if meta.ProjectScope == projectScopeSpecific {
		parentList := lo.FilterMap(meta.ProjectIDs, func(projectID string, _ int) (string, bool) {
			value := strings.TrimSpace(projectID)
			if value == "" {
				return "", false
			}

			return fmt.Sprintf("projects/%s", value), true
		})

		parentList = lo.Uniq(parentList)

		if len(parentList) == 0 {
			return nil, ErrProjectIDRequired
		}

		return parentList, nil
	}

	if meta.ProjectID != "" {
		return []string{fmt.Sprintf("projects/%s", meta.ProjectID)}, nil
	}

	if meta.OrganizationID != "" {
		return []string{fmt.Sprintf("organizations/%s", meta.OrganizationID)}, nil
	}

	return nil, ErrProjectIDRequired
}

// resolveSources resolves source resource names from config and metadata
func resolveSources(meta credentialMetadata, cfg FindingsConfig) ([]string, error) {
	raw := make([]string, 0, 1+len(meta.SourceIDs))

	if cfg.SourceID != "" {
		raw = append(raw, cfg.SourceID)
	}

	raw = append(raw, cfg.SourceIDs...)

	if len(raw) == 0 {
		raw = append(raw, meta.SourceIDs...)

		if len(raw) == 0 && meta.SourceID != "" {
			raw = append(raw, meta.SourceID)
		}
	}

	if len(raw) == 0 {
		return nil, ErrSourceIDRequired
	}

	parents, err := resolveParents(meta)
	if err != nil {
		return nil, err
	}

	out := lo.Uniq(lo.FlatMap(raw, func(source string, _ int) []string {
		source = strings.TrimSpace(source)
		if source == "" {
			return nil
		}

		switch {
		case strings.HasPrefix(source, "organizations/"), strings.HasPrefix(source, "projects/"):
			return []string{source}
		default:
			return lo.Map(parents, func(parent string, _ int) string {
				return fmt.Sprintf("%s/sources/%s", parent, source)
			})
		}
	}))

	if len(out) == 0 {
		return nil, ErrSourceIDRequired
	}

	return out, nil
}

// normalizeServiceAccountKey trims and unwraps JSON-encoded service account keys
func normalizeServiceAccountKey(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	var decoded string
	if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
		return strings.TrimSpace(decoded)
	}

	return trimmed
}
