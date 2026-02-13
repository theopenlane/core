package hooks

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
)

// TestShouldSkipWorkflowMutationForBypass verifies bypass semantics across workflow and Gala context markers.
func TestShouldSkipWorkflowMutationForBypass(t *testing.T) {
	t.Parallel()

	if shouldSkipWorkflowMutationForBypass(context.Background()) {
		t.Fatal("expected no skip for baseline context")
	}

	bypassCtx := workflows.WithContext(context.Background())
	if !shouldSkipWorkflowMutationForBypass(bypassCtx) {
		t.Fatal("expected bypass context to skip workflow mutation handling")
	}

	allowCtx := workflows.WithAllowWorkflowEventEmission(bypassCtx)
	if shouldSkipWorkflowMutationForBypass(allowCtx) {
		t.Fatal("expected allow-event context to permit workflow mutation handling")
	}

	galaBypassCtx := gala.WithFlag(context.Background(), gala.ContextFlagWorkflowBypass)
	if !shouldSkipWorkflowMutationForBypass(galaBypassCtx) {
		t.Fatal("expected Gala bypass flag to skip workflow mutation handling")
	}

	galaAllowCtx := gala.WithFlag(galaBypassCtx, gala.ContextFlagWorkflowAllowEventEmission)
	if shouldSkipWorkflowMutationForBypass(galaAllowCtx) {
		t.Fatal("expected Gala allow-event flag to permit workflow mutation handling")
	}
}
