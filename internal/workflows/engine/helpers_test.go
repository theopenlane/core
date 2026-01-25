package engine

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/workflows"
)

type testEntity struct {
	// Tags holds tag values
	Tags    []string `json:"tags"`
	// Aliases holds alias values
	Aliases []string `json:"aliases"`
	// Count is a numeric field
	Count   int      `json:"count"`
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
