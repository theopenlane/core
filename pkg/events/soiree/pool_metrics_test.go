package soiree

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPoolMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	pool := NewPool(WithWorkers(1), WithPoolMetricsRegisterer(reg))
	defer pool.Release()

	_ = pool.SubmitMultipleAndWait([]func(){func() {}})

	if pool.metrics == nil {
		t.Fatal("expected pool metrics to be configured")
	}

	if got := testutil.ToFloat64(pool.metrics.tasksSubmitted); got != 1 {
		t.Fatalf("expected tasksSubmitted to be 1, got %v", got)
	}
	if got := testutil.ToFloat64(pool.metrics.tasksStarted); got != 1 {
		t.Fatalf("expected tasksStarted to be 1, got %v", got)
	}
	if got := testutil.ToFloat64(pool.metrics.tasksCompleted); got != 1 {
		t.Fatalf("expected tasksCompleted to be 1, got %v", got)
	}
	if got := testutil.ToFloat64(pool.metrics.tasksPanicked); got != 0 {
		t.Fatalf("expected tasksPanicked to be 0, got %v", got)
	}
	if got := testutil.ToFloat64(pool.metrics.tasksQueued); got != 0 {
		t.Fatalf("expected tasksQueued to be 0, got %v", got)
	}
	if got := testutil.ToFloat64(pool.metrics.tasksRunning); got != 0 {
		t.Fatalf("expected tasksRunning to be 0, got %v", got)
	}
}
