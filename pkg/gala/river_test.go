package gala

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// riverTestInsertClient records Insert invocations for tests.
type riverTestInsertClient struct {
	called      int
	txCalled    int
	lastArgs    river.JobArgs
	lastTxArgs  river.JobArgs
	lastOpts    *river.InsertOpts
	lastTxOpts  *river.InsertOpts
	lastTx      pgx.Tx
	err         error
	insertTxErr error
}

// Insert records call metadata and returns the configured error.
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

// InsertTx records transactional insertion metadata and returns the configured error.
func (c *riverTestInsertClient) InsertTx(
	_ context.Context,
	tx pgx.Tx,
	args river.JobArgs,
	opts *river.InsertOpts,
) (*rivertype.JobInsertResult, error) {
	c.txCalled++
	c.lastTx = tx
	c.lastTxArgs = args
	if opts != nil {
		copied := *opts
		c.lastTxOpts = &copied
	} else {
		c.lastTxOpts = nil
	}

	if c.insertTxErr != nil {
		return nil, c.insertTxErr
	}

	return &rivertype.JobInsertResult{}, nil
}

// riverTestTx is a minimal pgx.Tx implementation used to exercise InsertTx dispatch paths.
type riverTestTx struct{}

// Begin starts a nested test transaction.
func (riverTestTx) Begin(context.Context) (pgx.Tx, error) { return riverTestTx{}, nil }

// Commit commits the test transaction.
func (riverTestTx) Commit(context.Context) error { return nil }

// Rollback rolls back the test transaction.
func (riverTestTx) Rollback(context.Context) error { return nil }

