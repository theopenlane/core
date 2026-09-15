package domainscan

import (
	"sort"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestBuildMissingComplianceLinks(t *testing.T) {
	tests := []struct {
		name       string
		compliance *CompliancePage
		want       string
	}{
		{
			name:       "nil compliance returns empty string",
			compliance: nil,
			want:       "",
		},
		{
			name: "found links and page type are excluded from missing",
			compliance: &CompliancePage{
				PageType: "privacy_policy",
				ComplianceLinks: []ComplianceLink{
					{URL: "https://example.com/terms", Type: "terms_of_service"},
				},
			},
			want: "- [ ] trust_center\n- [ ] dpa\n- [ ] security\n- [ ] cookie_policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMissingComplianceLinks(Enrichment{Compliance: tt.compliance})

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}

func TestBuildAgentReadinessChecklistMarkdown(t *testing.T) {
	failedChecks := []map[string]any{
		{"check": "markdown", "message": "missing markdown negotiation"},
		{"check": "mcp", "message": "no MCP server card"},
	}

	got := buildAgentReadinessChecklistMarkdown(failedChecks)

	want := "- [ ] missing markdown negotiation\n- [ ] no MCP server card"

	assert.Check(t, is.Equal(want, got))
}

func TestAgentReadinessResultFailed(t *testing.T) {
	node := map[string]any{
		"markdown": map[string]any{
			"status":  "fail",
			"message": "missing markdown negotiation",
		},
		"mcp": map[string]any{
			"status":  "pass",
			"message": "ok",
		},
		"nested": map[string]any{
			"deep": map[string]any{
				"status":  "fail",
				"message": "deep failure",
			},
		},
	}

	var checks []AgentReadinessCheck

	collectAgentReadinessChecks(node, "", &checks)
	sort.Slice(checks, func(i, j int) bool { return checks[i].Check < checks[j].Check })

	want := []AgentReadinessCheck{
		{Check: "markdown", Status: "fail", Message: "missing markdown negotiation"},
		{Check: "nested.deep", Status: "fail", Message: "deep failure"},
	}

	assert.Check(t, is.DeepEqual(want, agentReadinessResult{Checks: checks}.failed()))
}

func TestCollectAgentReadinessChecks(t *testing.T) {
	node := map[string]any{
		"discoverability": map[string]any{
			"robots": map[string]any{
				"status":  "pass",
				"message": "robots.txt present",
			},
			"sitemap": map[string]any{
				"status":  "fail",
				"message": "no sitemap.xml",
			},
		},
		"capabilities": map[string]any{
			"mcp": map[string]any{
				"status":  "fail",
				"message": "no MCP server card",
			},
		},
		"notACheck": "ignored scalar",
	}

	var checks []AgentReadinessCheck

	collectAgentReadinessChecks(node, "", &checks)

	sort.Slice(checks, func(i, j int) bool { return checks[i].Check < checks[j].Check })

	want := []AgentReadinessCheck{
		{Check: "capabilities.mcp", Status: "fail", Message: "no MCP server card"},
		{Check: "discoverability.robots", Status: "pass", Message: "robots.txt present"},
		{Check: "discoverability.sitemap", Status: "fail", Message: "no sitemap.xml"},
	}

	assert.DeepEqual(t, checks, want)
}

func TestCollectAgentReadinessChecksIgnoresIncompleteLeaves(t *testing.T) {
	node := map[string]any{
		"statusOnly":  map[string]any{"status": "pass"},
		"messageOnly": map[string]any{"message": "orphan"},
		"real":        map[string]any{"status": "fail", "message": "a real check"},
	}

	var checks []AgentReadinessCheck

	collectAgentReadinessChecks(node, "", &checks)

	assert.Equal(t, len(checks), 1)
	assert.Equal(t, checks[0].Check, "real")
}

func TestAgentReadinessResultFailedReportsOnlyFailures(t *testing.T) {
	node := map[string]any{
		"a": map[string]any{"status": "pass", "message": "fine"},
		"b": map[string]any{"status": "fail", "message": "broken"},
	}

	var checks []AgentReadinessCheck

	collectAgentReadinessChecks(node, "", &checks)

	failures := agentReadinessResult{Checks: checks}.failed()

	assert.Assert(t, is.Len(failures, 1))
	assert.Equal(t, failures[0].Check, "b")
	assert.Equal(t, failures[0].Message, "broken")
}

// the Markdown checklist is what the console renders, so it stays one unchecked item per
// failing check regardless of how the failures were collected
func TestBuildAgentReadinessChecklistMarkdown(t *testing.T) {
	got := buildAgentReadinessChecklistMarkdown([]AgentReadinessCheck{
		{Check: "markdown", Status: "fail", Message: "missing markdown negotiation"},
		{Check: "nested.deep", Status: "fail", Message: "deep failure"},
	})

	assert.Equal(t, got, "- [ ] missing markdown negotiation\n- [ ] deep failure")
}
