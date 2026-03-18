package gala

import (
	"context"
	"errors"
	"testing"

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
	_, err := NewRiverDispatcher(nil, "")
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

	args, err := NewRiverDispatchArgs(envelope)
	if err != nil {
		t.Fatalf("unexpected args build error: %v", err)
	}

	if args.Kind() != RiverDispatchJobKind {
		t.Fatalf("unexpected job kind %q", args.Kind())
	}

	decoded, err := args.DecodeEnvelope()
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
	dispatcher, err := NewRiverDispatcher(client, "queue_workflow")
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

	insertedArgs, ok := client.lastArgs.(RiverDispatchArgs)
	if !ok {
		t.Fatalf("unexpected args type %T", client.lastArgs)
	}

	decoded, err := insertedArgs.DecodeEnvelope()
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
	if err := RegisterTopic(runtime.Registry(), Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	called := 0
	if _, err := AttachListener(runtime.Registry(), Definition[runtimeTestPayload]{
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

	encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, runtimeTestPayload{Message: "from-worker"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	envelope := Envelope{
		ID:      NewEventID(),
		Topic:   topic.Name,
		Payload: encodedPayload,
	}

	args, err := NewRiverDispatchArgs(envelope)
	if err != nil {
		t.Fatalf("failed to build river args: %v", err)
	}

	worker := NewRiverDispatchWorker(func() *Gala {
		return runtime
	})

	job := &river.Job[RiverDispatchArgs]{Args: args}
	if err := worker.Work(context.Background(), job); err != nil {
		t.Fatalf("unexpected worker error: %v", err)
	}

	if called != 1 {
		t.Fatalf("expected listener to be called once, got %d", called)
	}
}

// TestRiverDispatchWorkerRequiresRuntimeProvider verifies runtime provider validation.
func TestRiverDispatchWorkerRequiresRuntimeProvider(t *testing.T) {
	worker := NewRiverDispatchWorker(nil)
	job := &river.Job[RiverDispatchArgs]{}

	err := worker.Work(context.Background(), job)
	if !errors.Is(err, ErrRiverGalaProviderRequired) {
		t.Fatalf("expected ErrRiverGalaProviderRequired, got %v", err)
	}
}

func TestRiverDispatchArgsDecodeEnvelopeErrors(t *testing.T) {
	_, err := (RiverDispatchArgs{}).DecodeEnvelope()
	if !errors.Is(err, ErrRiverDispatchJobEnvelopeRequired) {
		t.Fatalf("expected ErrRiverDispatchJobEnvelopeRequired, got %v", err)
	}

	_, err = (RiverDispatchArgs{Envelope: []byte("{bad")}).DecodeEnvelope()
	if !errors.Is(err, ErrRiverEnvelopeDecodeFailed) {
		t.Fatalf("expected ErrRiverEnvelopeDecodeFailed, got %v", err)
	}
}

func TestRiverDispatcherQueueSelection(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := NewRiverDispatcher(client, "queue_custom_default")
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
	dispatcher, err := NewRiverDispatcher(client, "queue_custom_default")
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
	dispatcher, err := NewRiverDispatcher(client, "queue_custom_default")
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
	dispatcher, err := NewRiverDispatcher(client, "queue_custom_default")
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

func TestRiverDispatchWorkerRequiresRuntimeInstance(t *testing.T) {
	worker := NewRiverDispatchWorker(func() *Gala {
		return nil
	})

	err := worker.Work(context.Background(), &river.Job[RiverDispatchArgs]{})
	if !errors.Is(err, ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}
}
