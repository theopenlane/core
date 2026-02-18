package graphapi

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/workflows"
)

func TestValidateWorkflowDefinitionInput(t *testing.T) {
	tests := []struct {
		name       string
		schemaType string
		doc        *models.WorkflowDefinitionDocument
		wantErr    error
	}{
		{
			name:       "nil document",
			schemaType: "Control",
			doc:        nil,
			wantErr:    ErrDefinitionRequired,
		},
		{
			name:       "no triggers",
			schemaType: "Control",
			doc: &models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{},
				Actions: []models.WorkflowAction{
					{Key: "action1", Type: string(enums.WorkflowActionTypeNotification)},
				},
			},
			wantErr: ErrNoTriggers,
		},
		{
			name:       "no actions",
			schemaType: "Control",
			doc: &models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl},
				},
				Actions: []models.WorkflowAction{},
			},
			wantErr: ErrNoActions,
		},
		{
			name:       "valid minimal workflow",
			schemaType: "Control",
			doc: &models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl},
				},
				Actions: []models.WorkflowAction{
					{Key: "notify", Type: string(enums.WorkflowActionTypeNotification)},
				},
			},
			wantErr: nil,
		},
		{
			name:       "schema type mismatch",
			schemaType: "Control",
			doc: &models.WorkflowDefinitionDocument{
				SchemaType: "InternalPolicy",
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl},
				},
				Actions: []models.WorkflowAction{
					{Key: "notify", Type: string(enums.WorkflowActionTypeNotification)},
				},
			},
			wantErr: ErrSchemaTypeMismatch,
		},
		{
			name:       "review action missing params",
			schemaType: "Control",
			doc: &models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl},
				},
				Actions: []models.WorkflowAction{
					{Key: "review", Type: string(enums.WorkflowActionTypeReview)},
				},
			},
			wantErr: ErrActionInvalidParams,
		},
		{
			name:       "invalid approval timing",
			schemaType: "Control",
			doc: &models.WorkflowDefinitionDocument{
				ApprovalTiming: enums.WorkflowApprovalTiming("INVALID"),
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl},
				},
				Actions: []models.WorkflowAction{
					{Key: "notify", Type: string(enums.WorkflowActionTypeNotification)},
				},
			},
			wantErr: ErrApprovalTimingInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWorkflowDefinitionInput(tt.schemaType, tt.doc, nil)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateReviewActionParams(t *testing.T) {
	params := workflows.ReviewActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: "user-1"},
			},
		},
		Label: "Review Required",
	}
	raw, err := json.Marshal(params)
	require.NoError(t, err)

	err = validateReviewActionParams(raw)
	assert.NoError(t, err)
}

