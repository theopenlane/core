//go:build test

package integrations

import (
	"testing"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
)

func TestBuilderRegistersAllSurfaces(t *testing.T) {
	reg := registry.New()
	if err := reg.RegisterAll(Builder()); err != nil {
		t.Fatalf("register testdef definition: %v", err)
	}

	names := map[string]struct{}{}
	for _, name := range []string{HealthOp.Name(), RepoSyncOp.Name(), ValidatedOp.Name(), RecurringOp.Name(), ExhaustingOp.Name(), UnresolvableOp.Name()} {
		if name == "" {
			t.Fatal("operation registered under empty name")
		}

		if _, dup := names[name]; dup {
			t.Fatalf("operation name %q is not unique", name)
		}

		names[name] = struct{}{}

		if _, err := reg.Operation(DefinitionID.ID(), name); err != nil {
			t.Fatalf("operation %q not registered: %v", name, err)
		}
	}

	for _, reconcileOp := range []string{RecurringOp.Name(), ExhaustingOp.Name(), UnresolvableOp.Name()} {
		op, err := reg.Operation(DefinitionID.ID(), reconcileOp)
		if err != nil {
			t.Fatalf("operation %q not registered: %v", reconcileOp, err)
		}

		if !op.Policy.Reconcile {
			t.Fatalf("operation %q is not a reconcile operation", reconcileOp)
		}
	}

	if _, err := reg.Client(DefinitionID.ID(), testClient.ID()); err != nil {
		t.Fatalf("test client not registered: %v", err)
	}

	if _, err := reg.Webhook(DefinitionID.ID(), "inbound.events"); err != nil {
		t.Fatalf("webhook contract not registered: %v", err)
	}
}
