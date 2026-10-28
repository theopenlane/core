package gcpscc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"github.com/samber/lo"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

const (
	// findingsPageSize is the default number of SCC findings requested per paginated API call
	findingsPageSize = 100
	// findingsMaxPageSize is the maximum number of findings that can be requested per API page
	findingsMaxPageSize = 1000
	// findingConsoleURLBase is the SCC console page for a single finding, followed by the escaped finding name
	findingConsoleURLBase = "https://console.cloud.google.com/security/command-center/findingsv2;name="
	// cvssv31Prefix is the version prefix for CVSS v3.1 vector strings
	cvssv31Prefix = "CVSS:3.1"
	// cvssVectorKey is the payload key holding the CVSS vector built from the SCC cvssv3 metrics
	cvssVectorKey = "cvss_vector"
)

// runFindingsCollect collects GCP SCC findings from configured sources
func runFindingsCollect(ctx context.Context, request types.OperationRequest, c Client, cfg FindingsSync) ([]types.IngestPayloadSet, error) {
	sources, err := resolveSources(c.Scope)
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

	marshaler := protojson.MarshalOptions{UseProtoNames: true}
	findingEnvelopes := make([]types.MappingEnvelope, 0)
	vulnEnvelopes := make([]types.MappingEnvelope, 0)
	riskEnvelopes := make([]types.MappingEnvelope, 0)

	var timeFilter string
	if request.LastRunAt != nil {
		timeFilter = fmt.Sprintf(`event_time >= "%s"`, request.LastRunAt.UTC().Format(time.RFC3339))
	}

	for _, sourceName := range sources {
		pageResults, err := listAllFindings(ctx, c.Client, sourceName, pageSize, timeFilter)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("gcpscc: error listing findings")
			return nil, ErrListFindingsFailed
		}

		for _, result := range pageResults {
			finding := result.GetFinding()
			if finding == nil {
				continue
			}

			finding.ExternalUri = findingConsoleURL(finding, result.GetResource())

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
		}
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaFinding.Name,
			Envelopes: findingEnvelopes,
		},
		{
			Schema:    entityops.SchemaRisk.Name,
			Envelopes: riskEnvelopes,
		},
		{
			Schema:    entityops.SchemaVulnerability.Name,
			Envelopes: vulnEnvelopes,
		},
	}, nil
}

