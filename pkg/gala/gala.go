package gala

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/samber/do/v2"
	"github.com/theopenlane/core/pkg/logx"
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
	// Enabled toggles Gala worker startup and dispatch support when true
	Enabled bool
	// ConnectionURI is the database connection URI used for the dedicated gala river client
	ConnectionURI string
	// QueueName is the gala queue used for durable dispatch jobs
	QueueName string
	// WorkerCount is the max worker concurrency for the gala queue
	WorkerCount int
	// QueueWorkers configures additional queue worker concurrency by queue name.
	QueueWorkers map[string]int
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
	registry *Registry
	// injector provides dependency resolution for listeners
	injector do.Injector
	// dispatcher handles envelope dispatch to listeners
	dispatcher Dispatcher
	// contextManager handles context capture and restoration
	contextManager *ContextManager
	// jobClient is the dedicated River client used for durable dispatch
	jobClient *riverqueue.Client
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
	if err := river.AddWorkerSafely(workers, NewRiverDispatchWorker(func() *Gala {
		return app
	})); err != nil {
		return nil, err
	}

	riverConf := river.Config{
		Workers: workers,
		Queues:  buildQueueConfig(config.QueueName, config.WorkerCount, config.QueueWorkers),
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

	dispatcher, err := NewRiverDispatcher(jobClient, config.QueueName)
	if err != nil {
		return nil, err
	}

	if err := app.initialize(dispatcher, DispatchModeDurable); err != nil {
		return nil, err
	}

	app.jobClient = jobClient

	return app, nil
}

// initialize sets Gala core dependencies and default runtime services
// future expansion / features may necessitate passing in additional dependencies or a more complex runtime config object but avoiding pre-optimization
func (g *Gala) initialize(dispatcher Dispatcher, dispatchMode DispatchMode) error {
	contextManager, err := NewContextManager(NewContextCodec())
	if err != nil {
		return err
	}

	g.registry = NewRegistry()
	g.injector = do.New()
	g.dispatcher = dispatcher
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

// buildQueueConfig constructs the River queue config with defaults and optional overrides
func buildQueueConfig(defaultQueueName string, defaultWorkerCount int, queueWorkers map[string]int) map[string]river.QueueConfig {
	queues := map[string]river.QueueConfig{
		defaultQueueName: {MaxWorkers: defaultWorkerCount},
	}

	for queueName, workers := range queueWorkers {
		queueName = strings.TrimSpace(queueName)
		if queueName == "" || workers < 1 {
			continue
		}

		queues[queueName] = river.QueueConfig{MaxWorkers: workers}
	}

	return queues
}

// Registry returns the Gala topic/listener registry
func (g *Gala) Registry() *Registry {
	return g.registry
}

// Injector returns the Gala dependency injector
func (g *Gala) Injector() do.Injector {
	return g.injector
}

// ContextManager returns the Gala context manager
func (g *Gala) ContextManager() *ContextManager {
	return g.contextManager
}

// EmitWithHeaders emits a payload with explicit headers
func (g *Gala) EmitWithHeaders(ctx context.Context, topic TopicName, payload any, headers Headers) EmitReceipt {
	registration, err := g.registry.topicRegistration(topic)
	if err != nil {
		return EmitReceipt{Err: err}
	}

	encodedPayload, err := registration.encode(payload)
	if err != nil {
		return EmitReceipt{Err: err}
	}

	snapshot, err := g.contextManager.Capture(ctx)
	if err != nil {
		return EmitReceipt{Err: err}
	}

	envelope := Envelope{
		ID:              NewEventID(),
		Topic:           topic,
		OccurredAt:      time.Now().UTC(),
		Headers:         headers,
		Payload:         encodedPayload,
		ContextSnapshot: snapshot,
	}

	if g.dispatcher == nil {
		return EmitReceipt{EventID: envelope.ID, Err: ErrDispatcherRequired}
	}

	if err := g.dispatcher.Dispatch(ctx, envelope); err != nil {
		logx.FromContext(ctx).Debug().Err(err).Str("event_id", string(envelope.ID)).Str("topic", string(topic)).Msg("gala event dispatch failed")

		return EmitReceipt{EventID: envelope.ID, Err: errors.Join(ErrDispatchFailed, err)}
	}

	logx.FromContext(ctx).Debug().Str("event_id", string(envelope.ID)).Str("topic", string(topic)).Msg("gala event emitted")

	return EmitReceipt{EventID: envelope.ID, Accepted: true}
}

// EmitEnvelope dispatches a pre-built envelope using its topic registration.
// When the envelope does not already carry a ContextSnapshot, one is captured
// from ctx so that durable dispatch can reconstruct auth and other values.
func (g *Gala) EmitEnvelope(ctx context.Context, envelope Envelope) error {
	if _, err := g.registry.topicRegistration(envelope.Topic); err != nil {
		return err
	}

	if g.dispatcher == nil {
		return ErrDispatcherRequired
	}

	if len(envelope.ContextSnapshot.Values) == 0 && len(envelope.ContextSnapshot.Flags) == 0 {
		snapshot, err := g.contextManager.Capture(ctx)
		if err != nil {
			return err
		}

		envelope.ContextSnapshot = snapshot
	}

	return g.dispatcher.Dispatch(ctx, envelope)
}

// DispatchEnvelope dispatches one envelope to all listeners on the topic
func (g *Gala) DispatchEnvelope(ctx context.Context, envelope Envelope) error {
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
			logx.FromContext(restoredContext).Warn().Err(err).Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("operation", operation).Str("listener", listener.name).Msg("gala listener failed")

			return err
		}
	}

	logx.FromContext(restoredContext).Debug().Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("operation", operation).Int("listener_count", len(listeners)).Msg("gala event processed")

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

// waitDurableIdle polls the River job table until no available or running jobs remain
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

		result, err := rc.JobList(ctx, river.NewJobListParams().States(rivertype.JobStateAvailable, rivertype.JobStateRunning, rivertype.JobStateScheduled).First(1))
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
