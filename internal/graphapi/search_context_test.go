package graphapi

import (
	"slices"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func TestSearchContextTracker(t *testing.T) {
	query := "test"
	tracker := newContextTracker(query)

	// test adding a match
	tracker.addMatch("entity-123", "Control", []string{"Title", "Description"}, &generated.Control{
		ID:          "entity-123",
		Title:       "Test Control",
		Description: "This is a test control for testing purposes",
	})

	contexts := tracker.getContexts()
	assert.Assert(t, contexts != nil)
	assert.Assert(t, is.Len(contexts, 1))
	assert.Check(t, is.Equal("entity-123", contexts[0].EntityID))
	assert.Check(t, is.Equal("Control", contexts[0].EntityType))
	assert.Check(t, is.Contains(contexts[0].MatchedFields, "Title"))
	assert.Check(t, is.Contains(contexts[0].MatchedFields, "Description"))
	assert.Check(t, len(contexts[0].Snippets) >= 1)
}

func TestFieldMatchChecker(t *testing.T) {
	checker := fieldMatchChecker{"policy"}

	control := &generated.Control{
		ID:          "ctrl-123",
		Title:       "Security Policy Control",
		Description: "Ensures the security policy is followed",
	}

	// check which fields match
	matches := checker.check(control, []string{"Title", "Description", "RefCode"})

	// both Title and Description contains "policy"
	assert.Check(t, is.Contains(matches, "Title"))
	assert.Check(t, is.Contains(matches, "Description"))

	// but RefCode doesn't contain "policy"
	assert.Check(t, !slices.Contains(matches, "RefCode"))
}

func TestCreateSnippet(t *testing.T) {
	tracker := newContextTracker("security")

	// test snippet creation with match in middle
	snippet := tracker.createSnippet("Description", "This is a very important security control that protects the system", "security")
	assert.Assert(t, snippet != nil)
	assert.Check(t, is.Equal("Description", snippet.Field))

	// snippet should contain the matched text and context
	assert.Check(t, is.Contains(snippet.Text, "security"))
	assert.Check(t, is.Contains(snippet.Text, "important"))

	// test snippet with match at beginning
	snippet2 := tracker.createSnippet("Title", "Security Policy", "security")
	assert.Assert(t, snippet2 != nil)
	assert.Check(t, is.Contains(snippet2.Text, "Security"))
}

func TestExtractSnippets(t *testing.T) {
	tracker := newContextTracker("test")

	entity := &generated.ActionPlan{
		ID:      "plan-123",
		Name:    "Test Plan",
		Details: "This is a test plan for testing the search functionality",
	}

	snippets := tracker.extractSnippets(entity, []string{"Name", "Details"})
	assert.Assert(t, snippets != nil)

	// should have snippets for both Name and Details
	var nameSnippetExists, detailSnippetExists bool

	for _, snippet := range snippets {
		if snippet.Field == "Name" {
			nameSnippetExists = true
			assert.Check(t, is.Contains(snippet.Text, "Test"))
		}
		if snippet.Field == "Details" {
			detailSnippetExists = true
			assert.Check(t, is.Contains(snippet.Text, "test"))
		}
	}

	assert.Check(t, is.Equal(true, nameSnippetExists), "Should have snippet for Name field")
	assert.Check(t, is.Equal(true, detailSnippetExists), "Should have snippet for Details field")
}

func TestSearchContextContentSanitization(t *testing.T) {
	tt := []struct {
		name         string
		query        string
		details      string
		wantSnippet  string
		wantMatch    bool
		wantSanitize string
	}{
		{
			name:  "html attributes ignored",
			query: "slate",
			details: `<span data-slate-node="text">` +
				`<span data-slate-string="true">Visible objective</span></span>`,
			wantMatch: false,
		},
		{
			name:  "html stripped",
			query: "objectives",
			details: `<span class="slate-bold"><span data-slate-string="true">objectives</span></span>` +
				`<span data-slate-node="text"> and responsibilities</span>`,
			wantSnippet: "objectives and responsibilities",
			wantMatch:   true,
		},
		{
			name:         "markdown stripped",
			details:      "## Recovery Objectives\n\n- [Objective owner](https://example.com)\n- **monitoring**",
			wantSanitize: "Recovery Objectives Objective owner monitoring",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantSanitize != "" {
				assert.Check(t, is.Equal(tc.wantSanitize, sanitizeContent(tc.details)))
				return
			}

			entity := &generated.ActionPlan{
				ID:      "plan-123",
				Details: tc.details,
			}

			checker := fieldMatchChecker{query: tc.query}
			matches := checker.check(entity, []string{"Details"})
			assert.Check(t, is.Equal(tc.wantMatch, slices.Contains(matches, "Details")))

			if tc.wantSnippet == "" {
				return
			}

			snippets := newContextTracker(tc.query).extractSnippets(entity, []string{"Details"})
			assert.Assert(t, is.Len(snippets, 1))
			assert.Check(t, is.Equal(tc.wantSnippet, snippets[0].Text))
			assert.Check(t, !strings.Contains(snippets[0].Text, "<span"))
			assert.Check(t, !strings.Contains(snippets[0].Text, "data-slate"))
		})
	}
}
