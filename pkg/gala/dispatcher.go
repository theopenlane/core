package gala

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// galaProvider resolves the gala instance used by River workers
type galaProvider func() *Gala

// riverDispatchWorker processes legacy-kind durable gala dispatch jobs from River
type riverDispatchWorker struct {
	river.WorkerDefaults[EnvelopeArgs]

	galaProvider galaProvider
}

// riverDispatchJobKind is the legacy River job kind, retained as the registered fallback
// so pre-kind jobs and unkinded emissions always have a worker
const riverDispatchJobKind = "gala_dispatch_v1"

// DefaultQueueName is the default queue used for gala durable dispatch jobs
const DefaultQueueName = "events"

// DefaultJobTimeout is the default maximum run time for one dispatch job
const DefaultJobTimeout = 15 * time.Minute

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
	// kindQueues maps each registered job kind to its dedicated queue
	kindQueues map[string]string
}

// dispatcher dispatches envelopes to the configured transport
type dispatcher interface {
	// Dispatch dispatches an envelope to the configured transport
	Dispatch(context.Context, Envelope) error
}

// dispatchResult reports whether River inserted a new row or returned the row
// already holding the unique key
type dispatchResult struct {
	inserted bool
	holder   *rivertype.JobRow
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
func newRiverDispatchArgs(envelope Envelope) (EnvelopeArgs, error) {
	encodedEnvelope, err := json.Marshal(envelope)
	if err != nil {
		return EnvelopeArgs{}, ErrRiverEnvelopeEncodeFailed
	}

	return EnvelopeArgs{
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
	result, err := d.dispatch(ctx, envelope)
	if err != nil {
		return err
	}
	if !dispatchHolderReady(result) {
		event := logx.FromContext(ctx).Error().Str("topic", string(envelope.Topic))
		if result.holder != nil {
			event = event.Int64("holder_job_id", result.holder.ID).Str("holder_state", string(result.holder.State))
		}

		event.Msg("gala: dispatch has no runnable or completed holder")

		return ErrRiverDispatchInsertFailed
	}

	return nil
}

// dispatch inserts an envelope and returns the inserted row or duplicate holder.
func (d *riverDispatcher) dispatch(ctx context.Context, envelope Envelope) (dispatchResult, error) {
	envelopeArgs, err := newRiverDispatchArgs(envelope)
	if err != nil {
		return dispatchResult{}, err
	}

	kind := strings.TrimSpace(envelope.Headers.Kind)
	if _, registered := d.kindQueues[kind]; kind != "" && !registered {
		logx.FromContext(ctx).Warn().Str("kind", kind).Str("topic", string(envelope.Topic)).Msg("gala: unregistered job kind, dispatching under the default kind")

		kind = ""
	}

	args := kindedEnvelopeArgs{EnvelopeArgs: envelopeArgs, kind: kind}

	queueName := strings.TrimSpace(envelope.Headers.Queue)
	if queueName == "" {
		queueName = d.kindQueues[kind]
	}

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
		return dispatchResult{}, ErrRiverEnvelopeEncodeFailed
	}

	insertOpts.Metadata = meta

	result, err := d.jobClient.Insert(ctx, args, insertOpts)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("gala: error inserting dispatch job")
		return dispatchResult{}, ErrRiverDispatchInsertFailed
	}
	if result == nil || result.Job == nil {
		logx.FromContext(ctx).Error().Msg("gala: river insert returned no holder row")

		return dispatchResult{}, ErrRiverDispatchInsertFailed
	}

	if result.UniqueSkippedAsDuplicate {
		logx.FromContext(ctx).Info().Str("topic", string(envelope.Topic)).Str("unique_key", envelope.Headers.UniqueKey).Msg("gala: dispatch skipped, a live job already holds the unique key")
	}

	return dispatchResult{
		inserted: !result.UniqueSkippedAsDuplicate,
		holder:   result.Job,
	}, nil
}

// dispatchHolderReady reports whether a transport outcome points at a row that
// can still run or has already completed successfully.
func dispatchHolderReady(result dispatchResult) bool {
	if result.holder == nil || result.holder.ID == 0 {
		return false
	}

	switch result.holder.State {
	case rivertype.JobStateAvailable,
		rivertype.JobStatePending,
		rivertype.JobStateRetryable,
		rivertype.JobStateRunning,
		rivertype.JobStateScheduled,
		rivertype.JobStateCompleted:
		return true
	default:
		return false
	}
}

// decodeDispatchEnvelope decodes a gala envelope from encoded dispatch payload bytes
func decodeDispatchEnvelope(payload []byte) (Envelope, error) {
	var envelope Envelope
	if len(payload) == 0 {
		return envelope, ErrRiverDispatchJobEnvelopeRequired
	}

	if err := jsonx.RoundTrip(payload, &envelope); err != nil {
		return envelope, ErrRiverEnvelopeDecodeFailed
	}

	return envelope, nil
}

// workEnvelope decodes an encoded envelope and dispatches it to the topic's listeners
func workEnvelope(ctx context.Context, provider galaProvider, payload []byte) error {
	if provider == nil {
		return ErrRiverGalaProviderRequired
	}

	g := provider()
	if g == nil {
		return ErrGalaRequired
	}

	envelope, err := decodeDispatchEnvelope(payload)
	if err != nil {
		return err
	}

	return g.dispatchEnvelope(ctx, envelope)
}

// Work processes one River dispatch job and invokes Gala dispatch
func (w *riverDispatchWorker) Work(ctx context.Context, job *river.Job[EnvelopeArgs]) error {
	return workEnvelope(ctx, w.galaProvider, job.Args.EnvelopePayload())
}
