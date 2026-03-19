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

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// findingsPageSize is the default number of SCC findings requested per paginated API call
	findingsPageSize = 100
	// findingsMaxPageSize is the maximum number of findings that can be requested per API page
	findingsMaxPageSize = 1000
)

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

// FindingsCollect collects GCP SCC findings for ingest
type FindingsCollect struct{}

// IngestHandle adapts findings collection to the ingest operation registration boundary
func (f FindingsCollect) IngestHandle() types.IngestHandler {
	return func(ctx context.Context, request types.OperationRequest) ([]types.IngestPayloadSet, error) {
		c, err := SCCClient.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		cfg, err := FindingsCollectOperation.UnmarshalConfig(request.Config)
		if err != nil {
			return nil, ErrOperationConfigInvalid
		}

		return f.Run(ctx, request.Credential, c, cfg)
	}
}

// Run collects GCP SCC findings from configured sources
func (FindingsCollect) Run(ctx context.Context, credential types.CredentialSet, c *cloudscc.Client, cfg FindingsConfig) ([]types.IngestPayloadSet, error) {
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
				return nil, ErrListFindingsFailed
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

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: envelopes,
		},
	}, nil
}

// buildFindingEnvelope serializes a finding into a mapping envelope
func buildFindingEnvelope(sourceName string, finding *securitycenterpb.Finding, marshaler protojson.MarshalOptions) (types.MappingEnvelope, error) {
	rawPayload, err := marshaler.Marshal(finding)
	if err != nil {
		return types.MappingEnvelope{}, ErrFindingEncode
	}

	return providerkit.RawEnvelope(resolveFindingResource(sourceName, finding), rawPayload), nil
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
func resolveParents(meta CredentialSchema) ([]string, error) {
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
func resolveSources(meta CredentialSchema, cfg FindingsConfig) ([]string, error) {
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