func TestValidateTrigger(t *testing.T) {
	tests := []struct {
		name       string
		schemaType string
		trigger    models.WorkflowTrigger
		wantErr    error
	}{
		{
			name:       "missing operation",
			schemaType: "Control",
			trigger:    models.WorkflowTrigger{ObjectType: enums.WorkflowObjectTypeControl},
			wantErr:    ErrTriggerMissingOperation,
		},
		{
			name:       "unsupported operation",
			schemaType: "Control",
			trigger:    models.WorkflowTrigger{Operation: "INVALID", ObjectType: enums.WorkflowObjectTypeControl},
			wantErr:    ErrTriggerUnsupportedOperation,
		},
		{
			name:       "missing object type",
			schemaType: "Control",
			trigger:    models.WorkflowTrigger{Operation: "UPDATE"},
			wantErr:    ErrTriggerMissingObjectType,
		},
		{
			name:       "object type mismatch",
			schemaType: "Control",
			trigger:    models.WorkflowTrigger{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeInternalPolicy},
			wantErr:    ErrTriggerObjectTypeMismatch,
		},
		{
			name:       "valid trigger",
			schemaType: "Control",
			trigger:    models.WorkflowTrigger{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl},
			wantErr:    nil,
		},
		{
			name:       "valid create operation",
			schemaType: "Control",
			trigger:    models.WorkflowTrigger{Operation: "create", ObjectType: enums.WorkflowObjectTypeControl},
			wantErr:    nil,
		},
		{
			name:       "empty field name",
			schemaType: "Control",
			trigger:    models.WorkflowTrigger{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status", ""}},
			wantErr:    ErrTriggerEmptyFieldName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTrigger(tt.schemaType, tt.trigger, 0, nil)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateApprovalActionParams(t *testing.T) {
	tests := []struct {
		name           string
		params         json.RawMessage
		eligibleFields map[string]struct{}
		wantErr        error
	}{
		{
			name:    "empty params",
			params:  nil,
			wantErr: ErrApprovalParamsRequired,
		},
		{
			name:    "no fields or edges",
			params:  json.RawMessage(`{"targets": [{"type": "user", "id": "user123"}]}`),
			wantErr: ErrApprovalFieldRequired,
		},
		{
			name:    "edges not supported",
			params:  json.RawMessage(`{"fields": [], "edges": ["owner"], "targets": [{"type": "user", "id": "user123"}]}`),
			wantErr: ErrApprovalEdgesNotSupported,
		},
		{
			name:    "no targets",
			params:  json.RawMessage(`{"fields": ["status"]}`),
			wantErr: ErrApprovalTargetsRequired,
		},
		{
			name:   "valid approval params with targets",
			params: json.RawMessage(`{"fields": ["status"], "targets": [{"type": "USER", "id": "user123"}]}`),
			eligibleFields: map[string]struct{}{
				"status": {},
			},
			wantErr: nil,
		},
		{
			name:   "valid approval params with legacy assignees",
			params: json.RawMessage(`{"fields": ["status"], "assignees": {"users": ["user123"]}}`),
			eligibleFields: map[string]struct{}{
				"status": {},
			},
			wantErr: nil,
		},
		{
			name:   "field not eligible",
			params: json.RawMessage(`{"fields": ["noteligible"], "targets": [{"type": "USER", "id": "user123"}]}`),
			eligibleFields: map[string]struct{}{
				"status": {},
			},
			wantErr: ErrApprovalFieldNotEligible,
		},
		{
			name:    "negative required_count",
			params:  json.RawMessage(`{"fields": ["status"], "targets": [{"type": "USER", "id": "user123"}], "required_count": -1}`),
			wantErr: ErrRequiredCountNegative,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateApprovalActionParams(tt.params, tt.eligibleFields)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWebhookActionParams(t *testing.T) {
	tests := []struct {
		name            string
		params          json.RawMessage
		wantErr         error
		wantErrContains string
	}{
		{
			name:    "empty params",
			params:  nil,
			wantErr: ErrWebhookParamsRequired,
		},
		{
			name:    "missing url",
			params:  json.RawMessage(`{"method": "POST"}`),
			wantErr: ErrWebhookURLRequired,
		},
		{
			name:    "invalid url",
			params:  json.RawMessage(`{"url": "not-a-url"}`),
			wantErr: ErrWebhookURLInvalid,
		},
		{
			name:    "valid webhook params",
			params:  json.RawMessage(`{"url": "https://example.com/webhook", "method": "POST"}`),
			wantErr: nil,
		},
		{
			name:    "legacy payload rejected",
			params:  json.RawMessage(`{"url": "https://example.com/webhook", "payload": {"text": "nope"}}`),
			wantErr: ErrWebhookPayloadUnsupported,
		},
		{
			name:            "invalid payload expression",
			params:          json.RawMessage(`{"url": "https://example.com/webhook", "payload_expr": "1 +"}`),
			wantErrContains: "CEL compilation failed",
		},
		{
			name:    "valid payload expression",
			params:  json.RawMessage(`{"url": "https://example.com/webhook", "payload_expr": "{\"value\": \"ok\"}"}`),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookActionParams(tt.params, nil)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
			} else if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFieldUpdateActionParams(t *testing.T) {
	tests := []struct {
		name    string
		params  json.RawMessage
		wantErr error
	}{
		{
			name:    "empty params",
			params:  nil,
			wantErr: ErrFieldUpdateParamsRequired,
		},
		{
			name:    "empty updates",
			params:  json.RawMessage(`{"updates": {}}`),
			wantErr: ErrFieldUpdateUpdatesRequired,
		},
		{
			name:    "valid field update params",
			params:  json.RawMessage(`{"updates": {"status": "active"}}`),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldUpdateActionParams(tt.params)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTarget(t *testing.T) {
	tests := []struct {
		name    string
		target  workflows.TargetConfig
		wantErr error
	}{
		{
			name:    "user target missing id",
			target:  workflows.TargetConfig{Type: enums.WorkflowTargetTypeUser},
			wantErr: ErrTargetMissingID,
		},
		{
			name:    "group target missing id",
			target:  workflows.TargetConfig{Type: enums.WorkflowTargetTypeGroup},
			wantErr: ErrTargetMissingID,
		},
		{
			name:    "role target invalid role",
			target:  workflows.TargetConfig{Type: enums.WorkflowTargetTypeRole, ID: "invalid_role"},
			wantErr: ErrTargetInvalidRole,
		},
		{
			name:    "resolver target missing key",
			target:  workflows.TargetConfig{Type: enums.WorkflowTargetTypeResolver},
			wantErr: ErrTargetMissingResolverKey,
		},
		{
			name:    "valid user target",
			target:  workflows.TargetConfig{Type: enums.WorkflowTargetTypeUser, ID: "user123"},
			wantErr: nil,
		},
		{
			name:    "valid group target",
			target:  workflows.TargetConfig{Type: enums.WorkflowTargetTypeGroup, ID: "group123"},
			wantErr: nil,
		},
		{
			name:    "valid role target",
			target:  workflows.TargetConfig{Type: enums.WorkflowTargetTypeRole, ID: "admin"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTarget(tt.target)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFilterApprovalActions(t *testing.T) {
	actions := []models.WorkflowAction{
		{Key: "approval1", Type: string(enums.WorkflowActionTypeApproval)},
		{Key: "webhook1", Type: string(enums.WorkflowActionTypeWebhook)},
		{Key: "approval2", Type: string(enums.WorkflowActionTypeApproval)},
		{Key: "notification1", Type: string(enums.WorkflowActionTypeNotification)},
	}

	result := filterApprovalActions(actions)

	assert.Len(t, result, 2)
	assert.Equal(t, "approval1", result[0].Key)
	assert.Equal(t, "approval2", result[1].Key)
}

func TestExtractFieldSetKey(t *testing.T) {
	tests := []struct {
		name     string
		action   models.WorkflowAction
		expected string
	}{
		{
			name:     "nil params",
			action:   models.WorkflowAction{Params: nil},
			expected: "",
		},
		{
			name:     "empty fields",
			action:   models.WorkflowAction{Params: json.RawMessage(`{"fields": []}`)},
			expected: "",
		},
		{
			name:     "single field",
			action:   models.WorkflowAction{Params: json.RawMessage(`{"fields": ["status"]}`)},
			expected: "status",
		},
		{
			name:     "multiple fields sorted",
			action:   models.WorkflowAction{Params: json.RawMessage(`{"fields": ["category", "status", "name"]}`)},
			expected: "category,name,status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFieldSetKey(tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateApprovalSubmissionMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    enums.WorkflowApprovalSubmissionMode
		wantErr error
	}{
		{
			name:    "empty mode is valid",
			mode:    "",
			wantErr: nil,
		},
		{
			name:    "auto submit mode is valid",
			mode:    enums.WorkflowApprovalSubmissionModeAutoSubmit,
			wantErr: nil,
		},
		{
			name:    "manual submit mode is not supported",
			mode:    enums.WorkflowApprovalSubmissionModeManualSubmit,
			wantErr: ErrManualSubmitModeNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateApprovalSubmissionMode(tt.mode)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequiredField(t *testing.T) {
	tests := []struct {
		name     string
		required any
		wantErr  error
	}{
		{
			name:     "nil required",
			required: nil,
			wantErr:  nil,
		},
		{
			name:     "bool true",
			required: true,
			wantErr:  nil,
		},
		{
			name:     "bool false",
			required: false,
			wantErr:  nil,
		},
		{
			name:     "positive float",
			required: float64(5),
			wantErr:  nil,
		},
		{
			name:     "negative float",
			required: float64(-1),
			wantErr:  ErrRequiredInvalid,
		},
		{
			name:     "string true",
			required: "true",
			wantErr:  nil,
		},
		{
			name:     "string false",
			required: "false",
			wantErr:  nil,
		},
		{
			name:     "string number",
			required: "5",
			wantErr:  nil,
		},
		{
			name:     "empty string",
			required: "",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredField(tt.required)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
