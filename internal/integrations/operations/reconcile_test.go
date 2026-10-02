package operations

import (
	"testing"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// reconcileEnvelope builds an installation-bound envelope for key tests
func reconcileEnvelope(t *testing.T, integrationID string) ReconcileEnvelope {
	t.Helper()

	oc := types.NewOperationContext("org-1", "DirectorySync", types.IntegrationSource{
		IntegrationID: integrationID,
		DefinitionID:  "test-def",
	})

	return ReconcileEnvelope{OperationContext: oc}
}

func TestReconcileUniqueKey(t *testing.T) {
	t.Parallel()

	installBound := reconcileEnvelope(t, "install-1")
	if got := ReconcileUniqueKey(installBound); got != "integration.reconcile:install-1:test-def:DirectorySync" {
		t.Fatalf("ReconcileUniqueKey = %q", got)
	}

	runtimeBound := reconcileEnvelope(t, "")
	if got := ReconcileUniqueKey(runtimeBound); got != "integration.reconcile::test-def:DirectorySync" {
		t.Fatalf("ReconcileUniqueKey = %q", got)
	}

	if ReconcileUniqueKey(installBound) == ReconcileUniqueKey(runtimeBound) {
		t.Fatal("installation-bound and runtime-bound keys must differ")
	}
}
