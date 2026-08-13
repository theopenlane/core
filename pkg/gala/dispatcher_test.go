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

	return &rivertype.JobInsertResult{}, nil
}

// TestNewRiverDispatcherRequiresJobClient verifies construction fails without a job client.
func TestNewRiverDispatcherRequiresJobClient(t *testing.T) {
	_, err := newRiverDispatcher(nil, "")
	if !errors.Is(err, ErrRiverJobClientRequired) {
		t.Fatalf("expected ErrRiverJobClientRequired, got %v", err)
	}
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
	dispatcher, err := newRiverDispatcher(client, "queue_workflow")
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.durable"),
		Payload: []byte(`{"message":"hello"}`),
	}

	err = dispatcher.Dispatch(context.Background(), envelope)
	if err != nil {
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

// TestRiverDispatchWorkerWorkDispatchesEnvelope verifies worker decoding and runtime dispatch.
func TestRiverDispatchWorkerWorkDispatchesEnvelope(t *testing.T) {
	runtime := newTestGala(t, nil)

	topic := Topic[runtimeTestPayload]{Name: TopicName("gala.test.worker")}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
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

// TestRiverDispatchWorkerRequiresRuntimeProvider verifies runtime provider validation.
func TestRiverDispatchWorkerRequiresRuntimeProvider(t *testing.T) {
	worker := newRiverDispatchWorker(nil)
	job := &river.Job[EnvelopeArgs]{}

	err := worker.Work(context.Background(), job)
	if !errors.Is(err, ErrRiverGalaProviderRequired) {
		t.Fatalf("expected ErrRiverGalaProviderRequired, got %v", err)
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

func TestRiverDispatcherQueueSelection(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := newRiverDispatcher(client, "queue_custom_default")
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.queue_selection"),
		Payload: []byte(`{"message":"hello"}`),
	}

	err = dispatcher.Dispatch(context.Background(), envelope)
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil || client.lastOpts.Queue != "queue_custom_default" {
		t.Fatalf("expected custom default queue, got %#v", client.lastOpts)
	}
}

func TestRiverDispatcherQueueSelectionUsesCustomDefaultQueue(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := newRiverDispatcher(client, "queue_custom_default")
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.queue_builtin_default"),
		Payload: []byte(`{"message":"hello"}`),
	}

	err = dispatcher.Dispatch(context.Background(), envelope)
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil || client.lastOpts.Queue != "queue_custom_default" {
		t.Fatalf("expected custom default queue, got %#v", client.lastOpts)
	}
}

func TestRiverDispatcherQueueSelectionUsesHeaderQueueOverride(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := newRiverDispatcher(client, "queue_custom_default")
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.queue_header_override"),
		Payload: []byte(`{"message":"hello"}`),
		Headers: Headers{
			Queue: "queue_integrations",
		},
	}

	err = dispatcher.Dispatch(context.Background(), envelope)
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil || client.lastOpts.Queue != "queue_integrations" {
		t.Fatalf("expected header queue override, got %#v", client.lastOpts)
	}
}

func TestRiverDispatcherPassesHeaderMaxAttempts(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := newRiverDispatcher(client, "queue_custom_default")
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.max_attempts_override"),
		Payload: []byte(`{"message":"hello"}`),
		Headers: Headers{
			MaxAttempts: 7,
		},
	}

	err = dispatcher.Dispatch(context.Background(), envelope)
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil || client.lastOpts.MaxAttempts != 7 {
		t.Fatalf("expected max attempts override, got %#v", client.lastOpts)
	}
}

func TestRiverDispatcherPassesHeaderScheduledAt(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := newRiverDispatcher(client, "queue_custom_default")
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	scheduledAt := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.scheduled_at"),
		Payload: []byte(`{"message":"hello"}`),
		Headers: Headers{
			ScheduledAt: &scheduledAt,
		},
	}

	err = dispatcher.Dispatch(context.Background(), envelope)
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil || !client.lastOpts.ScheduledAt.Equal(scheduledAt) {
		t.Fatalf("expected ScheduledAt %v, got %v", scheduledAt, client.lastOpts.ScheduledAt)
	}
}

func TestRiverDispatcherOmitsScheduledAtWhenNil(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := newRiverDispatcher(client, "queue_custom_default")
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.no_scheduled_at"),
		Payload: []byte(`{"message":"hello"}`),
	}

	err = dispatcher.Dispatch(context.Background(), envelope)
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil {
		t.Fatal("expected insert opts to be set")
	}

	if !client.lastOpts.ScheduledAt.IsZero() {
		t.Fatalf("expected zero ScheduledAt, got %v", client.lastOpts.ScheduledAt)
	}
}

func TestRiverDispatchWorkerRequiresRuntimeInstance(t *testing.T) {
	worker := newRiverDispatchWorker(func() *Gala {
		return nil
	})

	err := worker.Work(context.Background(), &river.Job[EnvelopeArgs]{})
	if !errors.Is(err, ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}
}

// TestRiverDispatcherKindRouting verifies registered kinds route to their queue and
// unregistered kinds fall back to the legacy kind on the default queue.
func TestRiverDispatcherKindRouting(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := newRiverDispatcher(client, "events")
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	dispatcher.kindQueues = map[string]string{"gala.mutation": "gala_mutation"}

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   TopicName("gala.test.kinds"),
		Payload: []byte(`{}`),
		Headers: Headers{Kind: "gala.mutation"},
	}

	if err := dispatcher.Dispatch(context.Background(), envelope); err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if kind := client.lastArgs.Kind(); kind != "gala.mutation" {
		t.Fatalf("expected registered kind on insert, got %q", kind)
	}

	if client.lastOpts.Queue != "gala_mutation" {
		t.Fatalf("expected kind queue, got %q", client.lastOpts.Queue)
	}

	envelope.Headers.Kind = "gala.unknown"
	if err := dispatcher.Dispatch(context.Background(), envelope); err != nil {
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

// TestQueueNameForKind verifies queue name derivation satisfies the river queue grammar.
func TestQueueNameForKind(t *testing.T) {
	if got := QueueNameForKind("gala.integration.run"); got != "gala_integration_run" {
		t.Fatalf("unexpected queue name %q", got)
	}
}