// CopyFrom implements pgx.Tx.
func (riverTestTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

// SendBatch implements pgx.Tx.
func (riverTestTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }

// LargeObjects implements pgx.Tx.
func (riverTestTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

// Prepare implements pgx.Tx.
func (riverTestTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

// Exec implements pgx.Tx.
func (riverTestTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

// Query implements pgx.Tx.
func (riverTestTx) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }

// QueryRow implements pgx.Tx.
func (riverTestTx) QueryRow(context.Context, string, ...any) pgx.Row { return nil }

// Conn implements pgx.Tx.
func (riverTestTx) Conn() *pgx.Conn { return nil }

// TestNewRiverDispatcherRequiresJobClient verifies construction fails without a job client.
func TestNewRiverDispatcherRequiresJobClient(t *testing.T) {
	_, err := NewRiverDispatcher(RiverDispatcherOptions{})
	if !errors.Is(err, ErrRiverJobClientRequired) {
		t.Fatalf("expected ErrRiverJobClientRequired, got %v", err)
	}
}

// TestRiverDispatchArgsRoundTrip verifies args envelope encode/decode round-trips.
func TestRiverDispatchArgsRoundTrip(t *testing.T) {
	envelope := Envelope{
		ID:            NewEventID(),
		Topic:         TopicName("gala.test.roundtrip"),
		SchemaVersion: 1,
		Payload:       []byte(`{"message":"hello"}`),
	}

	args, err := NewRiverDispatchArgs(envelope, "queue_a")
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

// TestRiverDispatcherDispatchDurableInsertsWithQueueMapping verifies queue selection and job insertion.
func TestRiverDispatcherDispatchDurableInsertsWithQueueMapping(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := NewRiverDispatcher(RiverDispatcherOptions{
		JobClient: client,
		QueueByClass: map[QueueClass]string{
			QueueClassWorkflow: "queue_workflow",
		},
	})
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:            NewEventID(),
		Topic:         TopicName("gala.test.durable"),
		SchemaVersion: 1,
		Payload:       []byte(`{"message":"hello"}`),
	}

	err = dispatcher.DispatchDurable(context.Background(), envelope, TopicPolicy{QueueClass: QueueClassWorkflow})
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

// TestRiverDispatcherDispatchDurableUsesInsertTxWhenContextCarriesTx verifies InsertTx is used when tx is present.
func TestRiverDispatcherDispatchDurableUsesInsertTxWhenContextCarriesTx(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := NewRiverDispatcher(RiverDispatcherOptions{
		JobClient: client,
		QueueByClass: map[QueueClass]string{
			QueueClassWorkflow: "queue_workflow",
		},
	})
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:            NewEventID(),
		Topic:         TopicName("gala.test.insert_tx"),
		SchemaVersion: 1,
		Payload:       []byte(`{"message":"hello"}`),
	}

	dispatchCtx := WithPGXTx(context.Background(), riverTestTx{})
	err = dispatcher.DispatchDurable(dispatchCtx, envelope, TopicPolicy{QueueClass: QueueClassWorkflow})
	if err != nil {
		t.Fatalf("unexpected durable dispatch error: %v", err)
	}

	if client.txCalled != 1 {
		t.Fatalf("expected one InsertTx call, got %d", client.txCalled)
	}

	if client.called != 0 {
		t.Fatalf("expected zero Insert calls when tx is present, got %d", client.called)
	}

	if client.lastTxOpts == nil || client.lastTxOpts.Queue != "queue_workflow" {
		t.Fatalf("unexpected tx queue opts: %#v", client.lastTxOpts)
	}
}

// TestRiverDispatchWorkerWorkDispatchesEnvelope verifies worker decoding and runtime dispatch.
func TestRiverDispatchWorkerWorkDispatchesEnvelope(t *testing.T) {
	runtime, err := NewRuntime(RuntimeOptions{})
	if err != nil {
		t.Fatalf("failed to build runtime: %v", err)
	}

	topic := Topic[runtimeTestPayload]{Name: TopicName("gala.test.worker")}
	if err := (Registration[runtimeTestPayload]{
		Topic: topic,
		Codec: JSONCodec[runtimeTestPayload]{},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	called := 0
	if _, err := (Definition[runtimeTestPayload]{
		Topic: topic,
		Name:  "gala.test.worker.listener",
		Handle: func(_ HandlerContext, payload runtimeTestPayload) error {
			called++
			if payload.Message != "from-worker" {
				t.Fatalf("unexpected payload message %q", payload.Message)
			}

			return nil
		},
	}).Register(runtime.Registry()); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	encodedPayload, schemaVersion, err := runtime.Registry().EncodePayload(topic.Name, runtimeTestPayload{Message: "from-worker"})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	envelope := Envelope{
		ID:            NewEventID(),
		Topic:         topic.Name,
		SchemaVersion: schemaVersion,
		Payload:       encodedPayload,
	}

	args, err := NewRiverDispatchArgs(envelope, "default")
	if err != nil {
		t.Fatalf("failed to build river args: %v", err)
	}

	worker := NewRiverDispatchWorker(func() *Runtime {
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
	if !errors.Is(err, ErrRiverRuntimeProviderRequired) {
		t.Fatalf("expected ErrRiverRuntimeProviderRequired, got %v", err)
	}
}

// TestAddRiverDispatchWorkerRequiresWorkers verifies worker registration fails when workers are nil.
func TestAddRiverDispatchWorkerRequiresWorkers(t *testing.T) {
	err := AddRiverDispatchWorker(nil, func() *Runtime { return nil })
	if !errors.Is(err, ErrRiverWorkersRequired) {
		t.Fatalf("expected ErrRiverWorkersRequired, got %v", err)
	}
}

// TestAddRiverDispatchWorkerRequiresRuntimeProvider verifies worker registration validates runtime provider.
func TestAddRiverDispatchWorkerRequiresRuntimeProvider(t *testing.T) {
	err := AddRiverDispatchWorker(river.NewWorkers(), nil)
	if !errors.Is(err, ErrRiverRuntimeProviderRequired) {
		t.Fatalf("expected ErrRiverRuntimeProviderRequired, got %v", err)
	}
}

// TestAddRiverDispatchWorkerRegistersWorker verifies helper-based worker registration succeeds.
func TestAddRiverDispatchWorkerRegistersWorker(t *testing.T) {
	runtime, err := NewRuntime(RuntimeOptions{})
	if err != nil {
		t.Fatalf("failed to build runtime: %v", err)
	}

	err = AddRiverDispatchWorker(river.NewWorkers(), func() *Runtime {
		return runtime
	})
	if err != nil {
		t.Fatalf("unexpected worker registration error: %v", err)
	}
}
