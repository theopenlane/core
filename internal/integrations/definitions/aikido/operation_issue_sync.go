package aikido

import (
	"context"
	"fmt"
	"slices"

	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/mapx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// issueTypeOpenSource is the Aikido issue type for vulnerable dependencies, the only type ingested as a vulnerability
const issueTypeOpenSource = "open_source"

// issueGroupKey is the issue payload key the matching issue group is attached under
const issueGroupKey = "group"

const (
	// issueTypeAIPentest is the Aikido issue type for AI pentest findings, which carry an attack analysis
	issueTypeAIPentest = "ai_pentest"
	// attackAnalysisKey is the issue payload key the pentest attack analysis is attached under
	attackAnalysisKey = "attack_analysis"
	// cveDetailsKey is the issue payload key the Aikido Intel CVE details are attached under
	cveDetailsKey = "cve_details"
)

const (
	// findingVariantDataleak is the finding mapping variant for leaked credentials found for a monitored domain
	findingVariantDataleak = "dataleak"
	// dataleakPasswordKey is the partially redacted leaked password, dropped so it is never stored
	dataleakPasswordKey = "password"
)

// issueResourceKeys are the issue fields naming the scanned resource, in the order they are preferred
var issueResourceKeys = []string{"code_repo_name", "container_repo_name", "cloud_name", "domain_name", "virtual_machine_name"}

// runIssueSync collects Aikido issues for vulnerability and finding ingestion
func runIssueSync(ctx context.Context, _ types.OperationRequest, client *APIClient, _ IssueSync) ([]types.IngestPayloadSet, error) {
	return CollectIssues(ctx, client)
}

// CollectIssues routes open source dependency issues to the vulnerability schema and every other issue type to the finding schema
func CollectIssues(ctx context.Context, c *APIClient) ([]types.IngestPayloadSet, error) {
	issues, err := c.exportIssues(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIssuesFetchFailed, err)
	}

	groups, err := c.issueGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIssueGroupsFetchFailed, err)
	}

	// bulk details must be enabled by Aikido support, so a failure keeps the export fields only
	details, err := c.issueDetails(ctx, issueIDs(issues))
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		logx.FromContext(ctx).Warn().Err(err).Msg("aikido: bulk issue details unavailable, ingesting export fields only")
	}

	// attack analyses need pentests:read, so a failure ingests pentest issues without them
	analyses, err := c.fetchEach(ctx, pentestIssueIDs(issues), pathAttackAnalysisFormat)
	if err != nil {
		return nil, err
	}

	// cve details need research:read, so a failure keeps the issue group description
	cves, err := c.fetchEach(ctx, issueCVEs(issues), pathCVEFormat)
	if err != nil {
		return nil, err
	}

	vulnerabilities := make([]types.MappingEnvelope, 0)
	findings := make([]types.MappingEnvelope, 0)

	for _, issue := range issues {
		// bulk details only fill keys the export lacks, so every export value wins
		issue = mapx.DeepMergeMapAny(details[mapx.Value[float64](issue, "id")], issue)
		attachGroup(issue, groups)

		if analysis, ok := analyses[recordID(issue)]; ok {
			issue[attackAnalysisKey] = analysis
		}

		if cve, ok := cves[mapx.Value[string](issue, "cve_id")]; ok {
			issue[cveDetailsKey] = cve
		}

		// orValue does not replace an explicit null, so nulls are pruned before mapping
		envelope, err := providerkit.MarshalEnvelope(issueResource(issue), mapx.PruneMapZeroAny(issue), ErrPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Any("issue_id", issue["id"]).Msg("aikido: failed to marshal issue envelope")

			return nil, err
		}

		if mapx.Value[string](issue, "type") == issueTypeOpenSource {
			vulnerabilities = append(vulnerabilities, envelope)

			continue
		}

		findings = append(findings, envelope)
	}

	leaks, err := collectDataleaks(ctx, c)
	if err != nil {
		return nil, err
	}

	findings = append(findings, leaks...)

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaVulnerability.Name,
			Envelopes: vulnerabilities,
		},
		{
			Schema:    entityops.SchemaFinding.Name,
			Envelopes: findings,
		},
	}, nil
}

