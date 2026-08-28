package gala

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
)

// DispatchMode controls whether envelopes are dispatched durably or in-memory.
type DispatchMode string

const (
	// DispatchModeDurable persists envelopes in River before worker execution.
	DispatchModeDurable DispatchMode = "durable"
	// DispatchModeInMemory dispatches envelopes immediately in-process.
	DispatchModeInMemory DispatchMode = "in_memory"
)

// Config configures cohesive Gala startup
type Config struct {
	// DispatchMode controls whether events are dispatched durably (River) or in-memory.
	DispatchMode DispatchMode
	// ConnectionURI is the database connection URI used for the dedicated gala river client
	ConnectionURI string
	// QueueName is the gala queue used for durable dispatch jobs
	QueueName string
	// WorkerCount is the max worker concurrency for the gala queue
	WorkerCount int
	// MaxRetries sets max attempts for gala dispatch jobs when greater than zero
	MaxRetries int
	// RunMigrations enables River schema migrations on startup (use for tests only)
	RunMigrations bool
	// FetchCooldown is the minimum time between job fetches per worker (default 100ms, min 1ms)
	// Lower values increase throughput but also database load. River enforces 1ms minimum.
	FetchCooldown time.Duration
	// FetchPollInterval is the fallback polling interval when LISTEN/NOTIFY misses events (default 1s)
	// This is only used when LISTEN/NOTIFY fails to deliver notifications.
	FetchPollInterval time.Duration
	// Kinds are the envelope kind namespaces this runtime registers alongside the default
	// dispatch kind; each kind gets a dedicated queue and shares the envelope worker
	Kinds []Namespace
	// TopicRenames maps retired topic names to their designated replacements, applied to
	// queued envelopes at dispatch and during job migration
	TopicRenames map[TopicName]TopicName
	// OperationRenames maps retired payload operation strings to their designated replacements
	OperationRenames map[string]string
	// JobTimeout is the maximum run time for one dispatch job; long-running batch
	// operations need far more than River's one-minute default
	JobTimeout time.Duration
	// SoftStopTimeout is how long in-flight jobs may keep running after a stop begins
	// before their contexts are cancelled
	SoftStopTimeout time.Duration
}

// Gala provides cohesive event dispatch + worker lifecycle management
// no black tie required, but a riverboat and some confetti wouldn't hurt
type Gala struct {
	// registry manages topic and listener registrations
	registry *registry
	// injector provides dependency resolution for listeners
	injector do.Injector
	// contextManager handles context capture and restoration
	contextManager *ContextManager
	// jobClient is the dedicated River client used for durable dispatch
	jobClient *riverqueue.Client
	// insertClient is the minimal insert capability used for durable dispatch
	insertClient riverInsertClient
	// jobController manages durable jobs after insertion
	jobController riverJobController
	// defaultQueue is the default River queue for dispatch jobs
	defaultQueue string
	// kindQueues maps each registered job kind to its dedicated queue
	kindQueues map[string]string
	// durableQueues tracks River queues this runtime is responsible for.
	durableQueues []string
	// dispatchMode captures the runtime dispatch mode.
	dispatchMode DispatchMode
	// topicRenames maps retired topic names to their designated replacements
	topicRenames map[TopicName]TopicName
	// operationRenames maps retired payload operation strings to their designated replacements
	operationRenames map[string]string
	// inMemoryPool backs in-process dispatch when DispatchModeInMemory is enabled.
	inMemoryPool *Pool
}

