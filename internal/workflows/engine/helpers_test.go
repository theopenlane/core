package engine

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/events/soiree"
)

type testEntity struct {
	// Tags holds tag values
	Tags []string `json:"tags"`
	// Aliases holds alias values
	Aliases []string `json:"aliases"`
	// Count is a numeric field
	Count int `json:"count"`
}

// TestStringSliceField verifies string slice extraction
func TestStringSliceField(t *testing.T) {
	t.Run("returns slice for json key", func(t *testing.T) {
		got, err := workflows.StringSliceField(&testEntity{Tags: []string{"a", "b"}}, "tags")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !slices.Equal(got, []string{"a", "b"}) {
			t.Fatalf("expected tags slice, got %#v", got)
		}
	})

	t.Run("returns nil for non-slice field", func(t *testing.T) {
		got, err := workflows.StringSliceField(&testEntity{Count: 1}, "count")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil for non-slice field, got %#v", got)
		}
	})

	t.Run("returns error for nil entity", func(t *testing.T) {
		_, err := workflows.StringSliceField(nil, "tags")
		if !errors.Is(err, workflows.ErrStringFieldNil) {
			t.Fatalf("expected ErrStringFieldNil, got %#v", err)
		}
	})

	t.Run("returns nil for empty field", func(t *testing.T) {
		got, err := workflows.StringSliceField(&testEntity{Tags: []string{"a"}}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil for empty field, got %#v", got)
		}
	})
}

