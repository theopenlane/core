package graphapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/generated"
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
	require.Len(t, contexts, 1)
	assert.Equal(t, "entity-123", contexts[0].EntityID)
	assert.Equal(t, "Control", contexts[0].EntityType)
	assert.Contains(t, contexts[0].MatchedFields, "Title")
	assert.Contains(t, contexts[0].MatchedFields, "Description")
	assert.NotEmpty(t, contexts[0].Snippets)
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
	assert.Contains(t, matches, "Title")
	assert.Contains(t, matches, "Description")

	// but RefCode doesn't contain "policy"
	assert.NotContains(t, matches, "RefCode")
}

func TestCreateSnippet(t *testing.T) {
	tracker := newContextTracker("security")

	// test snippet creation with match in middle
	snippet := tracker.createSnippet("Description", "This is a very important security control that protects the system", "security")
	require.NotNil(t, snippet)
	assert.Equal(t, "Description", snippet.Field)

	// snippet should contain the matched text and context
	assert.Contains(t, snippet.Text, "security")
	assert.Contains(t, snippet.Text, "important")

	// test snippet with match at beginning
	snippet2 := tracker.createSnippet("Title", "Security Policy", "security")
	require.NotNil(t, snippet2)
	assert.Contains(t, snippet2.Text, "Security")
}

func TestGetEntityID(t *testing.T) {
	control := &generated.Control{
		ID:    "ctrl-123",
		Title: "Test Control",
	}

	entityID := getEntityID(control)
	assert.Equal(t, "ctrl-123", entityID)

	// test with nil
	entityID = getEntityID(nil)
	assert.Empty(t, entityID)
}

func TestExtractSnippets(t *testing.T) {
	tracker := newContextTracker("test")

	entity := &generated.ActionPlan{
		ID:      "plan-123",
		Name:    "Test Plan",
		Details: "This is a test plan for testing the search functionality",
	}

	snippets := tracker.extractSnippets(entity, []string{"Name", "Details"})
	assert.NotEmpty(t, snippets)

	// should have snippets for both Name and Details
	var nameSnippetExists, detailSnippetExists bool

	for _, snippet := range snippets {
		if snippet.Field == "Name" {
			nameSnippetExists = true
			assert.Contains(t, snippet.Text, "Test")
		}
		if snippet.Field == "Details" {
			detailSnippetExists = true
			assert.Contains(t, snippet.Text, "test")
		}
	}

	assert.True(t, nameSnippetExists, "Should have snippet for Name field")
	assert.True(t, detailSnippetExists, "Should have snippet for Details field")
}
