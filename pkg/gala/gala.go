package gala

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/samber/do/v2"
	"github.com/theopenlane/core/pkg/logx"
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
}

// Gala provides cohesive event dispatch + worker lifecycle management
// no black tie required, but a riverboat and some confetti wouldn't hurt
type Gala struct {
	// registry manages topic and listener registrations
	registry *registry
	// injector provides dependency resolution for listeners
	injector do.Injector
	// dispatcher handles envelope dispatch to listeners
	dispatcher dispatcher
	// contextManager handles context capture and restoration
	contextManager *ContextManager
	// jobClient is the dedicated River client used for durable dispatch
	jobClient *riverqueue.Client
	// durableQueues tracks River queues this runtime is responsible for.
	durableQueues []string
	// dispatchMode captures the runtime dispatch mode.
	dispatchMode DispatchMode
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

	workers := river.NewWorkers()
	if err := river.AddWorkerSafely(workers, newRiverDispatchWorker(func() *Gala {
		return app
	})); err != nil {
		return nil, err
	}

	riverConf := river.Config{
		Workers: workers,
		Queues:  map[string]river.QueueConfig{config.QueueName: {MaxWorkers: config.WorkerCount}},
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

	riverDispatch, err := newRiverDispatcher(jobClient, config.QueueName)
	if err != nil {
		return nil, err
	}

	if err := app.initialize(riverDispatch, DispatchModeDurable); err != nil {
		return nil, err
	}

	app.jobClient = jobClient
	app.durableQueues = []string{config.QueueName}

	return app, nil
}

// initialize sets Gala core dependencies and default runtime services
// future expansion / features may necessitate passing in additional dependencies or a more complex runtime config object but avoiding pre-optimization
func (g *Gala) initialize(d dispatcher, dispatchMode DispatchMode) error {
	contextManager, err := newContextManager(
		NewKeyCodec("caller", auth.CallerKey),
		logFieldsCodec{},
	)
	if err != nil {
		return err
	}

	g.registry = newRegistry()
	g.injector = do.New()
	g.dispatcher = d
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

	if c.DispatchMode == DispatchModeDurable && c.ConnectionURI == "" {
		return ErrRiverConnectionURIRequired
	}

	return nil
}

// InterestedIn reports whether any registered listener matches the topic and operation
func (g *Gala) InterestedIn(topic TopicName, operation string) bool {
	return g.registry.InterestedIn(topic, operation)
}

// Injector returns the Gala dependency injector
func (g *Gala) Injector() do.Injector {
	return g.injector
}

// ContextManager returns the Gala context manager
func (g *Gala) ContextManager() *ContextManager {
	return g.contextManager
}

// EmitOption customizes one emitted envelope before dispatch
type EmitOption func(*Envelope)

// WithHeaders sets the operational headers on the emitted envelope
func WithHeaders(headers Headers) EmitOption {
	return func(e *Envelope) {
		e.Headers = headers
	}
}

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
// The payload argument passed to EmitWithHeaders is ignored when set; the
// topic must still be registered so listeners can decode at dispatch time
func WithRawPayload(raw json.RawMessage) EmitOption {
	return func(e *Envelope) {
		if len(raw) > 0 {
			e.Payload = append(json.RawMessage(nil), raw...)
		}
	}
}

// Emit emits a payload to the topic, applying any options to the envelope before
// dispatch, and returns the emitted event identifier
func (g *Gala) Emit(ctx context.Context, topic TopicName, payload any, opts ...EmitOption) (EventID, error) {
	receipt := g.EmitWithHeaders(ctx, topic, payload, Headers{}, opts...)

	return receipt.EventID, receipt.Err
}

// EmitWithHeaders emits a payload with explicit headers
func (g *Gala) EmitWithHeaders(ctx context.Context, topic TopicName, payload any, headers Headers, opts ...EmitOption) EmitReceipt {
	registration, err := g.registry.topicRegistration(topic)
	if err != nil {
		return EmitReceipt{Err: err}
	}

	envelope := Envelope{
		ID:         NewEventID(),
		Topic:      topic,
		OccurredAt: time.Now().UTC(),
		Headers:    headers,
	}

	for _, opt := range opts {
		opt(&envelope)
	}

	if len(envelope.Payload) == 0 {
		encodedPayload, err := registration.encode(payload)
		if err != nil {
			return EmitReceipt{Err: err}
		}

		envelope.Payload = encodedPayload
	}

	if envelope.Headers.UniqueKey == "" && !envelope.Headers.SkipUniqueKey && registration.uniqueKey != nil {
		envelope.Headers.UniqueKey = registration.uniqueKey(payload)
	}

	snapshot, err := g.contextManager.Capture(ctx)
	if err != nil {
		return EmitReceipt{Err: err}
	}

	envelope.ContextSnapshot = snapshot

	if g.dispatcher == nil {
		return EmitReceipt{EventID: envelope.ID, Err: ErrDispatcherRequired}
	}

	envelope.Headers.Listeners = g.registry.listenerNamesForTopic(topic)

	if err := g.dispatcher.Dispatch(ctx, envelope); err != nil {
		logx.FromContext(ctx).Debug().Err(err).Str("event_id", string(envelope.ID)).Str("topic", string(topic)).Msg("gala event dispatch failed")

		return EmitReceipt{EventID: envelope.ID, Err: errors.Join(ErrDispatchFailed, err)}
	}

	logx.FromContext(ctx).Debug().Str("event_id", string(envelope.ID)).Str("topic", string(topic)).Msg("gala event emitted")

	return EmitReceipt{EventID: envelope.ID}
}

// dispatchEnvelope dispatches one envelope to all listeners on the topic
func (g *Gala) dispatchEnvelope(ctx context.Context, envelope Envelope) error {
	registration, err := g.registry.topicRegistration(envelope.Topic)
	if err != nil {
		return err
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

	logx.FromContext(restoredContext).Debug().Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("operation", operation).Msg("gala processing event")

	handlerContext := HandlerContext{
		Context:  restoredContext,
		Envelope: envelope,
		Injector: g.injector,
	}

	listeners := g.registry.registeredListeners(envelope.Topic)

	for _, listener := range listeners {
		if !listenerInterestedInOperation(listener, operation) {
			continue
		}

		if err := g.executeListener(handlerContext, listener, decodedPayload); err != nil {
			// header properties are the only identity left when the emitting context carried no durable log fields
			logx.FromContext(restoredContext).Warn().Err(err).Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("operation", operation).Str("listener", listener.name).Interface("envelope_properties", envelope.Headers.Properties).Msg("gala listener failed")

			return err
		}
	}

	logx.FromContext(restoredContext).Info().Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("operation", operation).Int("listener_count", len(listeners)).Msg("gala event processed")

	return nil
}

// payloadOperation extracts a mutation-style operation string when present
func payloadOperation(payload any) string {
	value := reflect.ValueOf(payload)
	if !value.IsValid() {
		return ""
	}

	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}

		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return ""
	}

	field := value.FieldByName("Operation")
	if !field.IsValid() || field.Kind() != reflect.String {
		return ""
	}

	return strings.TrimSpace(field.String())
}

