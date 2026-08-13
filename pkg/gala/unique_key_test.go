package gala

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
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

	if _, err := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{Kind: JobKindSystem}); err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	if got := dispatcher.envelopes[0].Headers.UniqueKey; got != "key:loop-a" {
		t.Fatalf("UniqueKey = %q, want derived key", got)
	}
}

// TestEmitRawPayloadDerivesTopicUniqueKey proves raw and typed emits derive the same unique key
func TestEmitRawPayloadDerivesTopicUniqueKey(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)
	topic := registerUniqueTopic(t, runtime, "runtime.test.unique.raw")

	if _, err := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{Kind: JobKindSystem}); err != nil {
		t.Fatalf("unexpected typed emit error: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "loop-a"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	if _, err := runtime.EmitWithHeaders(context.Background(), topic.Name, nil, Headers{Kind: JobKindSystem}, WithRawPayload(encodedPayload)); err != nil {
		t.Fatalf("unexpected raw emit error: %v", err)
	}

	typedKey := dispatcher.envelopes[0].Headers.UniqueKey
	rawKey := dispatcher.envelopes[1].Headers.UniqueKey

	if typedKey != "key:loop-a" {
		t.Fatalf("typed UniqueKey = %q, want derived key", typedKey)
	}

	if rawKey != typedKey {
		t.Fatalf("raw UniqueKey = %q, want same key as typed emit %q", rawKey, typedKey)
	}
}

func TestEmitSkipUniqueKeySuppressesDerivation(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)
	topic := registerUniqueTopic(t, runtime, "runtime.test.unique.skip")

	if _, err := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{SkipUniqueKey: true, Kind: JobKindSystem}); err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	if got := dispatcher.envelopes[0].Headers.UniqueKey; got != "" {
		t.Fatalf("UniqueKey = %q, want empty for skipped derivation", got)
	}
}

func TestEmitExplicitUniqueKeyWins(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)
	topic := registerUniqueTopic(t, runtime, "runtime.test.unique.explicit")

	if _, err := runtime.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{UniqueKey: "explicit", Kind: JobKindSystem}); err != nil {
		t.Fatalf("unexpected emit error: %v", err)
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
		if _, err := runtime.EmitWithHeaders(ctx, topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{Kind: JobKindSystem}); err != nil {
			t.Fatalf("unexpected emit error: %v", err)
		}
	}

	if got := countJobs(t); got != 1 {
		t.Fatalf("live jobs = %d, want duplicate emit collapsed to 1", got)
	}

	if _, err := runtime.EmitWithHeaders(ctx, topic.Name, runtimeTestPayload{Message: "loop-a"}, Headers{SkipUniqueKey: true, Kind: JobKindSystem}); err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	if got := countJobs(t); got != 2 {
		t.Fatalf("live jobs = %d, want skip-key emit inserted alongside", got)
	}

	if _, err := runtime.EmitWithHeaders(ctx, topic.Name, runtimeTestPayload{Message: "loop-b"}, Headers{Kind: JobKindSystem}); err != nil {
		t.Fatalf("unexpected emit error: %v", err)
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

// capturingInsertClient records the insert opts passed to River
type capturingInsertClient struct {
	opts *river.InsertOpts
}

func (c *capturingInsertClient) Insert(_ context.Context, _ river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	c.opts = opts

	return &rivertype.JobInsertResult{Job: &rivertype.JobRow{}}, nil
}

func TestUniqueOnceExtendsByStateToTerminal(t *testing.T) {
	t.Parallel()

	client := &capturingInsertClient{}

	d, err := newRiverDispatcher(client, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := d.Dispatch(context.Background(), Envelope{Headers: Headers{UniqueKey: "once", UniqueOnce: true}}); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	states := map[rivertype.JobState]bool{}
	for _, state := range client.opts.UniqueOpts.ByState {
		states[state] = true
	}

	for _, want := range []rivertype.JobState{rivertype.JobStateCompleted, rivertype.JobStateCancelled, rivertype.JobStateDiscarded, rivertype.JobStateRunning} {
		if !states[want] {
			t.Fatalf("ByState missing %s with UniqueOnce set", want)
		}
	}

	if err := d.Dispatch(context.Background(), Envelope{Headers: Headers{UniqueKey: "live-only"}}); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	for _, state := range client.opts.UniqueOpts.ByState {
		if state == rivertype.JobStateCompleted {
			t.Fatal("ByState includes completed without UniqueOnce")
		}
	}
}
