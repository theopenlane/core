package runtime

import (
	"context"
	"testing"

	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

func TestRegisterContextCodecsRoundTripExecutionMetadata(t *testing.T) {
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

	metadata := types.ExecutionMetadata{
		OwnerID:       "org_123",
		IntegrationID: "int_123",
		DefinitionID:  "def_123",
		Operation:     "health.check",
		RunType:       enums.IntegrationRunTypeEvent,
		Workflow: &types.WorkflowMeta{
			InstanceID:  "wf_123",
			ActionKey:   "sync",
			ActionIndex: 2,
		},
	}

	snapshot, err := g.ContextManager().Capture(types.WithExecutionMetadata(context.Background(), metadata))
	if err != nil {
		t.Fatalf("failed to capture snapshot: %v", err)
	}

	restored, err := g.ContextManager().Restore(context.Background(), snapshot)
	if err != nil {
		t.Fatalf("failed to restore snapshot: %v", err)
	}

	restoredMetadata, ok := types.ExecutionMetadataFromContext(restored)
	if !ok {
		t.Fatal("expected execution metadata to be restored")
	}

	if restoredMetadata.OwnerID != metadata.OwnerID {
		t.Fatalf("expected owner %q, got %q", metadata.OwnerID, restoredMetadata.OwnerID)
	}
	if restoredMetadata.Workflow == nil || restoredMetadata.Workflow.InstanceID != metadata.Workflow.InstanceID {
		t.Fatalf("expected workflow metadata to round-trip, got %#v", restoredMetadata.Workflow)
	}
}
