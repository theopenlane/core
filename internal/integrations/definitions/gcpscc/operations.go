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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/internal/ent/generated"
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

// sccCredentialMetadata captures the persisted SCC credential metadata supplied during activation
type sccCredentialMetadata struct {
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
func (m sccCredentialMetadata) applyDefaults() sccCredentialMetadata {
	normalized := m
	if normalized.ProjectScope == "" {
		normalized.ProjectScope = projectScopeAll
	}

	normalized.ServiceAccountKey = normalizeServiceAccountKey(normalized.ServiceAccountKey)

	return normalized
}

type sccFindingsConfig struct {
	Filter      string   `json:"filter,omitempty"`
	SourceID    string   `json:"sourceId,omitempty"`
	SourceIDs   []string `json:"sourceIds,omitempty"`
	PageSize    int      `json:"page_size,omitempty"`
	MaxFindings int      `json:"max_findings,omitempty"`
}

type sccHealthDetails struct {
	Parents []string `json:"parents"`
}

type sccNotificationConfigSample struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PubSubTopic string `json:"pubsubTopic"`
	Parent      string `json:"parent"`
}

type sccSettingsDetails struct {
	Parents                   []string                      `json:"parents"`
	NotificationConfigCount   int                           `json:"notificationConfigCount"`
	SampleNotificationConfigs []sccNotificationConfigSample `json:"sampleNotificationConfigs"`
}

// buildSCCClient builds the GCP Security Command Center client for one installation
func buildSCCClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	meta, err := sccMetadataFromCredential(req.Credential)
	if err != nil {
		return nil, err
	}

	clientOpts, err := sccClientOptions(ctx, meta, req.Credential.OAuthAccessToken)
	if err != nil {
		return nil, err
	}

	opts := append([]option.ClientOption{}, clientOpts...)
	if meta.ProjectID != "" {
		opts = append(opts, option.WithQuotaProject(meta.ProjectID))
	}

	client, err := cloudscc.NewClient(ctx, opts...)
	if err != nil {
		return nil, ErrSecurityCenterClientCreate
	}

	return client, nil
}

// runHealthOperation verifies GCP SCC access by listing sources for each parent
func runHealthOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	sccClient, ok := client.(*cloudscc.Client)
	if !ok {
		return nil, ErrClientType
	}

	meta, err := sccMetadataFromCredential(credential)
	if err != nil {
		return nil, err
	}

	parents, err := resolveSecurityCenterParents(meta)
	if err != nil {
		return nil, err
	}

	for _, parent := range parents {
		req := &securitycenterpb.ListSourcesRequest{
			Parent:   parent,
			PageSize: 1,
		}

		it := sccClient.ListSources(ctx, req)
		_, err = it.Next()
		if errors.Is(err, iterator.Done) {
			err = nil
		}

		if err != nil {
			return nil, fmt.Errorf("gcpscc: list sources failed for %s: %w", parent, err)
		}
	}

	return jsonx.ToRawMessage(sccHealthDetails{Parents: parents})
}

// runFindingsCollectOperation collects GCP SCC findings from configured sources
func runFindingsCollectOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, config json.RawMessage) (json.RawMessage, error) {
	sccClient, ok := client.(*cloudscc.Client)
	if !ok {
		return nil, ErrClientType
	}

	meta, err := sccMetadataFromCredential(credential)
	if err != nil {
		return nil, err
	}

	var cfg sccFindingsConfig
	if err := jsonx.UnmarshalIfPresent(config, &cfg); err != nil {
		return nil, err
	}
	sources, err := resolveSecurityCenterSources(meta, cfg)
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

		it := sccClient.ListFindings(ctx, req)

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

// runSettingsScanOperation lists SCC notification configs for each parent
func runSettingsScanOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	sccClient, ok := client.(*cloudscc.Client)
	if !ok {
		return nil, ErrClientType
	}

	meta, err := sccMetadataFromCredential(credential)
	if err != nil {
		return nil, err
	}

	parents, err := resolveSecurityCenterParents(meta)
	if err != nil {
		return nil, err
	}

	configs := make([]sccNotificationConfigSample, 0, sampleConfigsCapacity)
	count := 0

	for _, parent := range parents {
		req := &securitycenterpb.ListNotificationConfigsRequest{
			Parent:   parent,
			PageSize: settingsPageSize,
		}

		it := sccClient.ListNotificationConfigs(ctx, req)
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
				configs = append(configs, sccNotificationConfigSample{
					Name:        cfg.GetName(),
					Description: cfg.GetDescription(),
					PubSubTopic: cfg.GetPubsubTopic(),
					Parent:      parent,
				})
			}
		}
	}

	return jsonx.ToRawMessage(sccSettingsDetails{
		Parents:                   parents,
		NotificationConfigCount:   count,
		SampleNotificationConfigs: configs,
	})
}

// sccMetadataFromCredential decodes SCC credential metadata from the credential set
func sccMetadataFromCredential(credential types.CredentialSet) (sccCredentialMetadata, error) {
	if len(credential.ProviderData) == 0 {
		return sccCredentialMetadata{}, ErrCredentialMetadataRequired
	}

	var meta sccCredentialMetadata
	if err := jsonx.UnmarshalIfPresent(credential.ProviderData, &meta); err != nil {
		return sccCredentialMetadata{}, ErrMetadataDecode
	}

	return meta.applyDefaults(), nil
}

// sccClientOptions builds client options based on available credentials
func sccClientOptions(ctx context.Context, meta sccCredentialMetadata, oauthToken string) ([]option.ClientOption, error) {
	if meta.ServiceAccountKey != "" {
		creds, err := serviceAccountCredentials(ctx, meta.ServiceAccountKey, meta.Scopes)
		if err != nil {
			return nil, err
		}

		return []option.ClientOption{option.WithCredentials(creds)}, nil
	}

	if oauthToken != "" {
		token := &oauth2.Token{AccessToken: oauthToken}
		return []option.ClientOption{option.WithTokenSource(oauth2.StaticTokenSource(token))}, nil
	}

	return nil, ErrAccessTokenMissing
}

// serviceAccountCredentials parses and validates a service account key
func serviceAccountCredentials(ctx context.Context, rawKey string, scopes []string) (*google.Credentials, error) {
	key := normalizeServiceAccountKey(rawKey)
	if key == "" {
		return nil, ErrServiceAccountKeyInvalid
	}

	scopeList := scopes
	if len(scopeList) == 0 {
		scopeList = []string{defaultScope}
	}

	creds, err := google.CredentialsFromJSONWithType(ctx, []byte(key), google.ServiceAccount, scopeList...)
	if err != nil {
		return nil, ErrServiceAccountKeyInvalid
	}

	return creds, nil
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

// resolveSecurityCenterParents chooses the SCC parent resources used for health/settings checks
func resolveSecurityCenterParents(meta sccCredentialMetadata) ([]string, error) {
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

// resolveSecurityCenterSources resolves source resource names from config and metadata
func resolveSecurityCenterSources(meta sccCredentialMetadata, cfg sccFindingsConfig) ([]string, error) {
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

	parents, err := resolveSecurityCenterParents(meta)
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