// NewGala initializes your gala, initializes dependencies, and starts workers
func NewGala(ctx context.Context, config Config) (app *Gala, err error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	if config.DispatchMode == DispatchModeInMemory {
		return newInMemoryGala(config)
	}

	app = &Gala{}

	kindNames := lo.Map(config.Kinds, func(kind Namespace, _ int) string {
		return kind.Kind()
	})

	// the alias registry must carry every kind before worker registration reads it
	registerEnvelopeKinds(kindNames)

	workers := river.NewWorkers()
	if err := river.AddWorkerSafely(workers, newRiverDispatchWorker(func() *Gala {
		return app
	})); err != nil {
		return nil, err
	}

	queues := map[string]river.QueueConfig{config.QueueName: {MaxWorkers: config.WorkerCount}}
	kindQueues := map[string]string{}

	for _, kind := range config.Kinds {
		// scope kind queues by the runtime's base queue so runtimes sharing a database
		// (test harnesses) don't steal each other's kind-routed jobs; pods sharing a queue
		// name still share workers
		queue := config.QueueName + "_" + kind.Queue()
		queues[queue] = river.QueueConfig{MaxWorkers: config.WorkerCount}
		kindQueues[kind.Kind()] = queue
	}

	riverConf := river.Config{
		Workers:         workers,
		Queues:          queues,
		JobTimeout:      config.JobTimeout,
		SoftStopTimeout: config.SoftStopTimeout,
	}

	if config.MaxRetries > 0 {
		riverConf.MaxAttempts = config.MaxRetries
	}

	if config.FetchCooldown > 0 {
		riverConf.FetchCooldown = config.FetchCooldown
	}

	if config.FetchPollInterval > 0 {
		riverConf.FetchPollInterval = config.FetchPollInterval
	}

	jobOpts := []riverqueue.Option{
		riverqueue.WithConnectionURI(config.ConnectionURI),
		riverqueue.WithRiverConfig(riverConf),
	}

	if config.RunMigrations {
		jobOpts = append(jobOpts, riverqueue.WithRunMigrations(true))
	}

	jobClient, err := riverqueue.New(ctx, jobOpts...)
	if err != nil {
		return nil, ErrRiverClientInitializationFailed
	}

	// close if anything fails; slightly awkward but cleaner than closing inside of all the error branches
	defer func() {
		if err != nil {
			_ = jobClient.Close()
		}
	}()

	if err := app.initialize(DispatchModeDurable); err != nil {
		return nil, err
	}

	app.jobClient = jobClient
	app.insertClient = jobClient
	app.jobController = jobClient.GetRiverClient()
	app.defaultQueue = config.QueueName
	app.kindQueues = kindQueues
	app.durableQueues = append([]string{config.QueueName}, lo.Values(kindQueues)...)
	app.topicRenames = config.TopicRenames
	app.operationRenames = config.OperationRenames

	return app, nil
}

// initialize sets Gala core dependencies and default runtime services
// future expansion / features may necessitate passing in additional dependencies or a more complex runtime config object but avoiding pre-optimization
func (g *Gala) initialize(dispatchMode DispatchMode) error {
	contextManager, err := newContextManager(
		NewKeyCodec("caller", auth.CallerKey),
		NewKeyCodec("workflow_flags", WorkflowFlagsKey),
		NewKeyCodec("integration_directory_sync_run_id", DirectorySyncRunIDKey),
		NewKeyCodec("active_trust_center_id", auth.ActiveTrustCenterIDKey),
		OperationContextCodec(),
		logFieldsCodec(),
	)
	if err != nil {
		return err
	}

	g.registry = newRegistry()
	g.injector = do.New()
	g.contextManager = contextManager
	g.dispatchMode = dispatchMode

	return nil
}

// validate normalizes config defaults and validates required fields
func (c *Config) validate() error {
	if c.DispatchMode == "" {
		c.DispatchMode = DispatchModeDurable
	}

	switch c.DispatchMode {
	case DispatchModeDurable, DispatchModeInMemory:
	default:
		return ErrDispatchModeInvalid
	}

	if c.QueueName == "" {
		c.QueueName = DefaultQueueName
	}

	if c.WorkerCount < 1 {
		c.WorkerCount = 1
	}

	if c.Kinds == nil {
		c.Kinds = JobKinds()
	}

	if c.JobTimeout <= 0 {
		c.JobTimeout = DefaultJobTimeout
	}

	if c.SoftStopTimeout <= 0 {
		c.SoftStopTimeout = DefaultSoftStopTimeout
	}

	if c.DispatchMode == DispatchModeDurable && c.ConnectionURI == "" {
		return ErrRiverConnectionURIRequired
	}

	return nil
}

// InterestedIn reports whether any registered listener matches the topic, operation, and
// soft-delete disposition
func (g *Gala) InterestedIn(topic TopicName, operation string) bool {
	return g.registry.InterestedIn(topic, operation)
}

