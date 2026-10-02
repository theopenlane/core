package runtime

import (
	"testing"

	"github.com/theopenlane/core/v2/internal/integrations/operations"
)

func TestIngestRunSummary(t *testing.T) {
	t.Parallel()

	result := operations.IngestResult{Attempted: 10, Persisted: 8, Changed: 4, Failed: 2, Removed: 1, Excluded: 3}

	want := "attempted 10, persisted 8, changed 4, failed 2, removed 1, excluded 3"
	if got := IngestRunSummary(result); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

// TestIngestMetricsExcluded verifies the excluded counter surfaces in the run's structured metrics
// payload under the "excluded" key, alongside every other record-count metric
func TestIngestMetricsExcluded(t *testing.T) {
	t.Parallel()

	result := operations.IngestResult{Attempted: 10, Persisted: 8, Changed: 4, Skipped: 1, Failed: 2, Filtered: 1, Removed: 1, Excluded: 3}

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
