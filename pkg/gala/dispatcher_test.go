package gala

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// riverTestInsertClient records Insert invocations for tests
type riverTestInsertClient struct {
	called   int
	lastArgs river.JobArgs
	lastOpts *river.InsertOpts
	result   *rivertype.JobInsertResult
	err      error
}

// Insert records call metadata and returns the configured error
func (c *riverTestInsertClient) Insert(_ context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	c.called++
	c.lastArgs = args
	if opts != nil {
		copied := *opts
		c.lastOpts = &copied
	} else {
		c.lastOpts = nil
	}

	if c.err != nil {
		return nil, c.err
	}

	if c.result != nil {
		return c.result, nil
	}

	return &rivertype.JobInsertResult{Job: &rivertype.JobRow{
		ID:    1,
		Kind:  args.Kind(),
		State: rivertype.JobStateAvailable,
	}}, nil
}

// newDispatchTestGala builds a durable-mode gala around a stub insert client
func newDispatchTestGala(client riverInsertClient, defaultQueue string) *Gala {
	return &Gala{insertClient: client, defaultQueue: defaultQueue}
}

// TestRiverDispatchArgsRoundTrip verifies args envelope encode/decode round-trips.
func TestRiverDispatchArgsRoundTrip(t *testing.T) {
	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.roundtrip"),
		Payload: []byte(`{"message":"hello"}`),
	}

	args, err := newRiverDispatchArgs(envelope)
	if err != nil {
		t.Fatalf("unexpected args build error: %v", err)
	}

	if args.Kind() != riverDispatchJobKind {
		t.Fatalf("unexpected job kind %q", args.Kind())
	}

	decoded, err := decodeDispatchEnvelope(args.Envelope)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	if decoded.Topic != envelope.Topic {
		t.Fatalf("unexpected decoded topic %q", decoded.Topic)
	}

	if string(decoded.Payload) != string(envelope.Payload) {
		t.Fatalf("unexpected decoded payload %q", string(decoded.Payload))
	}
}

// TestRiverDispatcherDispatchInsertsWithQueueMapping verifies queue selection and job insertion.
func TestRiverDispatcherDispatchInsertsWithQueueMapping(t *testing.T) {
	client := &riverTestInsertClient{}
	runtime := newDispatchTestGala(client, "queue_workflow")

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.durable"),
		Payload: []byte(`{"message":"hello"}`),
	}

	if err := runtime.dispatchDurable(context.Background(), envelope); err != nil {
		t.Fatalf("unexpected durable dispatch error: %v", err)
	}

	if client.called != 1 {
		t.Fatalf("expected one insert call, got %d", client.called)
	}

	if client.lastOpts == nil || client.lastOpts.Queue != "queue_workflow" {
		t.Fatalf("unexpected queue opts: %#v", client.lastOpts)
	}

	insertedArgs, ok := client.lastArgs.(kindedEnvelopeArgs)
	if !ok {
		t.Fatalf("unexpected args type %T", client.lastArgs)
	}

	decoded, err := decodeDispatchEnvelope(insertedArgs.Envelope)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	if decoded.Topic != envelope.Topic {
		t.Fatalf("unexpected decoded topic %q", decoded.Topic)
	}
}

func TestRiverDispatcherReportsDuplicateHolder(t *testing.T) {
	holder := &rivertype.JobRow{
		ID:    42,
		Kind:  Mutation.Kind(),
		State: rivertype.JobStateScheduled,
	}
	client := &riverTestInsertClient{result: &rivertype.JobInsertResult{
		Job:                      holder,
		UniqueSkippedAsDuplicate: true,
	}}
	runtime := newDispatchTestGala(client, "events")
	runtime.kindQueues = map[string]string{Mutation.Kind(): Mutation.Queue()}

	result, err := runtime.insertEnvelope(context.Background(), Envelope{
		Topic: TopicName("mutation.gala.test.duplicate_holder"),
		Headers: Headers{
			Kind:      Mutation.Kind(),
			UniqueKey: "duplicate",
		},
	})
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}
	if result.inserted {
		t.Fatal("expected duplicate dispatch outcome")
	}
	if result.holder != holder || result.holder.State != rivertype.JobStateScheduled {
		t.Fatalf("unexpected duplicate holder: %#v", result.holder)
	}
}

