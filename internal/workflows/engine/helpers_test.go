package engine

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"entgo.io/ent"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/workflows"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
	assert.Check(t, is.Equal("Hello Ada", got["title"]))

	nested := got["nested"].(map[string]any)
	assert.Check(t, is.Equal("42", nested["id"]))
	assert.Check(t, is.DeepEqual(float64(2), nested["count"]))

	list := got["list"].([]any)
	assert.Check(t, is.Equal("Ada", list[0]))
	assert.Check(t, is.Equal("42", list[1].(map[string]any)["id"]))
	assert.Check(t, is.DeepEqual(float64(7), list[2]))

	empty := applyStringTemplates(nil, replacements)
	assert.Check(t, is.Len(empty, 0))
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
		Version:               0,
		ExecutedNotifications: []string{"notify-a"},
	}

	got := applyTriggerContext(existing, "def-123", obj, input, "user-999")
	assert.Check(t, is.Equal(1, got.Version))
	assert.Check(t, is.Equal("def-123", got.WorkflowDefinitionID))
	assert.Check(t, is.DeepEqual(obj.Type, got.ObjectType))
	assert.Check(t, is.DeepEqual(obj.ID, got.ObjectID))
	assert.Check(t, is.DeepEqual(input.EventType, got.TriggerEventType))
	assert.Check(t, is.DeepEqual(input.ChangedFields, got.TriggerChangedFields))
	assert.Check(t, is.DeepEqual(input.ChangedEdges, got.TriggerChangedEdges))
	assert.Check(t, is.DeepEqual(input.AddedIDs, got.TriggerAddedIDs))
	assert.Check(t, is.DeepEqual(input.RemovedIDs, got.TriggerRemovedIDs))
	assert.Check(t, is.Equal("user-999", got.TriggerUserID))
	assert.Check(t, is.DeepEqual(existing.ExecutedNotifications, got.ExecutedNotifications))
}

func TestApprovalFailureMessage(t *testing.T) {
	assert.Check(t, is.DeepEqual(approvalRejectedMessage, approvalFailureMessage(AssignmentStatusCounts{
		RejectedRequired: true,
	}, 1)))

	assert.Check(t, is.DeepEqual(approvalQuorumNotMet, approvalFailureMessage(AssignmentStatusCounts{
		Approved: 1,
		Pending:  0,
	}, 2)))

	assert.Check(t, is.DeepEqual(approvalFailedMessage, approvalFailureMessage(AssignmentStatusCounts{}, 0)))
}

func TestActionFailureDetails(t *testing.T) {
	obj := &workflows.Object{ID: "obj-1", Type: enums.WorkflowObjectTypeControl}
	err := errors.New("boom")
	details := actionFailureDetails("action", 2, enums.WorkflowActionTypeNotification, obj, err)
	assert.Check(t, is.Equal("action", details.ActionKey))
	assert.Check(t, is.Equal(2, details.ActionIndex))
	assert.Check(t, is.DeepEqual(enums.WorkflowActionTypeNotification, details.ActionType))
	assert.Check(t, is.DeepEqual(obj.ID, details.ObjectID))
	assert.Check(t, is.DeepEqual(obj.Type, details.ObjectType))
	assert.Check(t, is.Equal(err.Error(), details.ErrorMessage))

	details = actionFailureDetails("action", 1, enums.WorkflowActionTypeNotification, nil, nil)
	assert.Check(t, is.DeepEqual(actionExecutionFailedMsg, details.ErrorMessage))
	assert.Check(t, is.Len(details.ObjectID, 0))
}

func TestActionIndexOutOfBoundsDetails(t *testing.T) {
	payload := WorkflowActionStartedPayload{
		ActionIndex: 3,
		ActionType:  enums.WorkflowActionTypeNotification,
		ObjectID:    "obj-2",
		ObjectType:  enums.WorkflowObjectTypeControl,
	}
	details := actionIndexOutOfBoundsDetails(payload)
	assert.Check(t, is.Equal(3, details.ActionIndex))
	assert.Check(t, is.DeepEqual(enums.WorkflowActionTypeNotification, details.ActionType))
	assert.Check(t, is.Equal("obj-2", details.ObjectID))
	assert.Check(t, is.DeepEqual(enums.WorkflowObjectTypeControl, details.ObjectType))
	assert.Check(t, is.DeepEqual(actionIndexOutOfBoundsMsg, details.ErrorMessage))
}

func TestWorkflowEventTypeFromEntOperation(t *testing.T) {
	assert.Check(t, is.Equal("UPDATE", workflowEventTypeFromEntOperation(ent.OpUpdate.String())))
	assert.Check(t, is.Equal("UPDATE", workflowEventTypeFromEntOperation(ent.OpUpdateOne.String())))
	assert.Check(t, is.Equal("CREATE", workflowEventTypeFromEntOperation(ent.OpCreate.String())))
	assert.Check(t, is.Equal("DELETE", workflowEventTypeFromEntOperation(ent.OpDelete.String())))
	assert.Check(t, is.Equal("DELETE", workflowEventTypeFromEntOperation(ent.OpDeleteOne.String())))
	assert.Check(t, is.Equal("CUSTOM", workflowEventTypeFromEntOperation("CUSTOM")))
}

func TestCELEvaluatorValidateAndEvaluate(t *testing.T) {
	cfg := workflows.NewDefaultConfig()
	env, err := workflows.NewCELEnv(cfg, workflows.CELScopeAction)
	assert.Check(t, err)

	eval := NewCELEvaluator(env, cfg)

	// Valid expression should pass validation
	err = eval.Validate("1 == 1")
	assert.Check(t, err)

	// Invalid expression should fail validation
	err = eval.Validate("1 +")
	assert.Check(t, err != nil)
	assert.Check(t, is.ErrorIs(err, ErrCELCompilationFailed))

	ok, err := eval.Evaluate(context.Background(), "1 == 1", map[string]any{})
	assert.Check(t, err)
	assert.Check(t, ok)

	ok, err = eval.Evaluate(context.Background(), "1 + 1", map[string]any{})
	assert.Check(t, err != nil)
	assert.Check(t, is.ErrorIs(err, ErrCELTypeMismatch))
	assert.Check(t, !ok)
}

func TestGetExecutedNotifications(t *testing.T) {
	listeners := &WorkflowListeners{}

	assert.Check(t, is.Len(listeners.getExecutedNotifications(nil), 0))

	instance := &generated.WorkflowInstance{
		Context: models.WorkflowInstanceContext{
			ExecutedNotifications: []string{"one", "two"},
		},
	}
	got := listeners.getExecutedNotifications(instance)
	assert.Check(t, got["one"])
	assert.Check(t, got["two"])

	instance.Context.ExecutedNotifications = nil
	got = listeners.getExecutedNotifications(instance)
	assert.Check(t, is.Len(got, 0))
}
