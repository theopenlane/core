package gala

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// galaProvider resolves the gala instance used by River workers
type galaProvider func() *Gala

// riverDispatchWorker processes durable gala dispatch jobs from River
type riverDispatchWorker struct {
	river.WorkerDefaults[riverDispatchArgs]

	galaProvider galaProvider
}

// riverDispatchJobKind is the River job kind used for durable gala dispatch
const riverDispatchJobKind = "gala_dispatch_v1"

// DefaultQueueName is the default queue used for gala durable dispatch jobs
const DefaultQueueName = "events"

// riverDispatchArgs stores a JSON-encoded gala envelope for durable dispatch
type riverDispatchArgs struct {
	// Envelope is the encoded gala envelope payload
	Envelope []byte `json:"envelope"`
	// UniqueKey scopes the ByArgs uniqueness hash to this field alone via the river tag
	UniqueKey string `json:"unique_key,omitempty" river:"unique"`
}

// riverInsertClient represents the minimal insert capability required for durable dispatch
type riverInsertClient interface {
	// Insert inserts a River job with optional insert options
	Insert(context.Context, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// riverDispatcher dispatches envelopes to River
type riverDispatcher struct {
	// jobClient is the River client used to insert dispatch jobs
	jobClient riverInsertClient
	// defaultQueue is the default River queue for dispatch jobs
	defaultQueue string
}

// dispatcher dispatches envelopes to the configured transport
type dispatcher interface {
	// Dispatch dispatches an envelope to the configured transport
	Dispatch(context.Context, Envelope) error
}

// newRiverDispatcher creates a River-backed durable dispatcher
func newRiverDispatcher(jobClient riverInsertClient, defaultQueue string) (*riverDispatcher, error) {
	if jobClient == nil {
		return nil, ErrRiverJobClientRequired
	}

	if defaultQueue == "" {
		defaultQueue = DefaultQueueName
	}

	return &riverDispatcher{
		jobClient:    jobClient,
		defaultQueue: defaultQueue,
	}, nil
}

// newRiverDispatchArgs builds River dispatch args from an envelope
func newRiverDispatchArgs(envelope Envelope) (riverDispatchArgs, error) {
	encodedEnvelope, err := json.Marshal(envelope)
	if err != nil {
		return riverDispatchArgs{}, ErrRiverEnvelopeEncodeFailed
	}

	return riverDispatchArgs{
		Envelope:  encodedEnvelope,
		UniqueKey: envelope.Headers.UniqueKey,
	}, nil
}

// uniqueLiveStates are the job states a Headers.UniqueKey enforces uniqueness across: every
// live state, excluding terminal ones so a finished job never blocks a fresh insert. River
// requires available, pending, running, and scheduled in any custom state set
func uniqueLiveStates() []rivertype.JobState {
	return []rivertype.JobState{
		rivertype.JobStateAvailable,
		rivertype.JobStatePending,
		rivertype.JobStateRunning,
		rivertype.JobStateRetryable,
		rivertype.JobStateScheduled,
	}
}

// newRiverDispatchWorker creates a riverDispatchWorker
func newRiverDispatchWorker(provider galaProvider) *riverDispatchWorker {
	return &riverDispatchWorker{galaProvider: provider}
}

// riverJobMetadata is the JSON structure attached to River jobs for UI visibility
type riverJobMetadata struct {
	// Topic is the gala topic name
	Topic string `json:"topic"`
	// EventID is the gala event identifier
	EventID string `json:"event_id"`
	// Listeners are the registered listener names for the topic
	Listeners []string `json:"listeners,omitempty"`
	// Properties contains envelope header properties for UI visibility
	Properties map[string]string `json:"properties,omitempty"`
	// Metadata carries structured operation context for UI visibility
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// Dispatch dispatches an envelope to River for processing by a Worker
func (d *riverDispatcher) Dispatch(ctx context.Context, envelope Envelope) error {
	args, err := newRiverDispatchArgs(envelope)
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

	if envelope.Headers.ScheduledAt != nil {
		insertOpts.ScheduledAt = *envelope.Headers.ScheduledAt
	}

	if envelope.Headers.UniqueKey != "" {
		byState := uniqueLiveStates()
		if envelope.Headers.UniqueOnce {
			byState = append(byState, rivertype.JobStateCompleted, rivertype.JobStateCancelled, rivertype.JobStateDiscarded)
		}

		insertOpts.UniqueOpts = river.UniqueOpts{
			ByArgs:  true,
			ByState: byState,
		}
	}

	meta, err := json.Marshal(riverJobMetadata{
		Topic:      string(envelope.Topic),
		EventID:    string(envelope.ID),
		Listeners:  envelope.Headers.Listeners,
		Properties: envelope.Headers.Properties,
		Metadata:   envelope.Headers.Metadata,
	})
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("gala: error marshaling envelope")
		return ErrRiverEnvelopeEncodeFailed
	}

	insertOpts.Metadata = meta

	result, err := d.jobClient.Insert(ctx, args, insertOpts)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("gala: error inserting dispatch job")
		return ErrRiverDispatchInsertFailed
	}

	if result.UniqueSkippedAsDuplicate {
		logx.FromContext(ctx).Info().Str("topic", string(envelope.Topic)).Str("unique_key", envelope.Headers.UniqueKey).Msg("gala: dispatch skipped, a live job already holds the unique key")
	}

	return nil
}

// Kind satisfies river.JobArgs
func (riverDispatchArgs) Kind() string {
	return riverDispatchJobKind
}

// decodeEnvelope decodes the gala envelope from dispatch args
func (a riverDispatchArgs) decodeEnvelope() (Envelope, error) {
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
func (w *riverDispatchWorker) Work(ctx context.Context, job *river.Job[riverDispatchArgs]) error {
	if w.galaProvider == nil {
		return ErrRiverGalaProviderRequired
	}

	g := w.galaProvider()
	if g == nil {
		return ErrGalaRequired
	}

	envelope, err := job.Args.decodeEnvelope()
	if err != nil {
		return err
	}

	return g.dispatchEnvelope(context.WithoutCancel(ctx), envelope)
}
