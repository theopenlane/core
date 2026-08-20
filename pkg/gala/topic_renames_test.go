package gala

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestDispatchTimeTopicRenameFallback verifies a queued pre-rename envelope is delivered
// through Config.TopicRenames when its stored topic is no longer registered
func TestDispatchTimeTopicRenameFallback(t *testing.T) {
	ctx := context.Background()

	legacy := NewTestGala(t, WithTestStartWorkers(false))

	oldTopic := TopicName("GalaRenameFallback")
	newTopic := NamespacedTopic[runtimeTestPayload](Mutation, "GalaRenameFallback")

	snapshot, err := legacy.Gala.contextManager.Capture(ctx)
	if err != nil {
		t.Fatalf("failed to capture snapshot: %v", err)
	}

	if err := legacy.Gala.dispatchDurable(ctx, Envelope{
		ID:              NewEventID(),
		Topic:           oldTopic,
		OccurredAt:      time.Now().UTC(),
		Payload:         []byte(`{"message":"legacy"}`),
		ContextSnapshot: snapshot,
	}); err != nil {
		t.Fatalf("failed to dispatch legacy job: %v", err)
	}

	upgraded, err := NewGala(ctx, Config{
		ConnectionURI: legacy.ConnectionURI,
		QueueName:     defaultTestQueueName,
		WorkerCount:   defaultTestWorkerCount,
		TopicRenames:  map[TopicName]TopicName{oldTopic: newTopic.Name},
	})
	if err != nil {
		t.Fatalf("failed to create upgraded runtime: %v", err)
	}

	t.Cleanup(func() { _ = upgraded.Close() })

	var delivered atomic.Int32

	if _, err := Register(upgraded, Definition[runtimeTestPayload]{
		Topic: newTopic,
		Handle: func(_ HandlerContext, payload runtimeTestPayload) error {
			if payload.Message == "legacy" {
				delivered.Add(1)
			}

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	if err := upgraded.StartWorkers(ctx); err != nil {
		t.Fatalf("failed to start workers: %v", err)
	}

	if err := upgraded.WaitIdle(t.Context()); err != nil {
		t.Fatalf("failed waiting for upgraded runtime: %v", err)
	}

	if got := delivered.Load(); got != 1 {
		t.Fatalf("expected the renamed envelope delivered once, got %d", got)
	}
}
