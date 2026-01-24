package workflows

import (
	"encoding/json"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
)

// TestDeriveDomainKey verifies domain key generation
func TestDeriveDomainKey(t *testing.T) {
	tests := []struct {
		name     string
		fields   []string
		expected string
	}{
		{
			name:     "empty fields",
			fields:   []string{},
			expected: "",
		},
		{
			name:     "single field",
			fields:   []string{"status"},
			expected: "status",
		},
		{
			name:     "multiple fields sorted",
			fields:   []string{"category", "status"},
			expected: "category,status",
		},
		{
			name:     "multiple fields unsorted",
			fields:   []string{"status", "category"},
			expected: "category,status",
		},
		{
			name:     "three fields",
			fields:   []string{"text", "description", "name"},
			expected: "description,name,text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeriveDomainKey(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDeriveDomainKey_Stability verifies stable domain key ordering
func TestDeriveDomainKey_Stability(t *testing.T) {
	fields1 := []string{"text", "description"}
	fields2 := []string{"description", "text"}

	key1 := DeriveDomainKey(fields1)
	key2 := DeriveDomainKey(fields2)

	assert.Equal(t, key1, key2, "domain keys should be stable regardless of input order")
}

// TestFilterChangesForDomain verifies filtering changes by domain
func TestFilterChangesForDomain(t *testing.T) {
	tests := []struct {
		name           string
		changes        map[string]any
		approvalFields []string
		expected       map[string]any
	}{
		{
			name:           "empty changes",
			changes:        map[string]any{},
			approvalFields: []string{"text", "description"},
			expected:       nil,
		},
		{
			name:           "empty approval fields",
			changes:        map[string]any{"text": "foo"},
			approvalFields: []string{},
			expected:       map[string]any{},
		},
		{
			name: "all fields in domain changed",
			changes: map[string]any{
				"text":        "new text",
				"description": "new desc",
			},
			approvalFields: []string{"text", "description"},
			expected: map[string]any{
				"text":        "new text",
				"description": "new desc",
			},
		},
		{
			name: "partial fields in domain changed",
			changes: map[string]any{
				"text": "new text",
			},
			approvalFields: []string{"text", "description"},
			expected: map[string]any{
				"text": "new text",
			},
		},
		{
			name: "mutation includes fields outside domain",
			changes: map[string]any{
				"text":        "new text",
				"other_field": "value",
			},
			approvalFields: []string{"text", "description"},
			expected: map[string]any{
				"text": "new text",
			},
		},
		{
			name: "no fields in domain changed",
			changes: map[string]any{
				"other_field": "value",
			},
			approvalFields: []string{"text", "description"},
			expected:       map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterChangesForDomain(tt.changes, tt.approvalFields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestComputeProposalHash verifies proposal hash generation
func TestComputeProposalHash(t *testing.T) {
	tests := []struct {
		name        string
		changes     map[string]any
		expectEmpty bool
		expectError bool
	}{
		{
			name:        "empty changes",
			changes:     map[string]any{},
			expectEmpty: false,
		},
		{
			name: "single field",
			changes: map[string]any{
				"text": "value",
			},
			expectEmpty: false,
		},
		{
			name: "multiple fields",
			changes: map[string]any{
				"text":        "value",
				"description": "desc",
			},
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := ComputeProposalHash(tt.changes)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expectEmpty {
				assert.Empty(t, hash)
			} else {
				assert.NotEmpty(t, hash)
				assert.Len(t, hash, 64, "SHA256 hash should be 64 hex characters")
			}
		})
	}
}

// TestComputeProposalHash_Stability verifies hash stability across map order
func TestComputeProposalHash_Stability(t *testing.T) {
	changes1 := map[string]any{
		"text":        "value",
		"description": "desc",
	}

	changes2 := map[string]any{
		"description": "desc",
		"text":        "value",
	}

	hash1, err1 := ComputeProposalHash(changes1)
	require.NoError(t, err1)

	hash2, err2 := ComputeProposalHash(changes2)
	require.NoError(t, err2)

	assert.Equal(t, hash1, hash2, "hashes should be stable regardless of map iteration order")
}

// TestComputeProposalHash_Deterministic verifies deterministic hashes
func TestComputeProposalHash_Deterministic(t *testing.T) {
	changes := map[string]any{
		"text":        "value",
		"description": "desc",
		"status":      "active",
	}

	hash1, err := ComputeProposalHash(changes)
	require.NoError(t, err)

	hash2, err := ComputeProposalHash(changes)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2, "hash should be deterministic")
}

// TestApprovalDomains verifies approval domain extraction
func TestApprovalDomains(t *testing.T) {
	primaryParams, err := json.Marshal(ApprovalActionParams{
		Fields: []string{" status ", "name", ""},
	})
	require.NoError(t, err)

	duplicateParams, err := json.Marshal(ApprovalActionParams{
		Fields: []string{"name", "status"},
	})
	require.NoError(t, err)

	secondaryParams, err := json.Marshal(ApprovalActionParams{
		Fields: []string{"priority"},
	})
	require.NoError(t, err)

	doc := models.WorkflowDefinitionDocument{
		Actions: []models.WorkflowAction{
			{
				Key:    "notify",
				Type:   enums.WorkflowActionTypeNotification.String(),
				Params: primaryParams,
			},
			{
				Key:    "approval_primary",
				Type:   enums.WorkflowActionTypeApproval.String(),
				Params: primaryParams,
			},
			{
				Key:    "approval_duplicate",
				Type:   enums.WorkflowActionTypeApproval.String(),
				Params: duplicateParams,
			},
			{
				Key:    "approval_secondary",
				Type:   enums.WorkflowActionTypeApproval.String(),
				Params: secondaryParams,
			},
		},
	}

	domains, err := ApprovalDomains(doc)
	require.NoError(t, err)

	assert.Equal(t, [][]string{
		{" status ", "name"},
		{"name", "status"},
		{"priority"},
	}, domains)
}

// TestApprovalDomainsInvalidParams verifies invalid params handling
func TestApprovalDomainsInvalidParams(t *testing.T) {
	doc := models.WorkflowDefinitionDocument{
		Actions: []models.WorkflowAction{
			{
				Key:    "approval",
				Type:   enums.WorkflowActionTypeApproval.String(),
				Params: json.RawMessage("{invalid"),
			},
		},
	}

	_, err := ApprovalDomains(doc)
	assert.ErrorIs(t, err, ErrApprovalActionParamsInvalid)
}

// TestSplitChangesByDomains verifies splitting changes into domains
func TestSplitChangesByDomains(t *testing.T) {
	changes := map[string]any{
		"status":      "approved",
		"description": "update",
		"priority":    "p1",
	}

	domains := [][]string{
		{"status"},
		{"description", "priority"},
		{"missing"},
	}

	matches := SplitChangesByDomains(changes, domains)
	require.Len(t, matches, 2)

	assert.Equal(t, DeriveDomainKey([]string{"status"}), matches[0].DomainKey)
	assert.Equal(t, []string{"status"}, matches[0].Fields)
	assert.Equal(t, map[string]any{"status": "approved"}, matches[0].Changes)

	assert.Equal(t, DeriveDomainKey([]string{"description", "priority"}), matches[1].DomainKey)
	assert.Equal(t, []string{"description", "priority"}, matches[1].Fields)
	assert.Equal(t, map[string]any{"description": "update", "priority": "p1"}, matches[1].Changes)
}

// TestDomainChangesForDefinition verifies domain change extraction
func TestDomainChangesForDefinition(t *testing.T) {
	params, err := json.Marshal(ApprovalActionParams{
		Fields: []string{"status", "priority"},
	})
	require.NoError(t, err)

	doc := models.WorkflowDefinitionDocument{
		Actions: []models.WorkflowAction{
			{
				Key:    "approval",
				Type:   enums.WorkflowActionTypeApproval.String(),
				Params: params,
			},
		},
	}

	changes := map[string]any{
		"status":      "approved",
		"description": "update",
	}

	domainChanges, err := DomainChangesForDefinition(doc, changes)
	require.NoError(t, err)
	require.Len(t, domainChanges, 1)

	expectedFields := []string{"priority", "status"}
	assert.Equal(t, DeriveDomainKey(expectedFields), domainChanges[0].DomainKey)
	assert.Equal(t, expectedFields, domainChanges[0].Fields)
	assert.Equal(t, map[string]any{"status": "approved"}, domainChanges[0].Changes)
}

// TestDomainChangesForDefinitionFallback verifies fallback behavior
func TestDomainChangesForDefinitionFallback(t *testing.T) {
	doc := models.WorkflowDefinitionDocument{
		Actions: []models.WorkflowAction{
			{Key: "notify", Type: enums.WorkflowActionTypeNotification.String()},
		},
	}

	changes := map[string]any{
		"status": "approved",
		"name":   "example",
	}

	domainChanges, err := DomainChangesForDefinition(doc, changes)
	require.NoError(t, err)
	require.Len(t, domainChanges, 1)

	fields := lo.Keys(changes)
	assert.Equal(t, DeriveDomainKey(fields), domainChanges[0].DomainKey)
	assert.ElementsMatch(t, fields, domainChanges[0].Fields)
	assert.Equal(t, changes, domainChanges[0].Changes)
}

// TestDefinitionHasApprovalAction verifies approval detection
func TestDefinitionHasApprovalAction(t *testing.T) {
	doc := models.WorkflowDefinitionDocument{
		Actions: []models.WorkflowAction{
			{Key: "notify", Type: enums.WorkflowActionTypeNotification.String()},
		},
	}

	assert.False(t, DefinitionHasApprovalAction(doc))

	doc.Actions = append(doc.Actions, models.WorkflowAction{
		Key:  "approval",
		Type: enums.WorkflowActionTypeApproval.String(),
	})

	assert.True(t, DefinitionHasApprovalAction(doc))
}

// TestApprovalDomainsFromDefinition verifies approval domains from definition
func TestApprovalDomainsFromDefinition(t *testing.T) {
	def := &generated.WorkflowDefinition{
		DefinitionJSON: models.WorkflowDefinitionDocument{
			Actions: []models.WorkflowAction{
				{Key: "notify", Type: enums.WorkflowActionTypeNotification.String()},
			},
		},
	}

	domains, err := ApprovalDomains(def.DefinitionJSON)
	require.NoError(t, err)
	assert.Empty(t, domains)

	params, err := json.Marshal(ApprovalActionParams{
		Fields: []string{"name"},
	})
	require.NoError(t, err)

	def.DefinitionJSON.Actions = append(def.DefinitionJSON.Actions, models.WorkflowAction{
		Key:    "approval",
		Type:   enums.WorkflowActionTypeApproval.String(),
		Params: params,
	})

	domains, err = ApprovalDomains(def.DefinitionJSON)
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"name"}}, domains)
}

// TestFieldsFromChanges verifies field list extraction
func TestFieldsFromChanges(t *testing.T) {
	changes := map[string]any{
		"b": 1,
		"a": 2,
		"c": 3,
	}

	assert.Equal(t, []string{"a", "b", "c"}, FieldsFromChanges(changes))
}
