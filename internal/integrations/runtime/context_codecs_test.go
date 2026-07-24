package runtime

import (
	"context"
	"testing"

	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

func TestRegisterContextCodecsRoundTripOperationContext(t *testing.T) {
	t.Parallel()

	g, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode: gala.DispatchModeInMemory,
	})
	if err != nil {
		t.Fatalf("failed to create gala: %v", err)
	}
	t.Cleanup(func() { _ = g.Close() })

	injector := do.New()
	do.ProvideValue(injector, g)
	rt := &Runtime{injector: injector}

	if err := rt.registerContextCodecs(); err != nil {
		t.Fatalf("failed to register codecs: %v", err)
	}

	oc := types.NewOperationContext("org_123", "health.check", types.IntegrationSource{
		IntegrationID: "int_123",
		DefinitionID:  "def_123",
		RunType:       enums.IntegrationRunTypeEvent,
		Workflow: &types.WorkflowMeta{
			InstanceID:  "wf_123",
			ActionKey:   "sync",
			ActionIndex: 2,
		},
	})

	snapshot, err := g.ContextManager().Capture(gala.WithOperationContext(context.Background(), oc))
	if err != nil {
		t.Fatalf("failed to capture snapshot: %v", err)
	}

	restored, err := g.ContextManager().Restore(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("failed to restore snapshot: %v", err)
	}

	restoredOC, ok := gala.OperationContextFromContext(restored)
	if !ok {
		t.Fatal("expected operation context to be restored")
	}

	if restoredOC.OwnerID != oc.OwnerID {
		t.Fatalf("expected owner %q, got %q", oc.OwnerID, restoredOC.OwnerID)
	}
	if restoredOC.EntityID != "int_123" {
		t.Fatalf("expected entity id to round-trip, got %q", restoredOC.EntityID)
	}

	restoredSrc := types.IntegrationSourceFrom(restoredOC)
	if restoredSrc.Workflow == nil || restoredSrc.Workflow.InstanceID != "wf_123" {
		t.Fatalf("expected workflow metadata to round-trip, got %#v", restoredSrc.Workflow)
	}
}
