package gala

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Provider resolves the gala instance used by River workers
type Provider func() *Gala

// RiverDispatchWorker processes durable gala dispatch jobs from River
type RiverDispatchWorker struct {
	river.WorkerDefaults[RiverDispatchArgs]

	galaProvider Provider
}

// RiverDispatchJobKind is the River job kind used for durable gala dispatch
const RiverDispatchJobKind = "gala_dispatch_v1"

// DefaultQueueName is the default queue used for gala durable dispatch jobs
const DefaultQueueName = "events"

// RiverDispatchArgs stores a JSON-encoded gala envelope for durable dispatch
type RiverDispatchArgs struct {
	// Envelope is the encoded gala envelope payload
	Envelope []byte `json:"envelope"`
}

// RiverInsertClient represents the minimal insert capability required for durable dispatch
type RiverInsertClient interface {
	// Insert inserts a River job with optional insert options
	Insert(context.Context, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// RiverDispatcher dispatches envelopes to River
type RiverDispatcher struct {
	// jobClient is the River client used to insert dispatch jobs
	jobClient RiverInsertClient
	// defaultQueue is the default River queue for dispatch jobs
	defaultQueue string
}

// Dispatcher dispatches envelopes to the configured transport
type Dispatcher interface {
	// Dispatch dispatches an envelope to the configured transport
	Dispatch(context.Context, Envelope) error
}

// NewRiverDispatcher creates a River-backed durable dispatcher
func NewRiverDispatcher(jobClient RiverInsertClient, defaultQueue string) (*RiverDispatcher, error) {
	if jobClient == nil {
		return nil, ErrRiverJobClientRequired
	}

	if defaultQueue == "" {
		defaultQueue = DefaultQueueName
	}

	return &RiverDispatcher{
		jobClient:    jobClient,
		defaultQueue: defaultQueue,
	}, nil
}

// NewRiverDispatchArgs builds River dispatch args from an envelope
func NewRiverDispatchArgs(envelope Envelope) (RiverDispatchArgs, error) {
	encodedEnvelope, err := json.Marshal(envelope)
	if err != nil {
		return RiverDispatchArgs{}, ErrRiverEnvelopeEncodeFailed
	}

	return RiverDispatchArgs{
		Envelope: encodedEnvelope,
	}, nil
}

// NewRiverDispatchWorker creates a RiverDispatchWorker
func NewRiverDispatchWorker(galaProvider Provider) *RiverDispatchWorker {
	return &RiverDispatchWorker{galaProvider: galaProvider}
}

// riverJobMetadata is the JSON structure attached to River jobs for UI visibility
type riverJobMetadata struct {
	// Topic is the gala topic name
	Topic string `json:"topic"`
	// EventID is the gala event identifier
	EventID string `json:"event_id"`
	// Properties contains envelope header properties (entity_id, operation, mutation_type, etc.)
	Properties map[string]string `json:"properties,omitempty"`
}

// Dispatch dispatches an envelope to River for processing by a Worker
func (d *RiverDispatcher) Dispatch(ctx context.Context, envelope Envelope) error {
	args, err := NewRiverDispatchArgs(envelope)
	if err != nil {
		return err
	}

	queueName := strings.TrimSpace(envelope.Headers.Queue)
	if queueName == "" {
		queueName = d.defaultQueue
	}

	insertOpts := &river.InsertOpts{
		Queue: queueName,
		Tags:  envelope.Headers.Tags,
	}

	if envelope.Headers.MaxAttempts > 0 {
		insertOpts.MaxAttempts = envelope.Headers.MaxAttempts
	}

	meta, err := json.Marshal(riverJobMetadata{
		Topic:      string(envelope.Topic),
		EventID:    string(envelope.ID),
		Properties: envelope.Headers.Properties,
	})
	if err != nil {
		return ErrRiverEnvelopeEncodeFailed
	}

	insertOpts.Metadata = meta

	if _, err = d.jobClient.Insert(ctx, args, insertOpts); err != nil {
		return ErrRiverDispatchInsertFailed
	}

	return nil
}

// Kind satisfies river.JobArgs
func (RiverDispatchArgs) Kind() string {
	return RiverDispatchJobKind
}

// DecodeEnvelope decodes the gala envelope from dispatch args
func (a RiverDispatchArgs) DecodeEnvelope() (Envelope, error) {
	var envelope Envelope
	if len(a.Envelope) == 0 {
		return envelope, ErrRiverDispatchJobEnvelopeRequired
	}

	if err := jsonx.RoundTrip(a.Envelope, &envelope); err != nil {
		return envelope, ErrRiverEnvelopeDecodeFailed
	}

	return envelope, nil
}

// Work processes one River dispatch job and invokes Gala dispatch
func (w *RiverDispatchWorker) Work(ctx context.Context, job *river.Job[RiverDispatchArgs]) error {
	if w.galaProvider == nil {
		return ErrRiverGalaProviderRequired
	}

	g := w.galaProvider()
	if g == nil {
		return ErrGalaRequired
	}

	envelope, err := job.Args.DecodeEnvelope()
	if err != nil {
		return err
	}

	return g.DispatchEnvelope(context.WithoutCancel(ctx), envelope)
}
