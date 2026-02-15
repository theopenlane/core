package gala

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/riverqueue/river"
	"github.com/samber/do/v2"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
)

// Config configures cohesive Gala startup
type Config struct {
	// Enabled toggles Gala worker startup and dispatch support when true
	Enabled bool
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
	registry *Registry
	// injector provides dependency resolution for listeners
	injector do.Injector
	// dispatcher handles envelope dispatch to listeners
	dispatcher Dispatcher
	// contextManager handles context capture and restoration
	contextManager *ContextManager
	// jobClient is the dedicated River client used for durable dispatch
	jobClient *riverqueue.Client
}

// NewGala initializes your gala, initializes dependencies, and starts workers
func NewGala(ctx context.Context, config Config) (app *Gala, err error) {
	if err := config.validate(); err != nil {
		return nil, err
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
		Queues: map[string]river.QueueConfig{
			config.QueueName: {MaxWorkers: config.WorkerCount},
		},
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

	if err := app.initialize(dispatcher); err != nil {
		return nil, err
	}

	app.jobClient = jobClient

	return app, nil
}

// initialize sets Gala core dependencies and default runtime services
// future expansion / features may necessitate passing in additional dependencies or a more complex runtime config object but avoiding pre-optimization
func (g *Gala) initialize(dispatcher Dispatcher) error {
	contextManager, err := NewContextManager(NewAuthContextCodec())
	if err != nil {
		return err
	}

	g.registry = NewRegistry()
	g.injector = do.New()
	g.dispatcher = dispatcher
	g.contextManager = contextManager

	return nil
}

// validate normalizes config defaults and validates required fields
func (c *Config) validate() error {
	if c.ConnectionURI == "" {
		return ErrRiverConnectionURIRequired
	}

	if c.QueueName == "" {
		c.QueueName = DefaultQueueName
	}

	if c.WorkerCount < 1 {
		c.WorkerCount = 1
	}

	return nil
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

// EmitEnvelope dispatches a pre-built envelope using its topic registration
func (g *Gala) EmitEnvelope(ctx context.Context, envelope Envelope) error {
	if _, err := g.registry.topicRegistration(envelope.Topic); err != nil {
		return err
	}

	if g.dispatcher == nil {
		return ErrDispatcherRequired
	}

	return g.dispatcher.Dispatch(ctx, envelope)
}

// DispatchEnvelope dispatches one envelope to all listeners on the topic
func (g *Gala) DispatchEnvelope(ctx context.Context, envelope Envelope) error {
	logx.FromContext(ctx).Debug().Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Msg("gala processing event")

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

	handlerContext := HandlerContext{
		Context:  restoredContext,
		Envelope: envelope,
		Injector: g.injector,
	}

	operation := payloadOperation(decodedPayload)
	listeners := g.registry.registeredListeners(envelope.Topic)

	for _, listener := range listeners {
		if !listenerInterestedInOperation(listener, operation) {
			continue
		}

		if err := g.executeListener(handlerContext, listener, decodedPayload); err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("listener", listener.name).Msg("gala listener failed")

			return err
		}

		logx.FromContext(ctx).Debug().Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Str("listener", listener.name).Msg("gala listener completed")
	}

	logx.FromContext(ctx).Debug().Str("event_id", string(envelope.ID)).Str("topic", string(envelope.Topic)).Int("listener_count", len(listeners)).Msg("gala event processed")

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
				Cause:        ErrListenerPanicked,
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

// Close closes the dedicated Gala queue client
func (g *Gala) Close() error {
	if g.jobClient == nil {
		return nil
	}

	if err := g.jobClient.Close(); err != nil {
		return ErrRiverClientCloseFailed
	}

	return nil
}
