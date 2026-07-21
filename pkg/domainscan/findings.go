package domainscan

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
)

// buildFindings reports the scan's overall security verdict, any failing agent-readiness checks,
// and any expected compliance links that weren't found. result may be nil, in which case only
// the enrichment-derived missing-compliance-links finding is included
func buildFindings(result *url_scanner.ScanGetResponse, enrichment Enrichment) Findings {
	var findings Findings

	if result != nil {
		findings.SecurityViolations = result.Verdicts.Overall.Categories
		findings.Risks = result.Verdicts.Overall.Tags
		findings.IsMalicious = result.Verdicts.Overall.Malicious

		if agentReadiness := buildAgentReadinessFindings(result.Meta.Processors.AgentReadiness); agentReadiness != nil {
			findings.AgentReadiness = []AgentReadinessFinding{*agentReadiness}
		}
	}

	findings.MissingComplianceLinks = buildMissingComplianceLinks(enrichment)

	return findings
}

// expectedComplianceLinkTypes are the compliance document types a company
// site is generally expected to publish
var expectedComplianceLinkTypes = []string{"privacy_policy", "terms_of_service", "trust_center", "dpa", "security", "cookie_policy"}

// buildMissingComplianceLinks renders a GitHub-flavored Markdown task list, one unchecked
// item per expectedComplianceLinkTypes entry not found in the domainscan enrichment's
// compliance links, so it surfaces as one actionable checklist finding
func buildMissingComplianceLinks(enrichment Enrichment) string {
	if enrichment.Compliance == nil {
		return ""
	}

	found := make(map[string]bool, len(enrichment.Compliance.ComplianceLinks))
	for _, link := range enrichment.Compliance.ComplianceLinks {
		found[link.Type] = true
	}

	if enrichment.Compliance.PageType != "" {
		found[enrichment.Compliance.PageType] = true
	}

	items := make([]string, 0, len(expectedComplianceLinkTypes))

	for _, t := range expectedComplianceLinkTypes {
		if !found[t] {
			items = append(items, fmt.Sprintf("- [ ] %s", t))
		}
	}

	return strings.Join(items, "\n")
}

// buildAgentReadinessFindings reports the failing checks from the scan's agent-readiness assessment
// (e.g. missing markdown negotiation, no MCP server card)
func buildAgentReadinessFindings(processor url_scanner.ScanGetResponseMetaProcessorsAgentReadiness) *AgentReadinessFinding {
	raw := processor.JSON.RawJSON()
	if raw == "" {
		return nil
	}

	var parsed struct {
		Level     int64          `json:"level"`
		LevelName string         `json:"levelName"`
		Checks    map[string]any `json:"checks"`
	}

	if err := json.Unmarshal([]byte(raw), &parsed); err != nil || len(parsed.Checks) == 0 {
		return nil
	}

	failedChecks := []map[string]any{}
	walkAgentReadinessChecks(parsed.Checks, "", &failedChecks)

	if len(failedChecks) == 0 {
		return nil
	}

	sort.Slice(failedChecks, func(i, j int) bool {
		return failedChecks[i]["check"].(string) < failedChecks[j]["check"].(string)
	})

	return &AgentReadinessFinding{
		Level:     parsed.Level,
		LevelName: parsed.LevelName,
		Checklist: buildAgentReadinessChecklistMarkdown(failedChecks),
		Reference: agentReadinessReferenceURL,
	}
}

// agentReadinessReferenceURL links to Cloudflare's writeup of what the agent-readiness
// assessment measures and why, for context alongside the failed-check checklist
const agentReadinessReferenceURL = "https://blog.cloudflare.com/agent-readiness/"

// buildAgentReadinessChecklistMarkdown renders failedChecks as a single GitHub-flavored
// Markdown task list, one unchecked item per failing check, so the assessment surfaces
// as one finding instead of one per check
func buildAgentReadinessChecklistMarkdown(failedChecks []map[string]any) string {
	items := make([]string, 0, len(failedChecks))

	for _, c := range failedChecks {
		items = append(items, fmt.Sprintf("- [ ] %s", fmt.Sprint(c["message"])))
	}

	return strings.Join(items, "\n")
}

// walkAgentReadinessChecks recursively descends a generic agent-readiness check result
func walkAgentReadinessChecks(node map[string]any, path string, failedChecks *[]map[string]any) {
	status, hasStatus := node["status"].(string)
	message, hasMessage := node["message"].(string)

	if hasStatus && hasMessage {
		if status == "fail" {
			*failedChecks = append(*failedChecks, map[string]any{
				"check":   path,
				"message": message,
			})
		}

		return
	}

	for key, value := range node {
		child, ok := value.(map[string]any)
		if !ok {
			continue
		}

		childPath := key
		if path != "" {
			childPath = path + "." + key
		}

		walkAgentReadinessChecks(child, childPath, failedChecks)
	}
}