// AttachOption provisions one dependency or durable context codec onto a gala runtime
type AttachOption func(*Gala) error

// WithValue provides a typed dependency that listeners resolve via HandlerContext.Injector
func WithValue[T any](value T) AttachOption {
	return func(g *Gala) error {
		do.ProvideValue(g.injector, value)

		return nil
	}
}

// ReplaceValue swaps a typed runtime dependency and returns an idempotent restore function
func ReplaceValue[T any](g *Gala, value T) (func(), error) {
	if g == nil {
		return nil, ErrGalaRequired
	}

	previous, err := do.Invoke[T](g.injector)
	if err != nil {
		return nil, err
	}

	do.OverrideValue(g.injector, value)

	var once sync.Once

	return func() {
		once.Do(func() {
			do.OverrideValue(g.injector, previous)
		})
	}, nil
}

// WithRestoredValue registers a durable context codec that re-resolves a live dependency
// from the runtime's injector on the handler side and attaches it to the restored context
func WithRestoredValue[T any](id ContextKey, setter func(context.Context, T) context.Context) AttachOption {
	return func(g *Gala) error {
		return g.contextManager.Register(newInjectorCodec(id, g.injector, setter))
	}
}

// Attach provisions dependencies and durable context codecs onto the runtime. It is
// called during wiring, after construction and before the first emit that relies on
// the provisioned values; registration failures are wiring errors
func (g *Gala) Attach(opts ...AttachOption) error {
	for _, opt := range opts {
		if err := opt(g); err != nil {
			return err
		}
	}

	return nil
}

// EmitOption customizes one emitted envelope before dispatch
type EmitOption func(*Envelope)

// WithEventID sets an explicit event identifier on the emitted envelope,
// making the caller's identity (e.g. a mutation event id or run id) the
// durable dedup and traceability key instead of a freshly minted ULID
func WithEventID(id EventID) EmitOption {
	return func(e *Envelope) {
		if id != "" {
			e.ID = id
		}
	}
}

// WithRawPayload emits pre-encoded payload bytes, bypassing the topic codec.
// The payload argument passed to Emit is ignored when set; the topic must
// still be registered so listeners can decode at dispatch time
func WithRawPayload(raw json.RawMessage) EmitOption {
	return func(e *Envelope) {
		if len(raw) > 0 {
			e.Payload = append(json.RawMessage(nil), raw...)
		}
	}
}

// EmitWithHeaders emits a payload to the topic under the given headers, applying any
// options to the envelope before dispatch, and returns the emitted event identifier;
// headers must carry a job kind
func (g *Gala) EmitWithHeaders(ctx context.Context, topic TopicName, payload any, headers Headers, opts ...EmitOption) (EventID, error) {
	registration, err := g.registry.topicRegistration(topic)
	if err != nil {
		return "", err
	}

	if headers.Kind == "" {
		headers.Kind = registration.kind
	}

	if headers.Kind == "" {
		return "", ErrJobKindRequired
	}

	envelope := Envelope{
		Headers:    headers,
		ID:         NewEventID(),
		Topic:      topic,
		OccurredAt: time.Now().UTC(),
	}

	for _, opt := range opts {
		opt(&envelope)
	}

	rawPayload := len(envelope.Payload) > 0

	if !rawPayload {
		encodedPayload, err := registration.encode(payload)
		if err != nil {
			return "", err
		}

		envelope.Payload = encodedPayload
	}

	if envelope.Headers.UniqueKey == "" && !envelope.Headers.SkipUniqueKey && registration.uniqueKey != nil {
		uniqueKey := registration.uniqueKey(payload)
		// decode raw emits so dedup keys match typed emits on the same topic
		if uniqueKey == "" && rawPayload {
			decodedPayload, err := registration.decode(envelope.Payload)
			if err != nil {
				return "", err
			}

			uniqueKey = registration.uniqueKey(decodedPayload)
		}

		envelope.Headers.UniqueKey = uniqueKey
	}

	snapshot, err := g.contextManager.Capture(ctx)
	if err != nil {
		return "", err
	}

	envelope.ContextSnapshot = snapshot

	envelope.Headers.Listeners = g.registry.listenerNamesForTopic(topic)

	if err := g.dispatch(ctx, envelope); err != nil {
		logx.FromContext(ctx).Debug().Err(err).Str("event_id", string(envelope.ID)).Str("topic", string(topic)).Msg("gala event dispatch failed")

		return envelope.ID, errors.Join(ErrDispatchFailed, err)
	}

	logx.FromContext(ctx).Debug().Str("event_id", string(envelope.ID)).Str("topic", string(topic)).Msg("gala event emitted")

	return envelope.ID, nil
}