// executeListener executes a single listener with panic recovery
func (g *Gala) executeListener(handlerContext HandlerContext, listener registeredListener, payload any) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = ListenerError{
				ListenerName: listener.name,
				Cause:        fmt.Errorf("%w: %v", ErrListenerPanicked, recovered),
				Panicked:     true,
			}
		}
	}()

	if listenerErr := listener.handle(handlerContext, payload); listenerErr != nil {
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

	if g.jobClient == nil {
		return ErrRiverJobClientRequired
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

	if g.jobClient == nil {
		return ErrRiverJobClientRequired
	}

	if err := g.jobClient.GetRiverClient().Stop(ctx); err != nil &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, context.DeadlineExceeded) {
		return ErrRiverWorkerStopFailed
	}

	return nil
}

// WaitIdle blocks until all dispatched work has completed.
// For in-memory mode, it waits for the pool to drain.
// For durable mode, it polls River for pending/running jobs.
func (g *Gala) WaitIdle() {
	switch g.dispatchMode {
	case DispatchModeInMemory:
		if g.inMemoryPool != nil {
			g.inMemoryPool.WaitIdle()
		}
	case DispatchModeDurable:
		g.waitDurableIdle()
	}
}

// waitDurableIdle polls configured Gala queues until no active jobs remain.
func (g *Gala) waitDurableIdle() {
	if g.jobClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), durableWaitTimeout)
	defer cancel()

	rc := g.jobClient.GetRiverClient()

	idleCount := 0
	for idleCount < durableIdleThreshold {
		if ctx.Err() != nil {
			return
		}

		params := river.NewJobListParams().
			States(rivertype.JobStateAvailable, rivertype.JobStateRunning, rivertype.JobStateScheduled).
			First(1)
		if len(g.durableQueues) > 0 {
			params = params.Queues(g.durableQueues...)
		}

		result, err := rc.JobList(ctx, params)
		if err != nil || len(result.Jobs) > 0 {
			idleCount = 0
		} else {
			idleCount++
		}

		time.Sleep(durableWaitPollInterval)
	}
}

