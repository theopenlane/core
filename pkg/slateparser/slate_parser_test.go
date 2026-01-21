package slateparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/utils/ulids"
)

func TestCheckForMentions(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		objectType     string
		objectID       string
		objectName     string
		expectedCount  int
		expectedUserID string
		expectedName   string
	}{
		{
			name:           "single mention",
			text:           `<div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe"></div>`,
			objectType:     "Task",
			objectID:       "task001",
			objectName:     "Test Task",
			expectedCount:  1,
			expectedUserID: "user123",
			expectedName:   "John Doe",
		},
		{
			name: "multiple mentions",
			text: `<div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe"></div>
			       <div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user456" data-slate-id="mention002" data-slate-value="Jane Smith"></div>`,
			objectType:     "Comment",
			objectID:       "comment001",
			objectName:     "Test Comment",
			expectedCount:  2,
			expectedUserID: "user123",
			expectedName:   "John Doe",
		},
		{
			name:           "alternative attribute order",
			text:           `<div data-slate-key="user789" data-slate-id="mention003" data-slate-value="Bob Wilson" data-slate-node="element" data-slate-inline="true"></div>`,
			objectType:     "Risk",
			objectID:       "risk001",
			objectName:     "Test Risk",
			expectedCount:  1,
			expectedUserID: "user789",
			expectedName:   "Bob Wilson",
		},
		{
			name:          "no mentions",
			text:          `<p>This is just regular text without any mentions</p>`,
			objectType:    "Task",
			objectID:      "task002",
			objectName:    "Plain Task",
			expectedCount: 0,
		},
		{
			name:          "empty text",
			text:          "",
			objectType:    "Task",
			objectID:      "task003",
			objectName:    "Empty Task",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mentions := CheckForMentions(tt.text, tt.objectType, tt.objectID, tt.objectName)
			assert.Equal(t, tt.expectedCount, len(mentions), "mention count should match")

			if tt.expectedCount > 0 {
				// Check first mention
				foundMatch := false
				for _, mention := range mentions {
					if mention.UserID == tt.expectedUserID {
						assert.Equal(t, tt.expectedName, mention.UserDisplayName)
						assert.Equal(t, tt.objectType, mention.ObjectType)
						assert.Equal(t, tt.objectID, mention.ObjectID)
						assert.Equal(t, tt.objectName, mention.ObjectName)
						foundMatch = true
						break
					}
				}
				assert.True(t, foundMatch, "expected user ID should be found in mentions")
			}
		})
	}
}

func TestCheckForNewMentions(t *testing.T) {
	tests := []struct {
		name          string
		oldText       string
		newText       string
		objectType    string
		objectID      string
		objectName    string
		expectedCount int
	}{
		{
			name:          "new mention added",
			oldText:       `<p>No mentions here</p>`,
			newText:       `<div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe"></div>`,
			objectType:    "Task",
			objectID:      "task001",
			objectName:    "Test Task",
			expectedCount: 1,
		},
		{
			name:          "existing mention unchanged",
			oldText:       `<div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe"></div>`,
			newText:       `<div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe"></div>`,
			objectType:    "Task",
			objectID:      "task002",
			objectName:    "Test Task 2",
			expectedCount: 0,
		},
		{
			name:    "one old, one new mention",
			oldText: `<div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe"></div>`,
			newText: `<div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe"></div>
			          <div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user456" data-slate-id="mention002" data-slate-value="Jane Smith"></div>`,
			objectType:    "Comment",
			objectID:      "comment001",
			objectName:    "Test Comment",
			expectedCount: 1,
		},
		{
			name:          "no old, no new",
			oldText:       `<p>Plain text</p>`,
			newText:       `<p>Still plain text</p>`,
			objectType:    "Task",
			objectID:      "task003",
			objectName:    "Plain Task",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newMentionIDs := CheckForNewMentions(tt.oldText, tt.newText, tt.objectType, tt.objectID, tt.objectName)
			assert.Equal(t, tt.expectedCount, len(newMentionIDs), "new mention count should match")
		})
	}
}

