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

func TestIngestRunSummary(t *testing.T) {
	t.Parallel()

	result := IngestResult{Attempted: 10, Persisted: 8, Changed: 4, Failed: 2, Removed: 1, Excluded: 3}

	want := "attempted 10, persisted 8, changed 4, failed 2, removed 1, excluded 3"
	if got := IngestRunSummary(result); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

// TestIngestMetrics_Excluded verifies the excluded counter surfaces in the run's structured metrics
// payload under the "excluded" key, alongside every other record-count metric
func TestIngestMetrics_Excluded(t *testing.T) {
	t.Parallel()

	result := IngestResult{Attempted: 10, Persisted: 8, Changed: 4, Skipped: 1, Failed: 2, Filtered: 1, Removed: 1, Excluded: 3}

	metrics := IngestMetrics(result)

	want := map[string]any{
		metricAttempted: 10,
		metricPersisted: 8,
		metricChanged:   4,
		metricSkipped:   1,
		metricFailed:    2,
		metricFiltered:  1,
		metricRemoved:   1,
		metricExcluded:  3,
	}

	for key, wantValue := range want {
		if got := metrics[key]; got != wantValue {
			t.Fatalf("metrics[%q]=%v, want %v", key, got, wantValue)
		}
	}

	if len(metrics) != len(want) {
		t.Fatalf("metrics has %d keys, want %d: %v", len(metrics), len(want), metrics)
	}
}
