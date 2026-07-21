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

func TestWalkAgentReadinessChecks(t *testing.T) {
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

	var failedChecks []map[string]any

	walkAgentReadinessChecks(node, "", &failedChecks)

	sort.Slice(failedChecks, func(i, j int) bool {
		return failedChecks[i]["check"].(string) < failedChecks[j]["check"].(string)
	})

	want := []map[string]any{
		{"check": "markdown", "message": "missing markdown negotiation"},
		{"check": "nested.deep", "message": "deep failure"},
	}

	assert.Check(t, is.DeepEqual(want, failedChecks))
}
