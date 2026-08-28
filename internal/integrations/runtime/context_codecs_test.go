package runtime

import (
	"context"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// codecRoundTripPayload is the fixture payload dispatched through the in-memory runtime
type codecRoundTripPayload struct {
	Message string `json:"message"`
}

func TestDefaultCodecsRoundTripOperationContext(t *testing.T) {
	t.Parallel()

	g, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode: gala.DispatchModeInMemory,
	})
	if err != nil {
		t.Fatalf("failed to create gala: %v", err)
	}
	t.Cleanup(func() { _ = g.Close() })

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

	type observedContext struct {
		oc gala.OperationContext
		ok bool
	}

	observedCh := make(chan observedContext, 1)

	if _, err := gala.Register(g, gala.Definition[codecRoundTripPayload]{
		Topic: gala.Topic[codecRoundTripPayload]{Name: gala.TopicName("runtime.test.codec.roundtrip")},
		Name:  "runtime.test.codec.roundtrip.listener",
		Handle: func(hc gala.HandlerContext, _ codecRoundTripPayload) error {
			restoredOC, ok := gala.OperationContextFromContext(hc.Context)
			observedCh <- observedContext{oc: restoredOC, ok: ok}

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	emitCtx := gala.WithOperationContext(context.Background(), oc)
	if _, err := g.EmitWithHeaders(emitCtx, gala.TopicName("runtime.test.codec.roundtrip"), codecRoundTripPayload{Message: "roundtrip"}, gala.Headers{Kind: gala.IntegrationRun.Kind()}); err != nil {
		t.Fatalf("failed to emit: %v", err)
	}

	if err := g.WaitIdle(t.Context()); err != nil {
		t.Fatalf("failed waiting for Gala runtime: %v", err)
	}

	var observed observedContext

	select {
	case observed = <-observedCh:
	default:
		t.Fatal("expected listener to run")
	}

	if !observed.ok {
		t.Fatal("expected operation context to be restored")
	}

	restoredOC := observed.oc
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
