package gcpscc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"github.com/samber/lo"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// findingsPageSize is the default number of SCC findings requested per paginated API call
	findingsPageSize = 100
	// findingsMaxPageSize is the maximum number of findings that can be requested per API page
	findingsMaxPageSize = 1000
)

// FindingsSync holds per-invocation parameters for the findings.collect operation
type FindingsSync struct {
	// PageSize controls the number of findings per API page
	PageSize int `json:"page_size,omitempty"`
	// MaxFindings caps the total number of findings returned
	MaxFindings int `json:"max_findings,omitempty"`
}

// FindingsCollect collects GCP SCC findings for ingest
type FindingsCollect struct{}

// IngestHandle adapts findings collection to the ingest operation registration boundary
func (f FindingsCollect) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(sccClient, findingsCollectOperation, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *cloudscc.Client, cfg FindingsSync) ([]types.IngestPayloadSet, error) {
		return f.Run(ctx, request.Credentials, client, cfg, request.LastRunAt)
	})
}

// Run collects GCP SCC findings from configured sources
func (FindingsCollect) Run(ctx context.Context, credentials types.CredentialBindings, c *cloudscc.Client, cfg FindingsSync, lastRunAt *time.Time) ([]types.IngestPayloadSet, error) {
	meta, err := resolveCredential(credentials)
	if err != nil {
		return nil, err
	}

	sources, err := resolveSources(meta)
	if err != nil {
		return nil, err
	}

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
	findingEnvelopes := make([]types.MappingEnvelope, 0)
	vulnEnvelopes := make([]types.MappingEnvelope, 0)
	riskEnvelopes := make([]types.MappingEnvelope, 0)

	collected := 0

	var timeFilter string
	if lastRunAt != nil {
		timeFilter = fmt.Sprintf(`event_time >= "%s"`, lastRunAt.UTC().Format(time.RFC3339))
	}

collectLoop:
	for _, sourceName := range sources {
		req := &securitycenterpb.ListFindingsRequest{
			PageSize: int32(min(pageSize, math.MaxInt32)), //nolint:gosec // bounds checked via min
			Filter:   timeFilter,
		}

		if sourceName != "" {
			req.Parent = sourceName
		}

		it := c.ListFindings(ctx, req)

		for {
			result, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}

			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("gcpscc: error listing findings")
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

			switch {
			case finding.ParentDisplayName == "Risk Engine":
				riskEnvelopes = append(riskEnvelopes, envelope)
			case strings.EqualFold(finding.FindingClass.String(), "VULNERABILITY") && finding.Vulnerability != nil:
				vulnEnvelopes = append(vulnEnvelopes, envelope)
			default:
				findingEnvelopes = append(findingEnvelopes, envelope)
			}

			collected++
		}
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaFinding,
			Envelopes: findingEnvelopes,
		},
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaRisk,
			Envelopes: riskEnvelopes,
		},
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: vulnEnvelopes,
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

// resolveSources resolves source resource names from credential metadata
func resolveSources(meta CredentialSchema) ([]string, error) {
	raw := make([]string, 0, len(meta.SourceIDs))

	raw = append(raw, meta.SourceIDs...)

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
		out = lo.Map(parents, func(parent string, _ int) string {
			return fmt.Sprintf("%s/sources/-", parent)
		})
	}

	return out, nil
}