// dispatch routes an envelope to the transport for the configured dispatch mode
func (g *Gala) dispatch(ctx context.Context, envelope Envelope) error {
	switch g.dispatchMode {
	case DispatchModeDurable:
		return g.dispatchDurable(ctx, envelope)
	case DispatchModeInMemory:
		return g.dispatchInMemory(ctx, envelope)
	default:
		return ErrDispatchModeInvalid
	}
}

// dispatchEnvelope dispatches one envelope to all listeners on the topic
func (g *Gala) dispatchEnvelope(ctx context.Context, envelope Envelope) error {
	registration, err := g.registry.topicRegistration(envelope.Topic)
	if err != nil {
		renamed, renameOK := g.topicRenames[envelope.Topic]
		if !renameOK {
			return err
		}

		envelope.Topic = renamed

		if registration, err = g.registry.topicRegistration(renamed); err != nil {
			return err
		}
	}

	decodedPayload, err := registration.decode(envelope.Payload)
	if err != nil {
		return err
	}

	restoredContext, err := g.contextManager.Restore(ctx, envelope.ContextSnapshot)
	if err != nil {
		return err
	}

	operation := payloadOperation(decodedPayload)
	if renamed, ok := g.operationRenames[operation]; ok {
		operation = renamed
		decodedPayload = payloadWithOperation(decodedPayload, renamed)
	}

	logx.FromContext(restoredContext).Debug().Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("operation", operation).Msg("gala processing event")

	handlerContext := HandlerContext{
		Context:  restoredContext,
		Envelope: envelope,
		Injector: g.injector,
	}

	listeners := g.registry.registeredListeners(envelope.Topic)

	for _, listener := range listeners {
		if !listenerMatches(listener, operation) {
			continue
		}

		if err := g.executeListener(handlerContext, listener, decodedPayload, operation); err != nil {
			// header properties are the only identity left when the emitting context carried no durable log fields
			logx.FromContext(restoredContext).Warn().Err(err).Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("operation", operation).Str("listener", listener.name).Interface("envelope_properties", envelope.Headers.Properties).Msg("gala listener failed")

			return err
		}
	}

	logx.FromContext(restoredContext).Info().Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("operation", operation).Int("listener_count", len(listeners)).Msg("gala event processed")

	return nil
}

// PayloadOperation is the payload contract for listener operation routing; payloads
// without it dispatch with an empty operation
type PayloadOperation interface {
	PayloadOperation() string
}

// PayloadOperationRenamer is an optional payload contract for operation renames, returning
// a copy of the payload carrying the renamed operation
type PayloadOperationRenamer interface {
	WithPayloadOperation(operation string) any
}

// payloadWithOperation returns the payload with its operation replaced when supported
func payloadWithOperation(payload any, operation string) any {
	if renamer, ok := payload.(PayloadOperationRenamer); ok {
		return renamer.WithPayloadOperation(operation)
	}

	return payload
}

// payloadOperation extracts the payload's operation when the payload declares one
func payloadOperation(payload any) string {
	if operationPayload, ok := payload.(PayloadOperation); ok {
		return strings.TrimSpace(operationPayload.PayloadOperation())
	}

	return ""
}

