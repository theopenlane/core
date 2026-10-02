package runtime

import (
	"errors"
	"fmt"
	"testing"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// reconcileEnvelope builds an installation-bound envelope for classification tests
func reconcileEnvelope(t *testing.T, integrationID string) operations.ReconcileEnvelope {
	t.Helper()

	oc := types.NewOperationContext("org-1", "DirectorySync", types.IntegrationSource{
		IntegrationID: integrationID,
		DefinitionID:  "test-def",
	})

	return operations.ReconcileEnvelope{OperationContext: oc}
}

func TestRegisterListenersRequiresGala(t *testing.T) {
	t.Parallel()

	err := NewForTesting(registry.New()).registerListeners()
	if !errors.Is(err, operations.ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}
}

func TestReconcileShouldCancelClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "plain error keeps the loop running",
			err:  ErrTestCycle,
			want: false,
		},
		{
			name: "operation disabled stops the loop",
			err:  fmt.Errorf("cycle: %w", operations.ErrOperationDisabled),
			want: true,
		},
		{
			name: "unhealthy failure stops the loop",
			err:  fmt.Errorf("cycle: %w", types.Unhealthy(ErrTestCycle, "needs reauthorization")),
			want: true,
		},
		{
			name: "degraded failure stops the loop",
			err:  fmt.Errorf("cycle: %w", types.Degraded(ErrTestCycle, "missing operation permission")),
			want: true,
		},
		{
			name: "definition no longer registered stops the loop",
			err:  fmt.Errorf("cycle: %w", registry.ErrDefinitionNotFound),
			want: true,
		},
		{
			name: "operation no longer registered stops the loop",
			err:  fmt.Errorf("cycle: %w", registry.ErrOperationNotFound),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := reconcileShouldCancel(t.Context(), registry.New(), reconcileEnvelope(t, "install-1"), tt.err)
			if got != tt.want {
				t.Fatalf("reconcileShouldCancel = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcileShouldCancelNotFoundScoping(t *testing.T) {
	t.Parallel()

	notFound := &ent.NotFoundError{}

	if !reconcileShouldCancel(t.Context(), registry.New(), reconcileEnvelope(t, "install-1"), notFound) {
		t.Fatal("expected not-found to stop an installation-bound loop")
	}

	if reconcileShouldCancel(t.Context(), registry.New(), reconcileEnvelope(t, ""), notFound) {
		t.Fatal("expected not-found to keep a runtime-bound sweep running")
	}
}
