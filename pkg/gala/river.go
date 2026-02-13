package gala

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/samber/lo"
)

const (
	// RiverDispatchJobKind is the River job kind used for durable gala dispatch
	RiverDispatchJobKind = "gala_dispatch_v1"
	// DefaultQueueName is the default queue used for gala durable dispatch jobs
	DefaultQueueName = "events"
)

// RiverInsertClient represents the minimal insert capability required for durable dispatch
type RiverInsertClient interface {
	// Insert inserts a River job with optional insert options
	Insert(context.Context, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// RiverInsertTxClient represents the optional transactional insert capability for durable dispatch
type RiverInsertTxClient interface {
	// InsertTx inserts a River job in an existing transaction
	InsertTx(context.Context, pgx.Tx, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// RiverDispatcherOptions configures a RiverDispatcher
type RiverDispatcherOptions struct {
	// JobClient inserts durable jobs into River
	JobClient RiverInsertClient
	// QueueByClass maps queue classes to concrete queue names
	QueueByClass map[QueueClass]string
	// DefaultQueue is used when class-specific mappings are absent
	DefaultQueue string
}

// RiverDispatcher dispatches envelopes to River
type RiverDispatcher struct {
	jobClient    RiverInsertClient
	jobTxClient  RiverInsertTxClient
	queueByClass map[QueueClass]string
	defaultQueue string
}

// NewRiverDispatcher creates a River-backed durable dispatcher
func NewRiverDispatcher(options RiverDispatcherOptions) (*RiverDispatcher, error) {
	if options.JobClient == nil {
		return nil, ErrRiverJobClientRequired
	}

	defaultQueue := options.DefaultQueue
	if defaultQueue == "" {
		defaultQueue = DefaultQueueName
	}

	queueByClass := map[QueueClass]string{
		QueueClassWorkflow:    defaultQueue,
		QueueClassIntegration: defaultQueue,
		QueueClassGeneral:     defaultQueue,
	}
	if len(options.QueueByClass) > 0 {
		queueByClass = lo.Assign(queueByClass, options.QueueByClass)
	}

	jobTxClient, _ := options.JobClient.(RiverInsertTxClient)

	return &RiverDispatcher{
		jobClient:    options.JobClient,
		jobTxClient:  jobTxClient,
		queueByClass: queueByClass,
		defaultQueue: defaultQueue,
	}, nil
}

// DispatchDurable dispatches an envelope to River using topic queue policy
func (d *RiverDispatcher) DispatchDurable(ctx context.Context, envelope Envelope, policy TopicPolicy) error {
	queueName := d.queueForPolicy(policy)
	args, err := NewRiverDispatchArgs(envelope, queueName)
	if err != nil {
		return err
	}

	if tx, ok := PGXTxFromContext(ctx); ok && d.jobTxClient != nil {
		_, err = d.jobTxClient.InsertTx(ctx, tx, args, &river.InsertOpts{Queue: queueName})
	} else {
		_, err = d.jobClient.Insert(ctx, args, &river.InsertOpts{Queue: queueName})
	}

	if err != nil {
		return ErrRiverDispatchInsertFailed
	}

	return nil
}

// queueForPolicy returns the queue to use for a topic policy
func (d *RiverDispatcher) queueForPolicy(policy TopicPolicy) string {
	if policy.QueueName != "" {
		return policy.QueueName
	}

	queueName, ok := d.queueByClass[policy.QueueClass]
	if !ok || queueName == "" {
		return d.defaultQueue
	}

	return queueName
}

// RiverDispatchArgs stores a JSON-encoded gala envelope for durable dispatch
type RiverDispatchArgs struct {
	// Envelope is the encoded gala envelope payload
	Envelope []byte `json:"envelope"`
	// Queue is the queue used for insertion
	Queue string `json:"queue,omitempty"`
}

// Kind satisfies river.JobArgs
func (RiverDispatchArgs) Kind() string {
	return RiverDispatchJobKind
}

// InsertOpts provides queue defaults for the River job
func (a RiverDispatchArgs) InsertOpts() river.InsertOpts {
	queueName := a.Queue
	if queueName == "" {
		queueName = DefaultQueueName
	}

	return river.InsertOpts{Queue: queueName}
}

// DecodeEnvelope decodes the gala envelope from dispatch args
func (a RiverDispatchArgs) DecodeEnvelope() (Envelope, error) {
	if len(a.Envelope) == 0 {
		return Envelope{}, ErrRiverDispatchJobEnvelopeRequired
	}

	var envelope Envelope
	if err := json.Unmarshal(a.Envelope, &envelope); err != nil {
		return Envelope{}, ErrRiverEnvelopeDecodeFailed
	}

	return envelope, nil
}

// NewRiverDispatchArgs builds River dispatch args from an envelope
func NewRiverDispatchArgs(envelope Envelope, queue string) (RiverDispatchArgs, error) {
	encodedEnvelope, err := json.Marshal(envelope)
	if err != nil {
		return RiverDispatchArgs{}, ErrRiverEnvelopeEncodeFailed
	}

	return RiverDispatchArgs{
		Envelope: encodedEnvelope,
		Queue:    queue,
	}, nil
}

// RuntimeProvider resolves the runtime used by River workers
type RuntimeProvider func() *Runtime

// RiverDispatchWorker processes durable gala dispatch jobs from River
type RiverDispatchWorker struct {
	river.WorkerDefaults[RiverDispatchArgs]

	runtimeProvider RuntimeProvider
}

// NewRiverDispatchWorker creates a RiverDispatchWorker
func NewRiverDispatchWorker(runtimeProvider RuntimeProvider) *RiverDispatchWorker {
	return &RiverDispatchWorker{runtimeProvider: runtimeProvider}
}

// AddRiverDispatchWorker registers the Gala River dispatch worker with a worker registry
func AddRiverDispatchWorker(workers *river.Workers, runtimeProvider RuntimeProvider) error {
	if workers == nil {
		return ErrRiverWorkersRequired
	}

	if runtimeProvider == nil {
		return ErrRiverRuntimeProviderRequired
	}

	return river.AddWorkerSafely(workers, NewRiverDispatchWorker(runtimeProvider))
}

// Work processes one River dispatch job and invokes runtime dispatch
func (w *RiverDispatchWorker) Work(ctx context.Context, job *river.Job[RiverDispatchArgs]) error {
	if w.runtimeProvider == nil {
		return ErrRiverRuntimeProviderRequired
	}

	runtime := w.runtimeProvider()
	if runtime == nil {
		return ErrRuntimeRequired
	}

	envelope, err := job.Args.DecodeEnvelope()
	if err != nil {
		return err
	}

	return runtime.DispatchEnvelope(context.WithoutCancel(ctx), envelope)
}
