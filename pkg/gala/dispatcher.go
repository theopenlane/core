package gala

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/jsonx"
)

// galaProvider resolves the gala instance used by River workers
type galaProvider func() *Gala

// riverDispatchWorker processes durable gala dispatch jobs from River
type riverDispatchWorker struct {
	river.WorkerDefaults[EnvelopeArgs]

	galaProvider galaProvider
}

// envelopeWorkerKind is the worker's registered kind; every namespaced kind is an alias of it and
// no emission is ever enqueued under it
const envelopeWorkerKind = "gala_dispatch_v1"

// DefaultQueueName is the default queue used for gala durable dispatch jobs
const DefaultQueueName = "events"

// DefaultJobTimeout is the default maximum run time for one dispatch job
const DefaultJobTimeout = 15 * time.Minute

// DefaultSoftStopTimeout is the default window in-flight jobs get to finish once a stop begins
const DefaultSoftStopTimeout = 8 * time.Second

// riverInsertClient represents the minimal insert capability required for durable dispatch
type riverInsertClient interface {
	// Insert inserts a River job with optional insert options
	Insert(context.Context, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// riverJobController manages durable jobs after insertion
type riverJobController interface {
	JobList(context.Context, *river.JobListParams) (*river.JobListResult, error)
	JobCancel(context.Context, int64) (*rivertype.JobRow, error)
	JobDelete(context.Context, int64) (*rivertype.JobRow, error)
	JobRetry(context.Context, int64) (*rivertype.JobRow, error)
	QueuePause(context.Context, string, *river.QueuePauseOpts) error
	QueueResume(context.Context, string, *river.QueuePauseOpts) error
}

// dispatchResult reports whether River inserted a new row or returned the row
// already holding the unique key
type dispatchResult struct {
	inserted bool
	holder   *rivertype.JobRow
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

// dispatchDurable dispatches an envelope to River for processing by a Worker
func (g *Gala) dispatchDurable(ctx context.Context, envelope Envelope) error {
	result, err := g.insertEnvelope(ctx, envelope)
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

// insertEnvelope inserts an envelope and returns the inserted row or duplicate holder
func (g *Gala) insertEnvelope(ctx context.Context, envelope Envelope) (dispatchResult, error) {
	envelopeArgs, err := newRiverDispatchArgs(envelope)
	if err != nil {
		return dispatchResult{}, err
	}

	kind := strings.TrimSpace(envelope.Headers.Kind)

	queueName, registered := g.kindQueues[kind]
	if !registered {
		logx.FromContext(ctx).Error().Str("kind", kind).Str("topic", string(envelope.Topic)).Msg("gala: job kind is not registered, refusing dispatch")

		return dispatchResult{}, ErrJobKindNotRegistered
	}

	args := kindedEnvelopeArgs{EnvelopeArgs: envelopeArgs, kind: kind}

	insertOpts := &river.InsertOpts{
		Queue: queueName,
		Tags:  envelope.Headers.Tags,
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

	result, err := g.insertClient.Insert(ctx, args, insertOpts)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("gala: error inserting dispatch job")

		return dispatchResult{}, ErrRiverDispatchInsertFailed
	}

	if result == nil || result.Job == nil {
		logx.FromContext(ctx).Error().Msg("gala: river insert returned no holder row")

		return dispatchResult{}, ErrRiverDispatchInsertFailed
	}

	if result.UniqueSkippedAsDuplicate {
		logx.FromContext(ctx).Info().
			Str("topic", string(envelope.Topic)).
			Str("unique_key", envelope.Headers.UniqueKey).
			Bool("unique_once", envelope.Headers.UniqueOnce).
			Msg("gala: dispatch skipped, an existing job holds the unique key; with unique_once the holder may be a completed job")
	}

	return dispatchResult{
		inserted: !result.UniqueSkippedAsDuplicate,
		holder:   result.Job,
	}, nil
}

// dispatchHolderReady reports whether a transport outcome points at a row that
// can still run or has already completed successfully
func dispatchHolderReady(result dispatchResult) bool {
	if result.holder == nil || result.holder.ID == 0 {
		return false
	}

	return slices.Contains(append(uniqueLiveStates(), rivertype.JobStateCompleted), result.holder.State)
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

// Work processes one River dispatch job and invokes Gala dispatch
func (w *riverDispatchWorker) Work(ctx context.Context, job *river.Job[EnvelopeArgs]) error {
	envelope, err := decodeDispatchEnvelope(job.Args.EnvelopePayload())
	if err != nil {
		return err
	}

	runtime := w.galaProvider()
	if err := runtime.dispatchEnvelope(ctx, envelope); err != nil {
		return err
	}

	if runtime.completeJobsInline {
		runtime.completeJobInline(ctx, job)
	}

	return nil
}

// completeJobInline marks a finished job completed now instead of leaving it running until
// River's next completer flush, a failure leaves the job for that flush
func (g *Gala) completeJobInline(ctx context.Context, job *river.Job[EnvelopeArgs]) {
	tx, err := g.jobClient.GetPool().Begin(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Int64("job_id", job.ID).Msg("failed to begin inline job completion")

		return
	}

	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Int64("job_id", job.ID).Msg("failed to complete job inline")

		return
	}

	if err := tx.Commit(ctx); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Int64("job_id", job.ID).Msg("failed to commit inline job completion")
	}
}