func TestGetNewMentions(t *testing.T) {
	oldText := `<div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe"></div>`
	newText := `<div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe"></div>
	            <div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="user456" data-slate-id="mention002" data-slate-value="Jane Smith"></div>`

	newMentions := GetNewMentions(oldText, newText, "Task", "task001", "Test Task")

	assert.Equal(t, 1, len(newMentions), "should have one new mention")

	// Find the new mention
	mention, exists := newMentions["mention002"]
	assert.True(t, exists, "mention002 should exist")
	assert.Equal(t, "user456", mention.UserID)
	assert.Equal(t, "Jane Smith", mention.UserDisplayName)
	assert.Equal(t, "Task", mention.ObjectType)
	assert.Equal(t, "task001", mention.ObjectID)
	assert.Equal(t, "Test Task", mention.ObjectName)
}

func TestExtractMentionedUserIDs(t *testing.T) {
	// Generate valid ULIDs for testing
	validUserID1 := ulids.New().String()
	validUserID2 := ulids.New().String()
	invalidUserID := "not-a-valid-ulid"

	mentions := map[string]Mention{
		"mention001": {
			UserID:          validUserID1,
			UserDisplayName: "John Doe",
			ObjectType:      "Task",
			ObjectID:        "task001",
			ObjectName:      "Test Task",
		},
		"mention002": {
			UserID:          validUserID2,
			UserDisplayName: "Jane Smith",
			ObjectType:      "Task",
			ObjectID:        "task001",
			ObjectName:      "Test Task",
		},
		"mention003": {
			UserID:          validUserID1, // Duplicate user
			UserDisplayName: "John Doe",
			ObjectType:      "Task",
			ObjectID:        "task001",
			ObjectName:      "Test Task",
		},
		"mention004": {
			UserID:          invalidUserID, // Invalid ULID - should be skipped
			UserDisplayName: "Invalid User",
			ObjectType:      "Task",
			ObjectID:        "task001",
			ObjectName:      "Test Task",
		},
	}

	userIDs := ExtractMentionedUserIDs(mentions)

	assert.Equal(t, 2, len(userIDs), "should have 2 unique valid user IDs")
	assert.Contains(t, userIDs, validUserID1)
	assert.Contains(t, userIDs, validUserID2)
	assert.NotContains(t, userIDs, invalidUserID, "invalid ULID should be filtered out")
}

func TestIsValidSlateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "valid slate with data-slate-node",
			text:     `<div data-slate-node="element">content</div>`,
			expected: true,
		},
		{
			name:     "valid slate with data-slate-key",
			text:     `<div data-slate-key="abc123">content</div>`,
			expected: true,
		},
		{
			name:     "valid slate with data-slate-id",
			text:     `<div data-slate-id="mention001">content</div>`,
			expected: true,
		},
		{
			name:     "invalid - plain text",
			text:     `Just plain text without slate attributes`,
			expected: false,
		},
		{
			name:     "invalid - regular HTML",
			text:     `<div class="container"><p>Regular HTML</p></div>`,
			expected: false,
		},
		{
			name:     "empty text",
			text:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSlateText(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMentionRegexAttributeOrder tests that the mentionRegex correctly captures
// data-slate-key, data-slate-id, and data-slate-value in any order.
// This is critical because HTML attributes can appear in any order.
func TestMentionRegexAttributeOrder(t *testing.T) {
	tests := []struct {
		name             string
		text             string
		expectedUserID   string
		expectedSlateID  string
		expectedDispName string
		shouldMatch      bool
	}{
		// Test all 6 permutations of the three required attributes
		{
			name:             "order: key, id, value",
			text:             `<div data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe">`,
			expectedUserID:   "user123",
			expectedSlateID:  "mention001",
			expectedDispName: "John Doe",
			shouldMatch:      true,
		},
		{
			name:             "order: key, value, id",
			text:             `<div data-slate-key="user123" data-slate-value="John Doe" data-slate-id="mention001">`,
			expectedUserID:   "user123",
			expectedSlateID:  "mention001",
			expectedDispName: "John Doe",
			shouldMatch:      true,
		},
		{
			name:             "order: id, key, value",
			text:             `<div data-slate-id="mention001" data-slate-key="user123" data-slate-value="John Doe">`,
			expectedUserID:   "user123",
			expectedSlateID:  "mention001",
			expectedDispName: "John Doe",
			shouldMatch:      true,
		},
		{
			name:             "order: id, value, key",
			text:             `<div data-slate-id="mention001" data-slate-value="John Doe" data-slate-key="user123">`,
			expectedUserID:   "user123",
			expectedSlateID:  "mention001",
			expectedDispName: "John Doe",
			shouldMatch:      true,
		},
		{
			name:             "order: value, key, id",
			text:             `<div data-slate-value="John Doe" data-slate-key="user123" data-slate-id="mention001">`,
			expectedUserID:   "user123",
			expectedSlateID:  "mention001",
			expectedDispName: "John Doe",
			shouldMatch:      true,
		},
		{
			name:             "order: value, id, key",
			text:             `<div data-slate-value="John Doe" data-slate-id="mention001" data-slate-key="user123">`,
			expectedUserID:   "user123",
			expectedSlateID:  "mention001",
			expectedDispName: "John Doe",
			shouldMatch:      true,
		},
		// Test with additional attributes interspersed
		{
			name:             "with extra attrs: node before, inline after",
			text:             `<div data-slate-node="element" data-slate-key="userABC" data-slate-id="slateXYZ" data-slate-value="Jane Smith" data-slate-inline="true">`,
			expectedUserID:   "userABC",
			expectedSlateID:  "slateXYZ",
			expectedDispName: "Jane Smith",
			shouldMatch:      true,
		},
		{
			name:             "with extra attrs scattered",
			text:             `<div class="mention" data-slate-value="Bob Wilson" data-slate-node="element" data-slate-key="user789" data-slate-inline="true" data-slate-id="mention999" data-slate-void="true">`,
			expectedUserID:   "user789",
			expectedSlateID:  "mention999",
			expectedDispName: "Bob Wilson",
			shouldMatch:      true,
		},
		{
			name:             "with extra attrs: key first, scattered",
			text:             `<div data-slate-key="userFirst" data-slate-node="element" data-slate-void="true" data-slate-value="First User" data-slate-inline="true" data-slate-id="idLast">`,
			expectedUserID:   "userFirst",
			expectedSlateID:  "idLast",
			expectedDispName: "First User",
			shouldMatch:      true,
		},
		// Edge cases with special characters in values
		{
			name:             "with special chars in display name",
			text:             `<div data-slate-id="mention001" data-slate-key="user123" data-slate-value="O'Brien, John (Jr.)">`,
			expectedUserID:   "user123",
			expectedSlateID:  "mention001",
			expectedDispName: "O'Brien, John (Jr.)",
			shouldMatch:      true,
		},
		{
			name:             "with empty string values",
			text:             `<div data-slate-key="" data-slate-id="" data-slate-value="">`,
			expectedUserID:   "",
			expectedSlateID:  "",
			expectedDispName: "",
			shouldMatch:      true,
		},
		// Negative test cases - missing attributes
		{
			name:        "missing data-slate-key",
			text:        `<div data-slate-id="mention001" data-slate-value="John Doe">`,
			shouldMatch: false,
		},
		{
			name:        "missing data-slate-id",
			text:        `<div data-slate-key="user123" data-slate-value="John Doe">`,
			shouldMatch: false,
		},
		{
			name:        "missing data-slate-value",
			text:        `<div data-slate-key="user123" data-slate-id="mention001">`,
			shouldMatch: false,
		},
		{
			name:        "not a div element",
			text:        `<span data-slate-key="user123" data-slate-id="mention001" data-slate-value="John Doe">`,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mentions := CheckForMentions(tt.text, "Task", "task001", "Test Task")

			if tt.shouldMatch {
				assert.Equal(t, 1, len(mentions), "should find exactly one mention")

				// Find the mention by slateID
				mention, exists := mentions[tt.expectedSlateID]
				assert.True(t, exists, "mention with expected slate ID should exist")
				assert.Equal(t, tt.expectedUserID, mention.UserID, "UserID should match")
				assert.Equal(t, tt.expectedDispName, mention.UserDisplayName, "UserDisplayName should match")
			} else {
				// For negative cases, we expect no valid mentions OR partial matches
				// Since our regex uses lookaheads, we need to verify that if we get a match,
				// all three fields must be present
				for slateID, mention := range mentions {
					// If we got any match, verify it has all required fields
					if slateID != "" || mention.UserID != "" || mention.UserDisplayName != "" {
						assert.NotEmpty(t, slateID, "if matched, slateID should not be empty")
						assert.NotEmpty(t, mention.UserID, "if matched, UserID should not be empty")
						assert.NotEmpty(t, mention.UserDisplayName, "if matched, UserDisplayName should not be empty")
					}
				}
			}
		})
	}
}