const (
	durableWaitTimeout      = 30 * time.Second
	durableWaitPollInterval = 50 * time.Millisecond
	durableIdleThreshold    = 3
)

// HasActiveJobWithMetadata reports whether at least one River job whose metadata
// JSONB contains the given fragment exists in an active state (available, scheduled,
// running, or retryable). Returns false without error when Gala is not in durable mode
func (g *Gala) HasActiveJobWithMetadata(ctx context.Context, metadataFragment string) (bool, error) {
	if g.jobClient == nil {
		return false, nil
	}

	params := river.NewJobListParams().
		Metadata(metadataFragment).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateScheduled,
			rivertype.JobStateRetryable,
		).
		First(1)

	result, err := g.jobClient.GetRiverClient().JobList(ctx, params)
	if err != nil {
		return false, err
	}

	return len(result.Jobs) > 0, nil
}

// activeJobsWithMetadata lists every River job whose metadata JSONB contains the given fragment
// and is in an active state (available, scheduled, running, or retryable)
func (g *Gala) activeJobsWithMetadata(ctx context.Context, metadataFragment string) ([]*rivertype.JobRow, error) {
	result, err := g.jobClient.GetRiverClient().JobList(ctx, river.NewJobListParams().
		Metadata(metadataFragment).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateScheduled,
			rivertype.JobStateRetryable,
		))
	if err != nil {
		return nil, err
	}

	return result.Jobs, nil
}

// CountActiveJobsWithMetadata returns how many River jobs whose metadata JSONB contains the
// given fragment are in an active state. Returns zero without error when Gala is not in durable mode
func (g *Gala) CountActiveJobsWithMetadata(ctx context.Context, metadataFragment string) (int, error) {
	if g.jobClient == nil {
		return 0, nil
	}

	jobs, err := g.activeJobsWithMetadata(ctx, metadataFragment)
	if err != nil {
		return 0, err
	}

	return len(jobs), nil
}

// CancelActiveJobsWithMetadata cancels every River job whose metadata JSONB contains the given
// fragment and is in an active state, returning how many were cancelled. Returns zero without
// error when Gala is not in durable mode
func (g *Gala) CancelActiveJobsWithMetadata(ctx context.Context, metadataFragment string) (int, error) {
	if g.jobClient == nil {
		return 0, nil
	}

	jobs, err := g.activeJobsWithMetadata(ctx, metadataFragment)
	if err != nil {
		return 0, err
	}

	client := g.jobClient.GetRiverClient()

	var cancelled int

	for _, job := range jobs {
		if _, err := client.JobCancel(ctx, job.ID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Int64("job_id", job.ID).Msg("gala: failed cancelling job")

			continue
		}

		cancelled++
	}

	return cancelled, nil
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
