package operations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestMarkRunRunning_EmptyRunID(t *testing.T) {
	t.Parallel()

	err := MarkRunRunning(context.Background(), nil, "")
	if !errors.Is(err, ErrRunIDRequired) {
		t.Fatalf("expected ErrRunIDRequired, got %v", err)
	}
}

func TestCompleteRun_EmptyRunID(t *testing.T) {
	t.Parallel()

	err := CompleteRun(context.Background(), nil, "", time.Now(), RunResult{})
	if !errors.Is(err, ErrRunIDRequired) {
		t.Fatalf("expected ErrRunIDRequired, got %v", err)
	}
}

func TestCreatePendingRun_NilInstallation(t *testing.T) {
	t.Parallel()

	_, err := CreatePendingRun(context.Background(), nil, nil, types.OperationRegistration{}, "", nil)
	if !errors.Is(err, ErrInstallationIDRequired) {
		t.Fatalf("expected ErrInstallationIDRequired, got %v", err)
	}
}

func TestRunResult_DefaultStatus(t *testing.T) {
	t.Parallel()

	result := RunResult{}
	if result.Status != "" {
		t.Fatalf("expected empty default status, got %q", result.Status)
	}
}

func TestOperationKind(t *testing.T) {
	t.Parallel()

	ingestOp := types.OperationRegistration{
		IngestHandle: func(context.Context, types.OperationRequest) ([]types.IngestPayloadSet, error) { return nil, nil },
	}
	if kind := operationKind(ingestOp); kind != enums.IntegrationOperationKindSync {
		t.Fatalf("expected %q, got %q", enums.IntegrationOperationKindSync, kind)
	}

	handleOp := types.OperationRegistration{
		Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) { return nil, nil },
	}
	if kind := operationKind(handleOp); kind != enums.IntegrationOperationKindPush {
		t.Fatalf("expected %q, got %q", enums.IntegrationOperationKindPush, kind)
	}
}
