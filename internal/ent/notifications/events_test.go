package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/export"
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/gala"
)

func TestNeedsTaskDBQuery(t *testing.T) {
	tests := []struct {
		name     string
		fields   *taskFields
		expected bool
	}{
		{name: "all fields empty", fields: &taskFields{}, expected: true},
		{
			name: "all fields present",
			fields: &taskFields{
				title:    "Task",
				entityID: "task-1",
				ownerID:  "owner-1",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, needsTaskDBQuery(tt.fields))
		})
	}
}

func TestNeedsDocumentDBQuery(t *testing.T) {
	tests := []struct {
		name     string
		fields   *documentFields
		expected bool
	}{
		{name: "all fields empty", fields: &documentFields{}, expected: true},
		{
			name: "all fields present",
			fields: &documentFields{
				name:       "Policy",
				entityID:   "policy-1",
				ownerID:    "owner-1",
				approverID: "approver-1",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, needsDocumentDBQuery(tt.fields))
		})
	}
}

func TestExtractTaskFromPayloadAndProps(t *testing.T) {
	fields := &taskFields{}
	payload := eventqueue.MutationGalaPayload{
		EntityID: "task-1",
		ProposedChanges: map[string]any{
			task.FieldTitle:   "Task One",
			task.FieldOwnerID: "owner-1",
		},
	}

	extractTaskFromPayload(payload, fields)
	assert.Equal(t, "task-1", fields.entityID)
	assert.Equal(t, "Task One", fields.title)
	assert.Equal(t, "owner-1", fields.ownerID)

	props := map[string]string{
		task.FieldTitle:   "Task From Props",
		task.FieldID:      "task-prop",
		task.FieldOwnerID: "owner-prop",
	}
	extractTaskFromProps(props, fields)
	assert.Equal(t, "task-1", fields.entityID)
	assert.Equal(t, "Task One", fields.title)
	assert.Equal(t, "owner-1", fields.ownerID)
}

func TestExtractDocumentFromPayloadAndProps(t *testing.T) {
	fields := &documentFields{}
	payload := eventqueue.MutationGalaPayload{
		EntityID: "policy-1",
		ProposedChanges: map[string]any{
			internalpolicy.FieldName:       "Policy One",
			internalpolicy.FieldOwnerID:    "owner-1",
			internalpolicy.FieldApproverID: "approver-1",
		},
	}

	extractDocumentFromPayload(payload, fields)
	assert.Equal(t, "policy-1", fields.entityID)
	assert.Equal(t, "Policy One", fields.name)
	assert.Equal(t, "owner-1", fields.ownerID)
	assert.Equal(t, "approver-1", fields.approverID)

	props := map[string]string{
		internalpolicy.FieldName:       "Policy From Props",
		internalpolicy.FieldID:         "policy-prop",
		internalpolicy.FieldOwnerID:    "owner-prop",
		internalpolicy.FieldApproverID: "approver-prop",
	}
	extractDocumentFromProps(props, fields)
	assert.Equal(t, "policy-1", fields.entityID)
	assert.Equal(t, "Policy One", fields.name)
	assert.Equal(t, "owner-1", fields.ownerID)
	assert.Equal(t, "approver-1", fields.approverID)
}

func TestExtractExportFromPayloadAndProps(t *testing.T) {
	fields := &exportFields{}
	payload := eventqueue.MutationGalaPayload{
		EntityID: "export-1",
		ProposedChanges: map[string]any{
			export.FieldOwnerID:      "owner-1",
			export.FieldRequestorID:  "requestor-1",
			export.FieldExportType:   enums.ExportTypeTask.String(),
			export.FieldStatus:       enums.ExportStatusReady.String(),
			export.FieldErrorMessage: "none",
		},
	}

	extractExportFromPayload(payload, fields)
	assert.Equal(t, "export-1", fields.entityID)
	assert.Equal(t, "owner-1", fields.ownerID)
	assert.Equal(t, "requestor-1", fields.requestorID)
	assert.Equal(t, enums.ExportTypeTask, fields.exportType)
	assert.Equal(t, enums.ExportStatusReady, fields.status)
	assert.Equal(t, "none", fields.errorMessage)

	props := map[string]string{
		export.FieldOwnerID:      "owner-prop",
		export.FieldRequestorID:  "requestor-prop",
		export.FieldExportType:   enums.ExportTypeRisk.String(),
		export.FieldStatus:       enums.ExportStatusFailed.String(),
		export.FieldErrorMessage: "error",
		export.FieldID:           "export-prop",
	}
	extractExportFromProps(props, fields)
	assert.Equal(t, "export-1", fields.entityID)
	assert.Equal(t, "owner-1", fields.ownerID)
	assert.Equal(t, "requestor-1", fields.requestorID)
	assert.Equal(t, enums.ExportTypeTask, fields.exportType)
	assert.Equal(t, enums.ExportStatusReady, fields.status)
	assert.Equal(t, "none", fields.errorMessage)
}

func TestRegisterGalaListeners(t *testing.T) {
	registry := gala.NewRegistry()

	ids, err := RegisterGalaListeners(registry)
	require.NoError(t, err)
	require.Len(t, ids, 6)

	assert.True(t, registry.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeTask), "create"))
	assert.True(t, registry.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeInternalPolicy), "update"))
	assert.True(t, registry.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeRisk), "delete"))
	assert.True(t, registry.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeProcedure), "update_one"))
	assert.True(t, registry.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeNote), "create"))
	assert.True(t, registry.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, generated.TypeExport), "update"))
}

func TestErrorConstants(t *testing.T) {
	assert.Equal(t, "failed to get client from context", ErrFailedToGetClient.Error())
	assert.Equal(t, "entity ID not found in payload metadata", ErrEntityIDNotFound.Error())
}