func TestRiverDispatcherRejectsTerminalFailedDuplicateHolder(t *testing.T) {
	client := &riverTestInsertClient{result: &rivertype.JobInsertResult{
		Job: &rivertype.JobRow{
			ID:    42,
			Kind:  Mutation.Kind(),
			State: rivertype.JobStateCancelled,
		},
		UniqueSkippedAsDuplicate: true,
	}}
	runtime := newDispatchTestGala(client, "events")
	runtime.kindQueues = map[string]string{Mutation.Kind(): Mutation.Queue()}

	err := runtime.dispatchDurable(context.Background(), Envelope{
		Topic: TopicName("mutation.gala.test.cancelled_duplicate_holder"),
		Headers: Headers{
			Kind:       Mutation.Kind(),
			UniqueKey:  "duplicate",
			UniqueOnce: true,
		},
	})
	if !errors.Is(err, ErrRiverDispatchInsertFailed) {
		t.Fatalf("expected cancelled duplicate holder to fail dispatch, got %v", err)
	}
}

// TestRiverDispatchWorkerWorkDispatchesEnvelope verifies worker decoding and runtime dispatch.
func TestRiverDispatchWorkerWorkDispatchesEnvelope(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("gala.test.worker")}
	if err := registerTopic(runtime.registry, topic); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	called := 0
	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "gala.test.worker.listener",
		Handle: func(_ HandlerContext, payload runtimeTestPayload) error {
			called++
			if payload.Message != "from-worker" {
				t.Fatalf("unexpected payload message %q", payload.Message)
			}

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "from-worker"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	envelope := Envelope{
		ID:              NewEventID(),
		Topic:           topic.Name,
		Payload:         encodedPayload,
		ContextSnapshot: testCallerSnapshot(t, runtime),
	}

	args, err := newRiverDispatchArgs(envelope)
	if err != nil {
		t.Fatalf("failed to build river args: %v", err)
	}

	worker := newRiverDispatchWorker(func() *Gala {
		return runtime
	})

	job := &river.Job[EnvelopeArgs]{Args: args}
	if err := worker.Work(context.Background(), job); err != nil {
		t.Fatalf("unexpected worker error: %v", err)
	}

	if called != 1 {
		t.Fatalf("expected listener to be called once, got %d", called)
	}
}

// TestRiverDispatchWorkerPreservesCancellation verifies River's worker timeout
// reaches listeners instead of being removed before dispatch.
func TestRiverDispatchWorkerPreservesCancellation(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("gala.test.worker.cancellation")}
	if err := registerTopic(runtime.registry, topic); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := attachListener(runtime, Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "gala.test.worker.cancellation.listener",
		Handle: func(handlerContext HandlerContext, _ runtimeTestPayload) error {
			return handlerContext.Context.Err()
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, err := json.Marshal(runtimeTestPayload{Message: "cancelled"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	envelope := Envelope{
		ID:              NewEventID(),
		Topic:           topic.Name,
		Payload:         encodedPayload,
		ContextSnapshot: testCallerSnapshot(t, runtime),
	}
	args, err := newRiverDispatchArgs(envelope)
	if err != nil {
		t.Fatalf("failed to build river args: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker := newRiverDispatchWorker(func() *Gala { return runtime })
	err = worker.Work(ctx, &river.Job[EnvelopeArgs]{Args: args})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected listener to observe context cancellation, got %v", err)
	}
}

func TestRiverDispatchArgsDecodeEnvelopeErrors(t *testing.T) {
	_, err := decodeDispatchEnvelope(nil)
	if !errors.Is(err, ErrRiverDispatchJobEnvelopeRequired) {
		t.Fatalf("expected ErrRiverDispatchJobEnvelopeRequired, got %v", err)
	}

	_, err = decodeDispatchEnvelope([]byte("{bad"))
	if !errors.Is(err, ErrRiverEnvelopeDecodeFailed) {
		t.Fatalf("expected ErrRiverEnvelopeDecodeFailed, got %v", err)
	}
}

func TestRiverDispatcherQueueSelectionUsesCustomDefaultQueue(t *testing.T) {
	client := &riverTestInsertClient{}
	runtime := newDispatchTestGala(client, "queue_custom_default")

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.queue_selection"),
		Payload: []byte(`{"message":"hello"}`),
	}

	if err := runtime.dispatchDurable(context.Background(), envelope); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil || client.lastOpts.Queue != "queue_custom_default" {
		t.Fatalf("expected custom default queue, got %#v", client.lastOpts)
	}
}

func TestRiverDispatcherPassesHeaderScheduledAt(t *testing.T) {
	client := &riverTestInsertClient{}
	runtime := newDispatchTestGala(client, "queue_custom_default")

	scheduledAt := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.scheduled_at"),
		Payload: []byte(`{"message":"hello"}`),
		Headers: Headers{
			ScheduledAt: &scheduledAt,
		},
	}

	if err := runtime.dispatchDurable(context.Background(), envelope); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil || !client.lastOpts.ScheduledAt.Equal(scheduledAt) {
		t.Fatalf("expected ScheduledAt %v, got %v", scheduledAt, client.lastOpts.ScheduledAt)
	}
}

func TestRiverDispatcherOmitsScheduledAtWhenNil(t *testing.T) {
	client := &riverTestInsertClient{}
	runtime := newDispatchTestGala(client, "queue_custom_default")

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.no_scheduled_at"),
		Payload: []byte(`{"message":"hello"}`),
	}

	if err := runtime.dispatchDurable(context.Background(), envelope); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil {
		t.Fatal("expected insert opts to be set")
	}

	if !client.lastOpts.ScheduledAt.IsZero() {
		t.Fatalf("expected zero ScheduledAt, got %v", client.lastOpts.ScheduledAt)
	}
}

// TestRiverDispatcherKindRouting verifies registered kinds route to their queue and
// unregistered kinds fall back to the legacy kind on the default queue.
func TestRiverDispatcherKindRouting(t *testing.T) {
	client := &riverTestInsertClient{}
	runtime := newDispatchTestGala(client, "events")
	runtime.kindQueues = map[string]string{"gala.mutation": "gala_mutation"}

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.kinds"),
		Payload: []byte(`{}`),
		Headers: Headers{Kind: "gala.mutation"},
	}

	if err := runtime.dispatchDurable(context.Background(), envelope); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if kind := client.lastArgs.Kind(); kind != "gala.mutation" {
		t.Fatalf("expected registered kind on insert, got %q", kind)
	}

	if client.lastOpts.Queue != "gala_mutation" {
		t.Fatalf("expected kind queue, got %q", client.lastOpts.Queue)
	}

	envelope.Headers.Kind = "gala.unknown"
	if err := runtime.dispatchDurable(context.Background(), envelope); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if kind := client.lastArgs.Kind(); kind != riverDispatchJobKind {
		t.Fatalf("expected legacy fallback kind, got %q", kind)
	}

	if client.lastOpts.Queue != "events" {
		t.Fatalf("expected default queue for fallback, got %q", client.lastOpts.Queue)
	}
}

