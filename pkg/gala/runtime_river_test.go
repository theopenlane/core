package gala

import (
	"context"
	"errors"
	"testing"
)

func TestNewRiverRuntimeRequiresConnectionURI(t *testing.T) {
	_, err := NewRiverRuntime(context.Background(), RiverRuntimeOptions{})
	if !errors.Is(err, ErrRiverConnectionURIRequired) {
		t.Fatalf("expected ErrRiverConnectionURIRequired, got %v", err)
	}
}

func TestRiverRuntimeLifecycleWithoutWorkersIsNoop(t *testing.T) {
	runtime := &RiverRuntime{workersEnabled: false}

	if err := runtime.StartWorkers(context.Background()); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}

	if err := runtime.StopWorkers(context.Background()); err != nil {
		t.Fatalf("unexpected stop error: %v", err)
	}

	if err := runtime.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
}

func TestRiverRuntimeWorkersRequireJobClient(t *testing.T) {
	runtime := &RiverRuntime{workersEnabled: true}

	if err := runtime.StartWorkers(context.Background()); !errors.Is(err, ErrRiverJobClientRequired) {
		t.Fatalf("expected ErrRiverJobClientRequired on start, got %v", err)
	}

	if err := runtime.StopWorkers(context.Background()); !errors.Is(err, ErrRiverJobClientRequired) {
		t.Fatalf("expected ErrRiverJobClientRequired on stop, got %v", err)
	}
}

func TestQueueByClassForRuntimeDefaultsAndOverrides(t *testing.T) {
	queueByClass := queueByClassForRuntime("events", map[QueueClass]string{
		QueueClassGeneral: "general_override",
	})

	if queueByClass[QueueClassWorkflow] != "events" {
		t.Fatalf("unexpected workflow queue %q", queueByClass[QueueClassWorkflow])
	}

	if queueByClass[QueueClassIntegration] != "events" {
		t.Fatalf("unexpected integration queue %q", queueByClass[QueueClassIntegration])
	}

	if queueByClass[QueueClassGeneral] != "general_override" {
		t.Fatalf("unexpected general queue %q", queueByClass[QueueClassGeneral])
	}
}
