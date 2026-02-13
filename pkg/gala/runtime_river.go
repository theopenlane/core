package gala

import (
	"context"
	"errors"

	"github.com/riverqueue/river"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
)

// RiverRuntimeOptions configures NewRiverRuntime setup.
type RiverRuntimeOptions struct {
	// ConnectionURI is the database connection URI used for the dedicated gala river client.
	ConnectionURI string
	// QueueName is the gala queue used for durable dispatch jobs.
	QueueName string
	// WorkerCount is the max worker concurrency for the gala queue when workers are enabled.
	WorkerCount int
	// MaxRetries sets max attempts for gala dispatch jobs when greater than zero.
	MaxRetries int
	// WorkersEnabled toggles registration/startup support for dedicated gala workers.
	WorkersEnabled bool
	// QueueByClass optionally overrides queue class mappings for durable dispatch.
	QueueByClass map[QueueClass]string
	// RuntimeOptions configures runtime creation.
	RuntimeOptions RuntimeOptions
	// ConfigureRuntime applies app-specific runtime setup after runtime construction.
	ConfigureRuntime func(*Runtime) error
}

// RiverRuntime bundles gala runtime state with its dedicated river queue client lifecycle.
type RiverRuntime struct {
	runtime        *Runtime
	jobClient      *riverqueue.Client
	workersEnabled bool
}

// NewRiverRuntime creates a dedicated river-backed gala runtime and queue client.
func NewRiverRuntime(ctx context.Context, options RiverRuntimeOptions) (*RiverRuntime, error) {
	if options.ConnectionURI == "" {
		return nil, ErrRiverConnectionURIRequired
	}

	queueName := options.QueueName
	if queueName == "" {
		queueName = DefaultQueueName
	}

	workerCount := options.WorkerCount
	if workerCount < 1 {
		workerCount = 1
	}

	var runtimeRef *Runtime
	jobOpts := []riverqueue.Option{
		riverqueue.WithConnectionURI(options.ConnectionURI),
	}

	if options.WorkersEnabled {
		workers := river.NewWorkers()
		if err := AddRiverDispatchWorker(workers, func() *Runtime {
			return runtimeRef
		}); err != nil {
			return nil, err
		}

		jobOpts = append(jobOpts,
			riverqueue.WithWorkers(workers),
			riverqueue.WithQueues(map[string]river.QueueConfig{
				queueName: {MaxWorkers: workerCount},
			}),
		)
	}

	if options.MaxRetries > 0 {
		jobOpts = append(jobOpts, riverqueue.WithMaxRetries(options.MaxRetries))
	}

	jobClient, err := riverqueue.New(ctx, jobOpts...)
	if err != nil {
		return nil, ErrRiverClientInitializationFailed
	}

	dispatcher, err := NewRiverDispatcher(RiverDispatcherOptions{
		JobClient:    jobClient,
		QueueByClass: queueByClassForRuntime(queueName, options.QueueByClass),
		DefaultQueue: queueName,
	})
	if err != nil {
		_ = jobClient.Close()
		return nil, err
	}

	runtimeOptions := options.RuntimeOptions
	runtimeOptions.DurableDispatcher = dispatcher

	runtime, err := NewRuntime(runtimeOptions)
	if err != nil {
		_ = jobClient.Close()
		return nil, err
	}

	if options.ConfigureRuntime != nil {
		if err := options.ConfigureRuntime(runtime); err != nil {
			_ = jobClient.Close()
			return nil, ErrRuntimeConfigureFailed
		}
	}

	runtimeRef = runtime

	return &RiverRuntime{
		runtime:        runtime,
		jobClient:      jobClient,
		workersEnabled: options.WorkersEnabled,
	}, nil
}

// Runtime returns the configured runtime instance.
func (r *RiverRuntime) Runtime() *Runtime {
	if r == nil {
		return nil
	}

	return r.runtime
}

// JobClient returns the dedicated river queue client.
func (r *RiverRuntime) JobClient() *riverqueue.Client {
	if r == nil {
		return nil
	}

	return r.jobClient
}

// StartWorkers starts gala workers when worker mode is enabled.
func (r *RiverRuntime) StartWorkers(ctx context.Context) error {
	if r == nil || !r.workersEnabled {
		return nil
	}

	if r.jobClient == nil {
		return ErrRiverJobClientRequired
	}

	if err := r.jobClient.GetRiverClient().Start(ctx); err != nil {
		return ErrRiverWorkerStartFailed
	}

	return nil
}

// StopWorkers stops gala workers when worker mode is enabled.
func (r *RiverRuntime) StopWorkers(ctx context.Context) error {
	if r == nil || !r.workersEnabled {
		return nil
	}

	if r.jobClient == nil {
		return ErrRiverJobClientRequired
	}

	if err := r.jobClient.GetRiverClient().Stop(ctx); err != nil &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, context.DeadlineExceeded) {
		return ErrRiverWorkerStopFailed
	}

	return nil
}

// Close closes the dedicated gala queue client.
func (r *RiverRuntime) Close() error {
	if r == nil || r.jobClient == nil {
		return nil
	}

	if err := r.jobClient.Close(); err != nil {
		return ErrRiverClientCloseFailed
	}

	return nil
}

// queueByClassForRuntime builds queue-class mappings for a river-backed runtime.
func queueByClassForRuntime(defaultQueue string, overrides map[QueueClass]string) map[QueueClass]string {
	queueByClass := map[QueueClass]string{
		QueueClassWorkflow:    defaultQueue,
		QueueClassIntegration: defaultQueue,
		QueueClassGeneral:     defaultQueue,
	}

	for class, queueName := range overrides {
		queueByClass[class] = queueName
	}

	return queueByClass
}
