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

// expectedComplianceLinkTypes are the compliance document types a company site is generally
// expected to publish.
//
// "security" is deliberately not among them. A security page and a trust center are the same
// document for most companies, so listing both asked a company that publishes one to publish
// it again, and the checklist item read as a different requirement than it was
var expectedComplianceLinkTypes = []string{"privacy_policy", "terms_of_service", "trust_center", "dpa", "cookie_policy"}

// buildMissingComplianceLinks renders a GitHub-flavored Markdown task list, one unchecked
// item per expectedComplianceLinkTypes entry not found in the domainscan enrichment's
// compliance links, so it surfaces as one actionable checklist finding
func buildMissingComplianceLinks(enrichment Enrichment) string {
	if enrichment.Compliance == nil {
		return ""
	}

	// types are normalized before comparison: the extraction is constrained to the canonical
	// set now, but a model that answers "Privacy Policy" or "tos" has found the document just
	// the same, and reporting it as missing sends the reader looking for a page they published
	found := make(map[string]bool, len(enrichment.Compliance.ComplianceLinks))
	for _, link := range enrichment.Compliance.ComplianceLinks {
		if normalized := NormalizeComplianceType(link.Type); normalized != "" {
			found[normalized] = true
		}
	}

	if normalized := NormalizeComplianceType(enrichment.Compliance.PageType); normalized != "" {
		found[normalized] = true
	}

	items := make([]string, 0, len(expectedComplianceLinkTypes))

	for _, t := range expectedComplianceLinkTypes {
		if !found[t] {
			items = append(items, fmt.Sprintf("- [ ] %s", t))
		}
	}

	return strings.Join(items, "\n")
}

// agentReadinessResult is the scan's agent-readiness assessment, parsed once from the
// processor payload. Both report shapes are derived from it rather than each unmarshalling and
// sorting the same JSON again
type agentReadinessResult struct {
	// Level is the assessment's score
	Level int64
	// LevelName is the human-readable name for Level
	LevelName string
	// Checks are every check that reported a status, ordered by check path
	Checks []AgentReadinessCheck
}

// parseAgentReadiness decodes the processor payload and flattens its nested check tree. ok is
// false when the processor reported nothing, which is different from reporting no failures
func parseAgentReadiness(processor url_scanner.ScanGetResponseMetaProcessorsAgentReadiness) (agentReadinessResult, bool) {
	raw := processor.JSON.RawJSON()
	if raw == "" {
		return agentReadinessResult{}, false
	}

	var parsed struct {
		Level     int64          `json:"level"`
		LevelName string         `json:"levelName"`
		Checks    map[string]any `json:"checks"`
	}

	if err := json.Unmarshal([]byte(raw), &parsed); err != nil || len(parsed.Checks) == 0 {
		return agentReadinessResult{}, false
	}

	var checks []AgentReadinessCheck

	collectAgentReadinessChecks(parsed.Checks, "", &checks)

	if len(checks) == 0 {
		return agentReadinessResult{}, false
	}

	sort.Slice(checks, func(i, j int) bool { return checks[i].Check < checks[j].Check })

	return agentReadinessResult{Level: parsed.Level, LevelName: parsed.LevelName, Checks: checks}, true
}

// failed returns the checks that did not pass, in check-path order
func (r agentReadinessResult) failed() []AgentReadinessCheck {
	failures := make([]AgentReadinessCheck, 0, len(r.Checks))

	for _, check := range r.Checks {
		if check.Status == agentReadinessStatusFail {
			failures = append(failures, check)
		}
	}

	return failures
}

// buildAgentReadinessFindings reports the failing checks from the scan's agent-readiness
// assessment, as one finding carrying a Markdown checklist. A domain with no failures raises
// no finding
func buildAgentReadinessFindings(processor url_scanner.ScanGetResponseMetaProcessorsAgentReadiness) *AgentReadinessFinding {
	result, ok := parseAgentReadiness(processor)
	if !ok {
		return nil
	}

	failures := result.failed()
	if len(failures) == 0 {
		return nil
	}

	return &AgentReadinessFinding{
		Level:     result.Level,
		LevelName: result.LevelName,
		Checklist: buildAgentReadinessChecklistMarkdown(failures),
		Reference: agentReadinessReferenceURL,
	}
}

// agentReadinessReferenceURL links to Cloudflare's writeup of what the agent-readiness
// assessment measures and why, for context alongside the failed-check checklist
const agentReadinessReferenceURL = "https://blog.cloudflare.com/agent-readiness/"

// buildAgentReadinessChecklistMarkdown renders checks as a single GitHub-flavored Markdown
// task list, one unchecked item each, so the assessment surfaces as one finding rather than
// one per check
func buildAgentReadinessChecklistMarkdown(checks []AgentReadinessCheck) string {
	items := make([]string, 0, len(checks))

	for _, check := range checks {
		items = append(items, fmt.Sprintf("- [ ] %s", check.Message))
	}

	return strings.Join(items, "\n")
}

// agentReadinessStatusFail is the status Cloudflare reports for a check that did not pass
const agentReadinessStatusFail = "fail"

// agentReadinessStatusPass is the status Cloudflare reports for a check that passed
const agentReadinessStatusPass = "pass"

// collectAgentReadinessChecks recursively descends a generic agent-readiness check result,
// appending every leaf that reports a status. A leaf is a node carrying both a status and a
// message; everything else is treated as a container and descended into
func collectAgentReadinessChecks(node map[string]any, path string, checks *[]AgentReadinessCheck) {
	status, hasStatus := node["status"].(string)
	message, hasMessage := node["message"].(string)

	if hasStatus && hasMessage {
		*checks = append(*checks, AgentReadinessCheck{
			Check:   path,
			Status:  status,
			Message: message,
		})

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

		collectAgentReadinessChecks(child, childPath, checks)
	}
}

// buildAgentReadinessAssessment reports the full agent-readiness assessment, including the
// checks that passed, so a domain with a perfect score is distinguishable from one the
// processor said nothing about
func buildAgentReadinessAssessment(processor url_scanner.ScanGetResponseMetaProcessorsAgentReadiness) *AgentReadinessAssessment {
	result, ok := parseAgentReadiness(processor)
	if !ok {
		return nil
	}

	assessment := &AgentReadinessAssessment{
		Level:       result.Level,
		LevelName:   result.LevelName,
		Reference:   agentReadinessReferenceURL,
		TotalChecks: len(result.Checks),
		Checks:      result.Checks,
	}

	for _, check := range result.Checks {
		switch check.Status {
		case agentReadinessStatusPass:
			assessment.PassedChecks++
		case agentReadinessStatusFail:
			assessment.FailedChecks++
		}
	}

	return assessment
}
