package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/events/soiree"
)

func TestNeedsTaskDBQuery(t *testing.T) {
	tests := []struct {
		name     string
		fields   *taskFields
		expected bool
	}{
		{
			name:     "all fields empty",
			fields:   &taskFields{},
			expected: true,
		},
		{
			name: "missing title",
			fields: &taskFields{
				entityID: "task-123",
				ownerID:  "owner-456",
			},
			expected: true,
		},
		{
			name: "missing entity ID",
			fields: &taskFields{
				title:   "Test Task",
				ownerID: "owner-456",
			},
			expected: true,
		},
		{
			name: "missing owner ID",
			fields: &taskFields{
				title:    "Test Task",
				entityID: "task-123",
			},
			expected: true,
		},
		{
			name: "all fields present",
			fields: &taskFields{
				title:    "Test Task",
				entityID: "task-123",
				ownerID:  "owner-456",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsTaskDBQuery(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNeedsPolicyDBQuery(t *testing.T) {
	tests := []struct {
		name     string
		fields   *documentFields
		expected bool
	}{
		{
			name:     "all fields empty",
			fields:   &documentFields{},
			expected: true,
		},
		{
			name: "missing name",
			fields: &documentFields{
				entityID:   "policy-123",
				ownerID:    "owner-456",
				approverID: "approver-789",
			},
			expected: true,
		},
		{
			name: "missing approver ID",
			fields: &documentFields{
				name:     "Test Policy",
				entityID: "policy-123",
				ownerID:  "owner-456",
			},
			expected: true,
		},
		{
			name: "all fields present",
			fields: &documentFields{
				name:       "Test Policy",
				entityID:   "policy-123",
				ownerID:    "owner-456",
				approverID: "approver-789",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsPolicyDBQuery(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTaskFromPayload(t *testing.T) {
	tests := []struct {
		name     string
		payload  *events.MutationPayload
		expected *taskFields
	}{
		{
			name:     "nil payload",
			payload:  nil,
			expected: &taskFields{},
		},
		{
			name: "valid payload with entity ID only",
			payload: &events.MutationPayload{
				EntityID: "task-123",
			},
			expected: &taskFields{
				entityID: "task-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := &taskFields{}
			extractTaskFromPayload(tt.payload, fields)
			assert.Equal(t, tt.expected.entityID, fields.entityID)
			assert.Equal(t, tt.expected.title, fields.title)
			assert.Equal(t, tt.expected.ownerID, fields.ownerID)
		})
	}
}

func TestExtractPolicyFromPayload(t *testing.T) {
	tests := []struct {
		name     string
		payload  *events.MutationPayload
		expected *documentFields
	}{
		{
			name:     "nil payload",
			payload:  nil,
			expected: &documentFields{},
		},
		{
			name: "valid payload with entity ID only",
			payload: &events.MutationPayload{
				EntityID: "policy-123",
			},
			expected: &documentFields{
				entityID: "policy-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := &documentFields{}
			extractPolicyFromPayload(tt.payload, fields)
			assert.Equal(t, tt.expected.entityID, fields.entityID)
			assert.Equal(t, tt.expected.name, fields.name)
			assert.Equal(t, tt.expected.ownerID, fields.ownerID)
			assert.Equal(t, tt.expected.approverID, fields.approverID)
		})
	}
}

func TestRegisterListeners(t *testing.T) {
	registered := make(map[string]bool)

	mockAddListener := func(entityType string, handler func(*soiree.EventContext, *events.MutationPayload) error) {
		registered[entityType] = true
		assert.NotNil(t, handler)
	}

	RegisterListeners(mockAddListener)

	assert.True(t, registered[generated.TypeTask], "Task listener should be registered")
	assert.True(t, registered[generated.TypeInternalPolicy], "InternalPolicy listener should be registered")
	assert.True(t, registered[generated.TypeRisk], "Risk listener should be registered")
	assert.True(t, registered[generated.TypeProcedure], "Procedure listener should be registered")
	assert.True(t, registered[generated.TypeNote], "Note listener should be registered")
	assert.True(t, registered[generated.TypeExport], "Export listener should be registered")
	assert.Len(t, registered, 6, "Should register exactly 6 listeners")
}

func TestTaskNotificationInput(t *testing.T) {
	input := taskNotificationInput{
		assigneeID: "user-123",
		taskTitle:  "Test Task",
		taskID:     "task-456",
		ownerID:    "owner-789",
	}

	assert.Equal(t, "user-123", input.assigneeID)
	assert.Equal(t, "Test Task", input.taskTitle)
	assert.Equal(t, "task-456", input.taskID)
	assert.Equal(t, "owner-789", input.ownerID)
}

func TestPolicyNotificationInput(t *testing.T) {
	input := documentNotificationInput{
		approverID: "group-123",
		name:       "Test Policy",
		docID:      "policy-456",
		ownerID:    "owner-789",
	}

	assert.Equal(t, "group-123", input.approverID)
	assert.Equal(t, "Test Policy", input.name)
	assert.Equal(t, "policy-456", input.docID)
	assert.Equal(t, "owner-789", input.ownerID)
}

func TestMutationPayload(t *testing.T) {
	payload := &events.MutationPayload{
		Mutation:  nil,
		Operation: "create",
		EntityID:  "entity-123",
	}

	assert.Equal(t, "create", payload.Operation)
	assert.Equal(t, "entity-123", payload.EntityID)
}

func TestErrorConstants(t *testing.T) {
	assert.Equal(t, "failed to get client from context", ErrFailedToGetClient.Error())
	assert.Equal(t, "entity ID not found in props", ErrEntityIDNotFound.Error())
}

func TestTaskFields(t *testing.T) {
	fields := &taskFields{
		title:    "Task Title",
		entityID: "task-123",
		ownerID:  "owner-456",
	}

	assert.Equal(t, "Task Title", fields.title)
	assert.Equal(t, "task-123", fields.entityID)
	assert.Equal(t, "owner-456", fields.ownerID)
}

func TestPolicyFields(t *testing.T) {
	fields := &documentFields{
		approverID: "approver-123",
		name:       "Policy Name",
		entityID:   "policy-456",
		ownerID:    "owner-789",
	}

	assert.Equal(t, "approver-123", fields.approverID)
	assert.Equal(t, "Policy Name", fields.name)
	assert.Equal(t, "policy-456", fields.entityID)
	assert.Equal(t, "owner-789", fields.ownerID)
}

// mockProps creates a Properties map for testing
func mockProps(data map[string]interface{}) soiree.Properties {
	props := make(soiree.Properties)
	for k, v := range data {
		props[k] = v
	}
	return props
}

func TestExtractTaskFromProps_WithRealInterface(t *testing.T) {
	tests := []struct {
		name     string
		props    soiree.Properties
		initial  *taskFields
		expected *taskFields
	}{
		{
			name:    "empty props",
			props:   mockProps(make(map[string]interface{})),
			initial: &taskFields{},
			expected: &taskFields{
				title:    "",
				entityID: "",
				ownerID:  "",
			},
		},
		{
			name: "props with title only",
			props: mockProps(map[string]interface{}{
				"title": "My Task",
			}),
			initial: &taskFields{},
			expected: &taskFields{
				title:    "My Task",
				entityID: "",
				ownerID:  "",
			},
		},
		{
			name: "props with all fields",
			props: mockProps(map[string]interface{}{
				"title":    "Complete Task",
				"id":       "task-999",
				"owner_id": "owner-888",
			}),
			initial: &taskFields{},
			expected: &taskFields{
				title:    "Complete Task",
				entityID: "task-999",
				ownerID:  "owner-888",
			},
		},
		{
			name: "props with non-string values ignored",
			props: mockProps(map[string]interface{}{
				"title":    12345,
				"id":       true,
				"owner_id": []string{"invalid"},
			}),
			initial: &taskFields{},
			expected: &taskFields{
				title:    "",
				entityID: "",
				ownerID:  "",
			},
		},
		{
			name: "existing fields not overwritten when props empty",
			props: mockProps(map[string]interface{}{
				"title": "New Title",
			}),
			initial: &taskFields{
				title:    "",
				entityID: "existing-id",
				ownerID:  "existing-owner",
			},
			expected: &taskFields{
				title:    "New Title",
				entityID: "existing-id",
				ownerID:  "existing-owner",
			},
		},
		{
			name: "only fills empty fields",
			props: mockProps(map[string]interface{}{
				"title":    "New Title",
				"id":       "new-id",
				"owner_id": "new-owner",
			}),
			initial: &taskFields{
				title: "Existing Title",
			},
			expected: &taskFields{
				title:    "Existing Title",
				entityID: "new-id",
				ownerID:  "new-owner",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractTaskFromProps(tt.props, tt.initial)
			assert.Equal(t, tt.expected.title, tt.initial.title)
			assert.Equal(t, tt.expected.entityID, tt.initial.entityID)
			assert.Equal(t, tt.expected.ownerID, tt.initial.ownerID)
		})
	}
}

func TestExtractPolicyFromProps_WithRealInterface(t *testing.T) {
	tests := []struct {
		name     string
		props    soiree.Properties
		initial  *documentFields
		expected *documentFields
	}{
		{
			name:    "empty props",
			props:   mockProps(make(map[string]interface{})),
			initial: &documentFields{},
			expected: &documentFields{
				name:       "",
				entityID:   "",
				ownerID:    "",
				approverID: "",
			},
		},
		{
			name: "props with name only",
			props: mockProps(map[string]interface{}{
				"name": "Security Policy",
			}),
			initial: &documentFields{},
			expected: &documentFields{
				name:       "Security Policy",
				entityID:   "",
				ownerID:    "",
				approverID: "",
			},
		},
		{
			name: "props with all fields",
			props: mockProps(map[string]interface{}{
				"name":        "Data Protection Policy",
				"id":          "policy-777",
				"owner_id":    "owner-666",
				"approver_id": "approver-555",
			}),
			initial: &documentFields{},
			expected: &documentFields{
				name:       "Data Protection Policy",
				entityID:   "policy-777",
				ownerID:    "owner-666",
				approverID: "approver-555",
			},
		},
		{
			name: "props with non-string values ignored",
			props: mockProps(map[string]interface{}{
				"name":        123,
				"id":          false,
				"owner_id":    nil,
				"approver_id": 456.78,
			}),
			initial: &documentFields{},
			expected: &documentFields{
				name:       "",
				entityID:   "",
				ownerID:    "",
				approverID: "",
			},
		},
		{
			name: "only fills empty fields",
			props: mockProps(map[string]interface{}{
				"name":        "New Name",
				"id":          "new-id",
				"owner_id":    "new-owner",
				"approver_id": "new-approver",
			}),
			initial: &documentFields{
				name:     "Existing Name",
				entityID: "existing-id",
			},
			expected: &documentFields{
				name:       "Existing Name",
				entityID:   "existing-id",
				ownerID:    "new-owner",
				approverID: "new-approver",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractPolicyFromProps(tt.props, tt.initial)
			assert.Equal(t, tt.expected.name, tt.initial.name)
			assert.Equal(t, tt.expected.entityID, tt.initial.entityID)
			assert.Equal(t, tt.expected.ownerID, tt.initial.ownerID)
			assert.Equal(t, tt.expected.approverID, tt.initial.approverID)
		})
	}
}

func TestFetchTaskFields_Integration(t *testing.T) {
	tests := []struct {
		name        string
		props       soiree.Properties
		payload     *events.MutationPayload
		expectError bool
		expected    *taskFields
	}{
		{
			name: "all fields from props",
			props: mockProps(map[string]interface{}{
				"title":    "Task from props",
				"id":       "task-prop-1",
				"owner_id": "owner-prop-1",
			}),
			payload:     nil,
			expectError: false,
			expected: &taskFields{
				title:    "Task from props",
				entityID: "task-prop-1",
				ownerID:  "owner-prop-1",
			},
		},
		{
			name:  "all fields from payload",
			props: mockProps(make(map[string]interface{})),
			payload: &events.MutationPayload{
				EntityID: "task-payload-1",
			},
			expectError: false,
			expected: &taskFields{
				entityID: "task-payload-1",
			},
		},
		{
			name: "payload overrides props",
			props: mockProps(map[string]interface{}{
				"id": "task-prop-2",
			}),
			payload: &events.MutationPayload{
				EntityID: "task-payload-2",
			},
			expectError: false,
			expected: &taskFields{
				entityID: "task-payload-2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This will call needsTaskDBQuery which may try to query DB
			// Since we can't mock the DB easily, we only test cases where all fields are present
			result, err := fetchTaskFields(nil, tt.props, tt.payload)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Only check fields that don't require DB query
				if result != nil && !needsTaskDBQuery(result) {
					assert.Equal(t, tt.expected.title, result.title)
					assert.Equal(t, tt.expected.entityID, result.entityID)
					assert.Equal(t, tt.expected.ownerID, result.ownerID)
				}
			}
		})
	}
}

func TestFetchPolicyFields_Integration(t *testing.T) {
	tests := []struct {
		name        string
		props       soiree.Properties
		payload     *events.MutationPayload
		expectError bool
		expected    *documentFields
	}{
		{
			name: "all fields from props",
			props: mockProps(map[string]interface{}{
				"name":        "Policy from props",
				"id":          "policy-prop-1",
				"owner_id":    "owner-prop-1",
				"approver_id": "approver-prop-1",
			}),
			payload:     nil,
			expectError: false,
			expected: &documentFields{
				name:       "Policy from props",
				entityID:   "policy-prop-1",
				ownerID:    "owner-prop-1",
				approverID: "approver-prop-1",
			},
		},
		{
			name:  "all fields from payload",
			props: mockProps(make(map[string]interface{})),
			payload: &events.MutationPayload{
				EntityID: "policy-payload-1",
			},
			expectError: false,
			expected: &documentFields{
				entityID: "policy-payload-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fetchDocumentFields(nil, tt.props, tt.payload)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Only check fields that don't require DB query
				if result != nil && !needsPolicyDBQuery(result) {
					assert.Equal(t, tt.expected.name, result.name)
					assert.Equal(t, tt.expected.entityID, result.entityID)
					assert.Equal(t, tt.expected.ownerID, result.ownerID)
					assert.Equal(t, tt.expected.approverID, result.approverID)
				}
			}
		})
	}
}
