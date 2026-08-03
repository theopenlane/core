package operations

import (
	"context"
	"errors"
	"fmt"
	"testing"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

var errTestCycle = errors.New("provider request failed")

// reconcileEnvelope builds an installation-bound envelope for classification tests
func reconcileEnvelope(t *testing.T, integrationID string) ReconcileEnvelope {
	t.Helper()

	oc := types.NewOperationContext("org-1", "DirectorySync", types.IntegrationSource{
		IntegrationID: integrationID,
		DefinitionID:  "test-def",
	})

	return ReconcileEnvelope{OperationContext: oc}
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
			err:  errTestCycle,
			want: false,
		},
		{
			name: "operation disabled stops the loop",
			err:  fmt.Errorf("cycle: %w", ErrOperationDisabled),
			want: true,
		},
		{
			name: "unhealthy failure stops the loop",
			err:  fmt.Errorf("cycle: %w", types.Unhealthy(errTestCycle, "needs reauthorization")),
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

			got := reconcileShouldCancel(context.Background(), registry.New(), reconcileEnvelope(t, "install-1"), tt.err)
			if got != tt.want {
				t.Fatalf("reconcileShouldCancel = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcileUniqueKey(t *testing.T) {
	t.Parallel()

	installBound := reconcileEnvelope(t, "install-1")
	if got := ReconcileUniqueKey(installBound); got != "reconcile:install-1:test-def:DirectorySync" {
		t.Fatalf("ReconcileUniqueKey = %q", got)
	}

	runtimeBound := reconcileEnvelope(t, "")
	if got := ReconcileUniqueKey(runtimeBound); got != "reconcile::test-def:DirectorySync" {
		t.Fatalf("ReconcileUniqueKey = %q", got)
	}

	if ReconcileUniqueKey(installBound) == ReconcileUniqueKey(runtimeBound) {
		t.Fatal("installation-bound and runtime-bound keys must differ")
	}
}

func TestReconcileShouldCancelNotFoundScoping(t *testing.T) {
	t.Parallel()

	notFound := &ent.NotFoundError{}

	if !reconcileShouldCancel(context.Background(), registry.New(), reconcileEnvelope(t, "install-1"), notFound) {
		t.Fatal("expected not-found to stop an installation-bound loop")
	}

	if reconcileShouldCancel(context.Background(), registry.New(), reconcileEnvelope(t, ""), notFound) {
		t.Fatal("expected not-found to keep a runtime-bound sweep running")
	}
}