// executeListener executes a single listener with panic recovery, recording delivery metrics
func (g *Gala) executeListener(handlerContext HandlerContext, listener registeredListener, payload any, operation string) (err error) {
	start := time.Now()
	labels := prometheus.Labels{
		metricLabelTopic:     string(handlerContext.Envelope.Topic),
		metricLabelListener:  listener.definitionName,
		metricLabelOperation: operation,
	}

	skipped := false

	defer func() {
		if recovered := recover(); recovered != nil {
			err = ListenerError{
				ListenerName: listener.name,
				Cause:        fmt.Errorf("%w: %v", ErrListenerPanicked, recovered),
				Panicked:     true,
			}
		}

		if skipped {
			return
		}

		listenerDeliveries.With(labels).Inc()
		listenerDuration.With(labels).Observe(time.Since(start).Seconds())

		if err != nil {
			listenerFailures.With(labels).Inc()
		}
	}()

	switch listenerErr := listener.handle(handlerContext, payload, operation); {
	case errors.Is(listenerErr, ErrListenerGated):
		skipped = true

		return nil
	case listenerErr != nil:
		return ListenerError{
			ListenerName: listener.name,
			Cause:        listenerErr,
			Panicked:     false,
		}
	}

	return nil
}

// StartWorkers starts Gala workers
func (g *Gala) StartWorkers(ctx context.Context) error {
	if g.dispatchMode == DispatchModeInMemory {
		return nil
	}

	if err := g.jobClient.GetRiverClient().Start(ctx); err != nil {
		return ErrRiverWorkerStartFailed
	}

	return nil
}

// StopWorkers stops Gala workers
func (g *Gala) StopWorkers(ctx context.Context) error {
	if g.dispatchMode == DispatchModeInMemory {
		return nil
	}

	if err := g.jobClient.GetRiverClient().Stop(ctx); err != nil &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, context.DeadlineExceeded) {
		return ErrRiverWorkerStopFailed
	}

	return nil
}

// WaitIdle blocks until runnable work drains or the context ends
// Future scheduled and retry work does not count as busy in durable mode
func (g *Gala) WaitIdle(ctx context.Context) error {
	switch g.dispatchMode {
	case DispatchModeInMemory:
		if g.inMemoryPool != nil {
			return g.inMemoryPool.WaitIdle(ctx)
		}
	case DispatchModeDurable:
		return g.waitDurableIdle(ctx)
	default:
		return ErrDispatchModeInvalid
	}

	return nil
}

// waitDurableIdle polls configured Gala queues until no runnable jobs remain
func (g *Gala) waitDurableIdle(ctx context.Context) error {
	if g.jobController == nil {
		return nil
	}

	return waitStable(ctx, func() (bool, error) {
		params := river.NewJobListParams().
			States(rivertype.JobStateAvailable, rivertype.JobStatePending, rivertype.JobStateRunning).
			First(1)
		if len(g.durableQueues) > 0 {
			params = params.Queues(g.durableQueues...)
		}

		result, err := g.jobController.JobList(ctx, params)
		if err != nil {
			return false, err
		}
		for _, job := range result.Jobs {
			switch job.State {
			case rivertype.JobStateAvailable, rivertype.JobStatePending, rivertype.JobStateRunning:
				return false, nil
			default:
				continue
			}
		}

		now := time.Now().UTC()
		retryParams := river.NewJobListParams().
			States(rivertype.JobStateRetryable).
			Where("scheduled_at <= @now", river.NamedArgs{"now": now}).
			First(1)
		if len(g.durableQueues) > 0 {
			retryParams = retryParams.Queues(g.durableQueues...)
		}

		retries, err := g.jobController.JobList(ctx, retryParams)
		if err != nil {
			return false, err
		}

		for _, job := range retries.Jobs {
			if job.State == rivertype.JobStateRetryable && !job.ScheduledAt.After(now) {
				return false, nil
			}
		}

		return true, nil
	}, durableWaitPollInterval, durableIdleThreshold)
}

const (
	durableWaitPollInterval       = 50 * time.Millisecond
	durableJobCleanupPollInterval = 10 * time.Millisecond
	durableIdleThreshold          = 3
	activeJobsPageSize            = 100
)

// activeJobListParams builds JobList params for matching jobs in an "active" state
func activeJobListParams(metadataFragment string) *river.JobListParams {
	return river.NewJobListParams().Metadata(metadataFragment).States(uniqueLiveStates()...)
}