// listAllFindings pages through all findings for a single source
func listAllFindings(ctx context.Context, c *cloudscc.Client, sourceName string, pageSize int, timeFilter string) ([]*securitycenterpb.ListFindingsResponse_ListFindingsResult, error) {
	req := &securitycenterpb.ListFindingsRequest{
		PageSize: int32(min(pageSize, math.MaxInt32)), //nolint:gosec // bounds checked via min
		Filter:   timeFilter,
	}

	if sourceName != "" {
		req.Parent = sourceName
	}

	it := c.ListFindings(ctx, req)

	var results []*securitycenterpb.ListFindingsResponse_ListFindingsResult

	for {
		result, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// buildFindingEnvelope serializes a finding into a mapping envelope
func buildFindingEnvelope(sourceName string, finding *securitycenterpb.Finding, marshaler protojson.MarshalOptions) (types.MappingEnvelope, error) {
	rawPayload, err := marshaler.Marshal(finding)
	if err != nil {
		return types.MappingEnvelope{}, ErrFindingEncode
	}

	rawPayload, err = withCVSSVector(rawPayload, cvssv31Vector(finding.GetVulnerability().GetCve().GetCvssv3()))
	if err != nil {
		return types.MappingEnvelope{}, ErrFindingEncode
	}

	return providerkit.RawEnvelope(resolveFindingResource(sourceName, finding), rawPayload), nil
}

// withCVSSVector adds the CVSS vector string to a serialized finding payload
func withCVSSVector(rawPayload []byte, vector string) ([]byte, error) {
	if vector == "" {
		return rawPayload, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, err
	}

	payload[cvssVectorKey] = vector

	return json.Marshal(payload)
}

// cvssv31Vector builds a CVSS v3.1 vector string from the SCC cvssv3 metrics, empty when any metric is unspecified
func cvssv31Vector(cvss *securitycenterpb.Cvssv3) string {
	metrics := []struct {
		key   string
		value string
	}{
		{key: "AV", value: cvssMetric(cvss.GetAttackVector().String(), "ATTACK_VECTOR_")},
		{key: "AC", value: cvssMetric(cvss.GetAttackComplexity().String(), "ATTACK_COMPLEXITY_")},
		{key: "PR", value: cvssMetric(cvss.GetPrivilegesRequired().String(), "PRIVILEGES_REQUIRED_")},
		{key: "UI", value: cvssMetric(cvss.GetUserInteraction().String(), "USER_INTERACTION_")},
		{key: "S", value: cvssMetric(cvss.GetScope().String(), "SCOPE_")},
		{key: "C", value: cvssMetric(cvss.GetConfidentialityImpact().String(), "IMPACT_")},
		{key: "I", value: cvssMetric(cvss.GetIntegrityImpact().String(), "IMPACT_")},
		{key: "A", value: cvssMetric(cvss.GetAvailabilityImpact().String(), "IMPACT_")},
	}

	parts := []string{cvssv31Prefix}

	for _, metric := range metrics {
		if metric.value == "" {
			return ""
		}

		parts = append(parts, metric.key+":"+metric.value)
	}

	return strings.Join(parts, "/")
}

// cvssMetric returns the CVSS code for an SCC metric enum, the first letter after its prefix
func cvssMetric(enumName, prefix string) string {
	value, ok := strings.CutPrefix(enumName, prefix)
	if !ok || value == "" || value == "UNSPECIFIED" {
		return ""
	}

	return value[:1]
}

// findingConsoleURL builds the SCC console deep link for a finding
func findingConsoleURL(finding *securitycenterpb.Finding, resource *securitycenterpb.ListFindingsResponse_ListFindingsResult_Resource) string {
	name := finding.GetName()
	if name == "" {
		return finding.GetExternalUri()
	}

	consoleURL := findingConsoleURLBase + url.PathEscape(name)

	if project := resource.GetGcpMetadata().GetProjectDisplayName(); project != "" {
		return consoleURL + "?" + url.Values{"project": {project}}.Encode()
	}

	if rest, ok := strings.CutPrefix(name, organizationParentPrefix); ok {
		orgID, _, _ := strings.Cut(rest, "/")

		return consoleURL + "?" + url.Values{"organizationId": {orgID}}.Encode()
	}

	return consoleURL
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
func resolveParents(meta CollectionScope) ([]string, error) {
	if meta.OrganizationID != "" && meta.ProjectScope != projectScopeSpecific {
		return []string{organizationParent(meta.OrganizationID)}, nil
	}

	if meta.ProjectScope == projectScopeSpecific {
		parentList := lo.FilterMap(meta.ProjectIDs, func(projectID string, _ int) (string, bool) {
			value := strings.TrimSpace(projectID)
			if value == "" {
				return "", false
			}

			return projectParent(value), true
		})

		parentList = lo.Uniq(parentList)

		if len(parentList) == 0 {
			return nil, ErrProjectIDRequired
		}

		return parentList, nil
	}

	if meta.ProjectID != "" {
		return []string{projectParent(meta.ProjectID)}, nil
	}

	if meta.OrganizationID != "" {
		return []string{organizationParent(meta.OrganizationID)}, nil
	}

	return nil, ErrProjectIDRequired
}

// resolveSources resolves source resource names from the collection scope
func resolveSources(meta CollectionScope) ([]string, error) {
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
		case strings.HasPrefix(source, organizationParentPrefix), strings.HasPrefix(source, projectParentPrefix):
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
