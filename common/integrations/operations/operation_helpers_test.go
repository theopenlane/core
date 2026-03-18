package operations

import (
	"context"
	"testing"

	"github.com/theopenlane/core/common/integrations/types"
)

func TestOperationFailure(t *testing.T) {
	res, retErr := OperationFailure("failed", nil, nil)
	if res.Status != types.OperationStatusFailed {
		t.Fatalf("expected failed status")
	}
	if res.Summary != "failed" {
		t.Fatalf("expected summary")
	}
	if res.Details != nil {
		t.Fatalf("expected no details when err is nil")
	}
	if retErr != nil {
		t.Fatalf("expected nil error return")
	}

	err := context.Canceled
	res, retErr = OperationFailure("failed", err, nil)
	if res.Details == nil || res.Details["error"] != err.Error() {
		t.Fatalf("expected error details")
	}
	if retErr != err {
		t.Fatalf("expected returned error to match input")
	}

	res, retErr = OperationFailure("with context", err, map[string]any{"region": "us-east-1"})
	if res.Details["region"] != "us-east-1" {
		t.Fatalf("expected region in details")
	}
	if res.Details["error"] != err.Error() {
		t.Fatalf("expected auto-injected error in details")
	}
	if retErr != err {
		t.Fatalf("expected returned error to match input")
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