// HasActiveJobWithMetadata reports whether at least one active-state River job matches the
// metadata fragment; false without error when Gala is not durable
func (g *Gala) HasActiveJobWithMetadata(ctx context.Context, metadataFragment string) (bool, error) {
	if g.jobController == nil {
		return false, nil
	}

	params := activeJobListParams(metadataFragment).First(1)
	if len(g.durableQueues) > 0 {
		params = params.Queues(g.durableQueues...)
	}

	result, err := g.jobController.JobList(ctx, params)
	if err != nil {
		return false, err
	}

	return len(result.Jobs) > 0, nil
}

// activeJobsWithMetadata lists every active-state River job matching the metadata fragment
func (g *Gala) activeJobsWithMetadata(ctx context.Context, metadataFragment string) ([]*rivertype.JobRow, error) {
	params := activeJobListParams(metadataFragment).First(activeJobsPageSize)
	if len(g.durableQueues) > 0 {
		params = params.Queues(g.durableQueues...)
	}

	var jobs []*rivertype.JobRow

	for {
		result, err := g.jobController.JobList(ctx, params)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, result.Jobs...)

		if len(result.Jobs) < activeJobsPageSize || result.LastCursor == nil {
			return jobs, nil
		}

		params = params.After(result.LastCursor)
	}
}

// CountActiveJobsWithMetadata returns how many River jobs whose metadata JSONB contains the
// given fragment are in an active state. Returns zero without error when Gala is not in durable mode
func (g *Gala) CountActiveJobsWithMetadata(ctx context.Context, metadataFragment string) (int, error) {
	if g.jobController == nil {
		return 0, nil
	}

	jobs, err := g.activeJobsWithMetadata(ctx, metadataFragment)
	if err != nil {
		return 0, err
	}

	return len(jobs), nil
}

// PurgeActiveJobsWithMetadata deletes live River jobs matching a JSONB metadata fragment
// Running jobs are cancelled before deletion and rescanned for recurring successors
func (g *Gala) PurgeActiveJobsWithMetadata(ctx context.Context, metadataFragment string) (int, error) {
	return g.purgeActiveJobsWithMetadata(ctx, metadataFragment, nil)
}

type activeJobPredicate func(*rivertype.JobRow) (bool, error)

func (g *Gala) purgeActiveJobsWithMetadata(ctx context.Context, metadataFragment string, predicate activeJobPredicate) (int, error) {
	if g.jobController == nil {
		return 0, nil
	}

	pendingDeletes := map[int64]bool{}
	targeted := map[int64]struct{}{}

	for {
		jobs, err := g.activeJobsWithMetadata(ctx, metadataFragment)
		if err != nil {
			return len(targeted), err
		}

		found := false

		for _, job := range jobs {
			if predicate != nil {
				matched, err := predicate(job)
				if err != nil {
					return len(targeted), err
				}

				if !matched {
					continue
				}
			}

			found = true
			targeted[job.ID] = struct{}{}

			if _, tracked := pendingDeletes[job.ID]; !tracked {
				pendingDeletes[job.ID] = false
			}
		}

		for jobID, cancelled := range pendingDeletes {
			if !cancelled {
				if _, err := g.jobController.JobCancel(ctx, jobID); err != nil {
					if errors.Is(err, rivertype.ErrNotFound) {
						delete(pendingDeletes, jobID)

						continue
					}

					return len(targeted), err
				}

				pendingDeletes[jobID] = true
			}

			if _, err := g.jobController.JobDelete(ctx, jobID); err != nil {
				switch {
				case errors.Is(err, rivertype.ErrNotFound):
					delete(pendingDeletes, jobID)
				case errors.Is(err, rivertype.ErrJobRunning):
				default:
					return len(targeted), err
				}

				continue
			}

			delete(pendingDeletes, jobID)
		}

		if len(pendingDeletes) == 0 {
			if !found {
				return len(targeted), nil
			}

			// Rescan for successors emitted by a concurrently finishing handler
			continue
		}

		timer := time.NewTimer(durableJobCleanupPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()

			return len(targeted), ctx.Err()
		case <-timer.C:
		}
	}
}

// Close closes the dedicated Gala queue client
func (g *Gala) Close() error {
	if g.inMemoryPool != nil {
		g.inMemoryPool.Release()
	}

	if g.jobClient == nil {
		return nil
	}

	if err := g.jobClient.Close(); err != nil {
		return ErrRiverClientCloseFailed
	}

	return nil
}