// attachGroup nests the issue's group under the group key so mappings can read its description and remediation
func attachGroup(issue map[string]any, groups map[float64]map[string]any) {
	if group, ok := groups[mapx.Value[float64](issue, "group_id")]; ok {
		issue[issueGroupKey] = group
	}
}

// collectDataleaks returns a finding envelope per leaked credential across the monitored domains, a listing failure
// is logged and skipped since dataleaks need the domains:read scope older installations may not have granted
func collectDataleaks(ctx context.Context, c *APIClient) ([]types.MappingEnvelope, error) {
	domains, err := c.listAll(ctx, pathDomains, nil, assetPageSize)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		logx.FromContext(ctx).Warn().Err(err).Msg("aikido: unable to list domains, skipping dataleaks")

		return nil, nil
	}

	envelopes := make([]types.MappingEnvelope, 0)

	for _, domain := range domains {
		domainID := recordID(domain)

		leaks, err := c.domainDataleaks(ctx, domainID)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			logx.FromContext(ctx).Warn().Err(err).Str("domain_id", domainID).Msg("aikido: unable to list dataleaks for domain, skipping it")

			continue
		}

		for _, leak := range leaks {
			delete(leak, dataleakPasswordKey)

			envelope, err := providerkit.MarshalEnvelopeVariant(findingVariantDataleak, mapx.Value[string](domain, "domain"), mapx.PruneMapZeroAny(leak), ErrPayloadEncode)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Any("dataleak_id", leak["id"]).Msg("aikido: failed to marshal dataleak envelope")

				return nil, err
			}

			envelopes = append(envelopes, envelope)
		}
	}

	return envelopes, nil
}

// openIssueStatuses are the issue statuses still needing action, matching the open mapping, only these are enriched
// with per issue calls so each sync stays within the rate limit
var openIssueStatuses = []string{"open", "snoozed"}

// isOpenIssue reports whether the issue still needs action
func isOpenIssue(issue map[string]any) bool {
	return slices.Contains(openIssueStatuses, mapx.Value[string](issue, "status"))
}

// pentestIssueIDs returns the ids of the open AI pentest issues, which carry an attack analysis
func pentestIssueIDs(issues []map[string]any) []string {
	ids := make([]string, 0)

	for _, issue := range issues {
		if mapx.Value[string](issue, "type") == issueTypeAIPentest && isOpenIssue(issue) {
			ids = append(ids, recordID(issue))
		}
	}

	return ids
}

// issueCVEs returns the distinct CVE or AIKIDO ids of the open open source issues
func issueCVEs(issues []map[string]any) []string {
	seen := map[string]struct{}{}
	cves := make([]string, 0)

	for _, issue := range issues {
		cve := mapx.Value[string](issue, "cve_id")
		if cve == "" || mapx.Value[string](issue, "type") != issueTypeOpenSource || !isOpenIssue(issue) {
			continue
		}

		if _, ok := seen[cve]; ok {
			continue
		}

		seen[cve] = struct{}{}
		cves = append(cves, cve)
	}

	return cves
}

// issueIDs returns the ids of the exported issues as the bulk detail endpoint expects them
func issueIDs(issues []map[string]any) []string {
	ids := make([]string, 0, len(issues))

	for _, issue := range issues {
		if mapx.Value[float64](issue, "id") != 0 {
			ids = append(ids, recordID(issue))
		}
	}

	return ids
}

// issueResource returns the name of the resource an issue was found in, matching the asset name ingested for it
func issueResource(issue map[string]any) string {
	for _, key := range issueResourceKeys {
		if name := mapx.Value[string](issue, key); name != "" {
			return name
		}
	}

	return ""
}
