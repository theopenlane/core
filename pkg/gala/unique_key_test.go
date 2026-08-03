package gala

import (
	"context"
	"testing"
)

// registerUniqueTopic registers a topic deriving its unique key from the payload message
func registerUniqueTopic(t *testing.T, runtime *Gala, name string) Topic[runtimeTestPayload] {
	t.Helper()

	topic := Topic[runtimeTestPayload]{
		Name:      TopicName(name),
		UniqueKey: func(p runtimeTestPayload) string { return "key:" + p.Message },
	}

	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	return topic
}

func TestEmitDerivesTopicUniqueKey(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)
	topic := registerUniqueTopic(t, runtime, "runtime.test.unique")

	receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{})
	if receipt.Err != nil {
		t.Fatalf("unexpected emit error: %v", receipt.Err)
	}

	if got := dispatcher.envelopes[0].Headers.UniqueKey; got != "key:loop-a" {
		t.Fatalf("UniqueKey = %q, want derived key", got)
	}
}

func TestEmitSkipUniqueKeySuppressesDerivation(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)
	topic := registerUniqueTopic(t, runtime, "runtime.test.unique.skip")

	receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{SkipUniqueKey: true})
	if receipt.Err != nil {
		t.Fatalf("unexpected emit error: %v", receipt.Err)
	}

	if got := dispatcher.envelopes[0].Headers.UniqueKey; got != "" {
		t.Fatalf("UniqueKey = %q, want empty for skipped derivation", got)
	}
}

func TestEmitExplicitUniqueKeyWins(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)
	topic := registerUniqueTopic(t, runtime, "runtime.test.unique.explicit")

	receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{UniqueKey: "explicit"})
	if receipt.Err != nil {
		t.Fatalf("unexpected emit error: %v", receipt.Err)
	}

	if got := dispatcher.envelopes[0].Headers.UniqueKey; got != "explicit" {
		t.Fatalf("UniqueKey = %q, want explicit key preserved", got)
	}
}

// TestUniqueKeyEnforcedByRiver proves uniqueness at the database level with a real River
// backend: duplicate emits with the same derived key collapse to one live job, SkipUniqueKey
// bypasses the constraint, and a distinct key inserts freely
func TestUniqueKeyEnforcedByRiver(t *testing.T) {
	fixture := NewTestGala(t, WithTestStartWorkers(false))
	runtime := fixture.Gala

	topic := registerUniqueTopic(t, runtime, "runtime.test.unique.river")
	ctx := context.Background()

	countJobs := func(t *testing.T) int {
		t.Helper()

		count, err := runtime.CountActiveJobsWithMetadata(ctx, `{"topic":"runtime.test.unique.river"}`)
		if err != nil {
			t.Fatalf("counting jobs: %v", err)
		}

		return count
	}

	for range 2 {
		if receipt := runtime.EmitWithHeaders(ctx, topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{}); receipt.Err != nil {
			t.Fatalf("unexpected emit error: %v", receipt.Err)
		}
	}

	if got := countJobs(t); got != 1 {
		t.Fatalf("live jobs = %d, want duplicate emit collapsed to 1", got)
	}

	if receipt := runtime.EmitWithHeaders(ctx, topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{SkipUniqueKey: true}); receipt.Err != nil {
		t.Fatalf("unexpected emit error: %v", receipt.Err)
	}

	if got := countJobs(t); got != 2 {
		t.Fatalf("live jobs = %d, want skip-key emit inserted alongside", got)
	}

	if receipt := runtime.EmitWithHeaders(ctx, topic.Name, runtimeTestPayload{Message: "loop-b"}, Headers{}); receipt.Err != nil {
		t.Fatalf("unexpected emit error: %v", receipt.Err)
	}

	if got := countJobs(t); got != 3 {
		t.Fatalf("live jobs = %d, want distinct key inserted", got)
	}
}

func TestDispatchArgsCarryUniqueKey(t *testing.T) {
	t.Parallel()

	args, err := newRiverDispatchArgs(Envelope{Headers: Headers{UniqueKey: "key:loop-a"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if args.UniqueKey != "key:loop-a" {
		t.Fatalf("UniqueKey = %q, want header key on args", args.UniqueKey)
	}
}