// TestResolveApproval verifies approval resolution logic
func TestResolveApproval(t *testing.T) {
	cases := []struct {
		name          string
		requiredCount int
		counts        AssignmentStatusCounts
		want          approvalResolution
	}{
		{
			name:          "satisfied quorum",
			requiredCount: 2,
			counts: AssignmentStatusCounts{
				Approved: 2,
				Pending:  0,
			},
			want: approvalSatisfied,
		},
		{
			name:          "pending approvals",
			requiredCount: 2,
			counts: AssignmentStatusCounts{
				Approved: 1,
				Pending:  1,
			},
			want: approvalPending,
		},
		{
			name:          "failed quorum without pending",
			requiredCount: 2,
			counts: AssignmentStatusCounts{
				Approved: 1,
				Pending:  0,
			},
			want: approvalFailed,
		},
		{
			name:          "pending optional approvals",
			requiredCount: 0,
			counts: AssignmentStatusCounts{
				Pending: 1,
			},
			want: approvalPending,
		},
		{
			name:          "satisfied optional approvals",
			requiredCount: 0,
			counts: AssignmentStatusCounts{
				Pending: 0,
			},
			want: approvalSatisfied,
		},
		{
			name:          "rejected required",
			requiredCount: 1,
			counts: AssignmentStatusCounts{
				RejectedRequired: true,
			},
			want: approvalFailed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveApproval(tc.requiredCount, tc.counts)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// TestNotifyActionKey verifies notification action key selection
func TestNotifyActionKey(t *testing.T) {
	action := models.WorkflowAction{Key: "review"}
	got := notifyActionKey(action, 3)
	if got != "notify_review_3" {
		t.Fatalf("expected notify_review_3, got %s", got)
	}
}

// TestWrapCreationError verifies creation error wrapping
func TestWrapCreationError(t *testing.T) {
	baseErr := errors.New("boom")
	instanceErr := wrapCreationError(&workflows.WorkflowCreationError{
		Stage: workflows.WorkflowCreationStageInstance,
		Err:   baseErr,
	})
	if !strings.Contains(instanceErr.Error(), "failed to create workflow instance") {
		t.Fatalf("expected instance creation error, got %v", instanceErr)
	}

	refErr := wrapCreationError(&workflows.WorkflowCreationError{
		Stage: workflows.WorkflowCreationStageObjectRef,
		Err:   baseErr,
	})
	if !strings.Contains(refErr.Error(), "failed to create object reference") {
		t.Fatalf("expected object reference error, got %v", refErr)
	}

	unknownErr := wrapCreationError(baseErr)
	if !strings.Contains(unknownErr.Error(), "failed to create workflow instance") {
		t.Fatalf("expected generic creation error, got %v", unknownErr)
	}
}

func TestApplyStringTemplates(t *testing.T) {
	replacements := map[string]string{
		"name": "Ada",
		"id":   "42",
	}
	input := map[string]any{
		"title": "Hello {{name}}",
		"nested": map[string]any{
			"id":    "{{id}}",
			"count": 2,
		},
		"list": []any{
			"{{name}}",
			map[string]any{"id": "{{id}}"},
			7,
		},
	}

	got := applyStringTemplates(input, replacements)
	assert.Equal(t, "Hello Ada", got["title"])

	nested := got["nested"].(map[string]any)
	assert.Equal(t, "42", nested["id"])
	assert.Equal(t, 2, nested["count"])

	list := got["list"].([]any)
	assert.Equal(t, "Ada", list[0])
	assert.Equal(t, "42", list[1].(map[string]any)["id"])
	assert.Equal(t, 7, list[2])

	empty := applyStringTemplates(nil, replacements)
	assert.Empty(t, empty)
}

func TestApplyTriggerContext(t *testing.T) {
	obj := &workflows.Object{ID: "obj-123", Type: enums.WorkflowObjectTypeControl}
	input := TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
		ChangedEdges:  []string{"controls"},
		AddedIDs:      map[string][]string{"controls": {"one"}},
		RemovedIDs:    map[string][]string{"controls": {"two"}},
	}
	existing := models.WorkflowInstanceContext{
		Version: 0,
		Data:    []byte(`{"foo":"bar"}`),
	}

	got := applyTriggerContext(existing, "def-123", obj, input, "user-999")
	assert.Equal(t, 1, got.Version)
	assert.Equal(t, "def-123", got.WorkflowDefinitionID)
	assert.Equal(t, obj.Type, got.ObjectType)
	assert.Equal(t, obj.ID, got.ObjectID)
	assert.Equal(t, input.EventType, got.TriggerEventType)
	assert.Equal(t, input.ChangedFields, got.TriggerChangedFields)
	assert.Equal(t, input.ChangedEdges, got.TriggerChangedEdges)
	assert.Equal(t, input.AddedIDs, got.TriggerAddedIDs)
	assert.Equal(t, input.RemovedIDs, got.TriggerRemovedIDs)
	assert.Equal(t, "user-999", got.TriggerUserID)
	assert.Equal(t, existing.Data, got.Data)
}

func TestApprovalFailureMessage(t *testing.T) {
	assert.Equal(t, approvalRejectedMessage, approvalFailureMessage(AssignmentStatusCounts{
		RejectedRequired: true,
	}, 1))

	assert.Equal(t, approvalQuorumNotMet, approvalFailureMessage(AssignmentStatusCounts{
		Approved: 1,
		Pending:  0,
	}, 2))

	assert.Equal(t, approvalFailedMessage, approvalFailureMessage(AssignmentStatusCounts{}, 0))
}

func TestActionFailureDetails(t *testing.T) {
	obj := &workflows.Object{ID: "obj-1", Type: enums.WorkflowObjectTypeControl}
	err := errors.New("boom")
	details := actionFailureDetails("action", 2, enums.WorkflowActionTypeNotification, obj, err)
	assert.Equal(t, "action", details.ActionKey)
	assert.Equal(t, 2, details.ActionIndex)
	assert.Equal(t, enums.WorkflowActionTypeNotification, details.ActionType)
	assert.Equal(t, obj.ID, details.ObjectID)
	assert.Equal(t, obj.Type, details.ObjectType)
	assert.Equal(t, err.Error(), details.ErrorMessage)

	details = actionFailureDetails("action", 1, enums.WorkflowActionTypeNotification, nil, nil)
	assert.Equal(t, actionExecutionFailedMsg, details.ErrorMessage)
	assert.Empty(t, details.ObjectID)
}

func TestActionIndexOutOfBoundsDetails(t *testing.T) {
	payload := soiree.WorkflowActionStartedPayload{
		ActionIndex: 3,
		ActionType:  enums.WorkflowActionTypeNotification,
		ObjectID:    "obj-2",
		ObjectType:  enums.WorkflowObjectTypeControl,
	}
	details := actionIndexOutOfBoundsDetails(payload)
	assert.Equal(t, 3, details.ActionIndex)
	assert.Equal(t, enums.WorkflowActionTypeNotification, details.ActionType)
	assert.Equal(t, "obj-2", details.ObjectID)
	assert.Equal(t, enums.WorkflowObjectTypeControl, details.ObjectType)
	assert.Equal(t, actionIndexOutOfBoundsMsg, details.ErrorMessage)
}

func TestWorkflowEventTypeFromEntOperation(t *testing.T) {
	assert.Equal(t, "UPDATE", workflowEventTypeFromEntOperation(ent.OpUpdate.String()))
	assert.Equal(t, "UPDATE", workflowEventTypeFromEntOperation(ent.OpUpdateOne.String()))
	assert.Equal(t, "CREATE", workflowEventTypeFromEntOperation(ent.OpCreate.String()))
	assert.Equal(t, "DELETE", workflowEventTypeFromEntOperation(ent.OpDelete.String()))
	assert.Equal(t, "DELETE", workflowEventTypeFromEntOperation(ent.OpDeleteOne.String()))
	assert.Equal(t, "CUSTOM", workflowEventTypeFromEntOperation("CUSTOM"))
}

func TestCELEvaluatorValidateAndEvaluate(t *testing.T) {
	cfg := workflows.NewDefaultConfig()
	env, err := workflows.NewCELEnvWithConfig(cfg)
	assert.NoError(t, err)

	eval := NewCELEvaluator(env, cfg)

	// Valid expression should pass validation
	err = eval.Validate("1 == 1")
	assert.NoError(t, err)

	// Invalid expression should fail validation
	err = eval.Validate("1 +")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCELCompilationFailed)

	ok, err := eval.Evaluate(context.Background(), "1 == 1", map[string]any{})
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = eval.Evaluate(context.Background(), "1 + 1", map[string]any{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCELTypeMismatch)
	assert.False(t, ok)
}

func TestGetExecutedNotifications(t *testing.T) {
	listeners := &WorkflowListeners{}

	assert.Empty(t, listeners.getExecutedNotifications(nil))

	instance := &generated.WorkflowInstance{
		Context: models.WorkflowInstanceContext{
			Data: []byte(`{"executed_notifications":["one","two"]}`),
		},
	}
	got := listeners.getExecutedNotifications(instance)
	assert.True(t, got["one"])
	assert.True(t, got["two"])

	instance.Context.Data = []byte(`{"executed_notifications":["alpha", 3, "beta"]}`)
	got = listeners.getExecutedNotifications(instance)
	assert.True(t, got["alpha"])
	assert.True(t, got["beta"])

	instance.Context.Data = []byte(`{"executed_notifications":null}`)
	got = listeners.getExecutedNotifications(instance)
	assert.Empty(t, got)

	instance.Context.Data = []byte(`not-json`)
	got = listeners.getExecutedNotifications(instance)
	assert.Empty(t, got)
}
