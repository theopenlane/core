package operations

import (
	"context"
	"testing"

	"github.com/theopenlane/core/common/integrations/types"
)

func TestOperationFailure(t *testing.T) {
	res := OperationFailure("failed", nil)
	if res.Status != types.OperationStatusFailed {
		t.Fatalf("expected failed status")
	}
	if res.Summary != "failed" {
		t.Fatalf("expected summary")
	}
	if res.Details != nil {
		t.Fatalf("expected no details when err is nil")
	}

	err := context.Canceled
	res = OperationFailure("failed", err)
	if res.Details == nil || res.Details["error"] != err.Error() {
		t.Fatalf("expected error details")
	}
}

func TestHealthOperation(t *testing.T) {
	run := func(context.Context, types.OperationInput) (types.OperationResult, error) {
		return types.OperationResult{}, nil
	}
	desc := HealthOperation("health.default", "desc", "client", run)
	if desc.Kind != types.OperationKindHealth {
		t.Fatalf("expected health kind")
	}
	if desc.Run == nil {
		t.Fatalf("expected run function")
	}
}
