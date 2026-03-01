package operations

import (
	"context"
	"testing"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

func TestOperationFailure(t *testing.T) {
	type regionDetails struct {
		Region string `json:"region"`
	}

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
	details, _ := jsonx.ToMap(res.Details)
	if details == nil || details["error"] != err.Error() {
		t.Fatalf("expected error details")
	}
	if retErr != err {
		t.Fatalf("expected returned error to match input")
	}

	res, retErr = OperationFailure("with context", err, regionDetails{Region: "us-east-1"})
	details, _ = jsonx.ToMap(res.Details)
	if details["region"] != "us-east-1" {
		t.Fatalf("expected region in details")
	}
	if details["error"] != err.Error() {
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
