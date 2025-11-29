package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
		fields   *policyFields
		expected bool
	}{
		{
			name:     "all fields empty",
			fields:   &policyFields{},
			expected: true,
		},
		{
			name: "missing name",
			fields: &policyFields{
				entityID:   "policy-123",
				ownerID:    "owner-456",
				approverID: "approver-789",
			},
			expected: true,
		},
		{
			name: "missing approver ID",
			fields: &policyFields{
				name:     "Test Policy",
				entityID: "policy-123",
				ownerID:  "owner-456",
			},
			expected: true,
		},
		{
			name: "all fields present",
			fields: &policyFields{
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
		payload  interface{}
		expected *taskFields
	}{
		{
			name:     "nil payload",
			payload:  nil,
			expected: &taskFields{},
		},
		{
			name:     "non-mutation payload",
			payload:  "invalid",
			expected: &taskFields{},
		},
		{
			name: "valid payload with entity ID only",
			payload: &mutationPayload{
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
		payload  interface{}
		expected *policyFields
	}{
		{
			name:     "nil payload",
			payload:  nil,
			expected: &policyFields{},
		},
		{
			name:     "non-mutation payload",
			payload:  "invalid",
			expected: &policyFields{},
		},
		{
			name: "valid payload with entity ID only",
			payload: &mutationPayload{
				EntityID: "policy-123",
			},
			expected: &policyFields{
				entityID: "policy-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := &policyFields{}
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

	mockAddListener := func(entityType string, handler func(*soiree.EventContext, any) error) {
		registered[entityType] = true
		assert.NotNil(t, handler)
	}

	RegisterListeners(mockAddListener)

	assert.True(t, registered[generated.TypeTask], "Task listener should be registered")
	assert.True(t, registered[generated.TypeInternalPolicy], "InternalPolicy listener should be registered")
	assert.Len(t, registered, 2, "Should register exactly 2 listeners")
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
	input := policyNotificationInput{
		approverID: "group-123",
		policyName: "Test Policy",
		policyID:   "policy-456",
		ownerID:    "owner-789",
	}

	assert.Equal(t, "group-123", input.approverID)
	assert.Equal(t, "Test Policy", input.policyName)
	assert.Equal(t, "policy-456", input.policyID)
	assert.Equal(t, "owner-789", input.ownerID)
}

func TestMutationPayload(t *testing.T) {
	payload := &mutationPayload{
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
	fields := &policyFields{
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
