package reconciler

import (
	"context"
	"testing"

	"github.com/theopenlane/core/pkg/entitlements"
	ent "github.com/theopenlane/ent/generated"
)

func TestNew_MissingDeps(t *testing.T) {
	tests := []struct {
		name        string
		opts        []Option
		expectError bool
	}{
		{
			name:        "missing stripe client",
			opts:        []Option{WithDB(&ent.Client{}), WithStripeClient(nil)},
			expectError: true,
		},
		{
			name:        "missing db",
			opts:        []Option{WithDB(nil), WithStripeClient(&entitlements.StripeClient{})},
			expectError: false,
		},
		{
			name:        "both stripe and db provided",
			opts:        []Option{WithDB(&ent.Client{}), WithStripeClient(&entitlements.StripeClient{})},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.opts...)
			if tt.expectError && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAnalyzeOrgNoSubscription(t *testing.T) {
	WithStripeClient(&entitlements.StripeClient{})
	r, err := New(
		WithDB(&ent.Client{}),
		WithStripeClient(&entitlements.StripeClient{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	org := &ent.Organization{}
	action, err := r.analyzeOrg(context.Background(), org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "create stripe customer & subscription" {
		t.Fatalf("unexpected action: %s", action)
	}
}

func TestReconcileResult(t *testing.T) {
	result := &ReconcileResult{
		Actions: []actionRow{
			{OrgID: "test-org-1", Action: "create stripe customer"},
			{OrgID: "test-org-2", Action: "create stripe subscription"},
		},
	}

	if len(result.Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(result.Actions))
	}

	if result.Actions[0].OrgID != "test-org-1" {
		t.Fatalf("expected first org ID to be 'test-org-1', got %s", result.Actions[0].OrgID)
	}

	if result.Actions[0].Action != "create stripe customer" {
		t.Fatalf("expected first action to be 'create stripe customer', got %s", result.Actions[0].Action)
	}
}