// TestEnvelopeKindAliases verifies registered kinds surface as worker kind aliases.
func TestEnvelopeKindAliases(t *testing.T) {
	registerEnvelopeKinds([]string{"gala.test.alias", "gala.test.alias"})

	aliases := EnvelopeArgs{}.KindAliases()
	count := 0

	for _, alias := range aliases {
		if alias == "gala.test.alias" {
			count++
		}
	}

	if count != 1 {
		t.Fatalf("expected exactly one alias registration, got %d in %v", count, aliases)
	}
}

// TestNamespaceQueue verifies queue name derivation satisfies the river queue grammar.
func TestNamespaceQueue(t *testing.T) {
	if got := IntegrationRun.Queue(); got != "gala_integration_run" {
		t.Fatalf("unexpected queue name %q", got)
	}
}

// TestSoftStopKeepsInFlightJobsRunning verifies cancelling the workers' start context
// begins a soft stop instead of cancelling in-flight listeners
func TestSoftStopKeepsInFlightJobsRunning(t *testing.T) {
	fixture := NewTestGala(t, WithTestStartWorkers(false))

	topic := NamespacedTopic[runtimeTestPayload](Mutation, "GalaSoftStop")

	started := make(chan struct{})
	release := make(chan struct{})
	ctxErr := make(chan error, 1)

	if _, err := Register(fixture.Gala, Definition[runtimeTestPayload]{
		Topic: topic,
		Handle: func(handlerContext HandlerContext, _ runtimeTestPayload) error {
			close(started)
			<-release
			ctxErr <- handlerContext.Context.Err()

			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	startCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := fixture.Gala.StartWorkers(startCtx); err != nil {
		t.Fatalf("failed to start workers: %v", err)
	}

	t.Cleanup(func() { _ = fixture.Gala.StopWorkers(context.Background()) })

	if _, err := fixture.Gala.EmitWithHeaders(context.Background(), topic.Name, runtimeTestPayload{Message: "soft"}, Headers{}); err != nil {
		t.Fatalf("failed to emit event: %v", err)
	}

	select {
	case <-started:
	case <-time.After(30 * time.Second):
		t.Fatal("listener never started")
	}

	cancel()
	close(release)

	select {
	case err := <-ctxErr:
		if err != nil {
			t.Fatalf("expected the in-flight listener context to stay live through the soft stop, got %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("listener never completed")
	}
}
