package graphapi

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/samber/do/v2"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
)

// ErrConnectionURIRequired is returned when SetupWorkflowEngine is called without a connection URI
var ErrConnectionURIRequired = errors.New("connection URI is required for durable workflow dispatch")

const (
	defaultPollInterval = 50 * time.Millisecond
	defaultPollTimeout  = 5 * time.Second
	defaultWorkerCount  = 5
	defaultFetchPoll    = 10 * time.Millisecond
)

var (
	// workflowTestQueueSeq generates unique queue names per setup to isolate River workers
	workflowTestQueueSeq           atomic.Uint64
	ErrTimedOutWaitingForCondition = errors.New("timed out waiting for condition")
	ErrClientRequired              = errors.New("client is required")
)

// pollUntil executes query repeatedly until condition returns true or timeout expires.
// Returns the latest query result and a timeout error when the condition is not satisfied in time.
func pollUntil[T any](ctx context.Context, timeout time.Duration, query func() (T, error), condition func(T) bool) (T, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		result, err := query()
		if err != nil {
			var zero T
			return zero, err
		}

		if condition(result) {
			return result, nil
		}

		select {
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		case <-time.After(defaultPollInterval):
		}
	}

	result, err := query()
	if err != nil {
		var zero T
		return zero, err
	}

	if condition(result) {
		return result, nil
	}

	return result, ErrTimedOutWaitingForCondition
}

// WorkflowTestSetup contains the workflow engine and gala runtime for tests.
type WorkflowTestSetup struct {
	Engine  *engine.WorkflowEngine
	Runtime *gala.Gala
}

// Teardown stops workflow runtime workers and releases connections.
// Call this via defer after SetupWorkflowEngine.
func (s *WorkflowTestSetup) Teardown() {
	if s == nil || s.Runtime == nil {
		return
	}

	_ = s.Runtime.StopWorkers(context.Background())
	_ = s.Runtime.Close()
}

// SetupWorkflowEngine creates a workflow engine with a durable Gala runtime for integration tests.
// Each call creates a fresh runtime with its own River queue, so connections are never stale.
// The EmitGalaEventHook is additive on the shared client; previous hooks referencing closed
// runtimes fail silently (errors are logged, not returned). Call Teardown on the returned
// setup when the test completes.
func SetupWorkflowEngine(ctx context.Context, client *generated.Client, connectionURI string) (*WorkflowTestSetup, error) {
	if client == nil {
		return nil, ErrClientRequired
	}

	if connectionURI == "" {
		return nil, ErrConnectionURIRequired
	}

	queueName := fmt.Sprintf("workflow_test_%d", workflowTestQueueSeq.Add(1))

	runtime, err := gala.NewGala(ctx, gala.Config{
		DispatchMode:      gala.DispatchModeDurable,
		ConnectionURI:     connectionURI,
		QueueName:         queueName,
		WorkerCount:       defaultWorkerCount,
		RunMigrations:     true,
		FetchCooldown:     time.Millisecond,
		FetchPollInterval: defaultFetchPoll,
	})
	if err != nil {
		return nil, err
	}

	client.Use(hooks.EmitGalaEventHook(func() *gala.Gala {
		return runtime
	}))

	if _, err := hooks.RegisterGalaWorkflowListeners(runtime.Registry()); err != nil {
		return nil, err
	}

	wfEngine, err := engine.NewWorkflowEngine(client, runtime)
	if err != nil {
		return nil, err
	}

	do.ProvideValue(runtime.Injector(), runtime)
	do.ProvideValue(runtime.Injector(), client)
	do.ProvideValue(runtime.Injector(), wfEngine)

	if err := runtime.StartWorkers(ctx); err != nil {
		return nil, err
	}

	client.WorkflowEngine = wfEngine

	return &WorkflowTestSetup{
		Engine:  wfEngine,
		Runtime: runtime,
	}, nil
}

// workflowsEnabled reports whether workflows are enabled for this resolver
func workflowsEnabled(client *generated.Client) bool {
	return client != nil && client.WorkflowEngine != nil
}

// WaitForInstanceState polls until the workflow instance reaches the expected state or times out.
func WaitForInstanceState(ctx context.Context, client *generated.Client, instanceID string, expectedState enums.WorkflowInstanceState) (*generated.WorkflowInstance, error) {
	return WaitForInstanceStateWithTimeout(ctx, client, instanceID, expectedState, defaultPollTimeout)
}

// WaitForInstanceStateWithTimeout polls until the workflow instance reaches the expected state or times out.
func WaitForInstanceStateWithTimeout(ctx context.Context, client *generated.Client, instanceID string, expectedState enums.WorkflowInstanceState, timeout time.Duration) (*generated.WorkflowInstance, error) {
	return pollUntil(ctx, timeout,
		func() (*generated.WorkflowInstance, error) {
			return client.WorkflowInstance.Query().Where(workflowinstance.IDEQ(instanceID)).Only(ctx)
		},
		func(instance *generated.WorkflowInstance) bool {
			return instance.State == expectedState
		},
	)
}

// WaitForAssignments polls until at least minCount assignments exist for the instance.
func WaitForAssignments(ctx context.Context, client *generated.Client, instanceID string, minCount int) ([]*generated.WorkflowAssignment, error) {
	return WaitForAssignmentsWithTimeout(ctx, client, instanceID, minCount, defaultPollTimeout)
}

// WaitForAssignmentsWithTimeout polls until at least minCount assignments exist for the instance or times out.
func WaitForAssignmentsWithTimeout(ctx context.Context, client *generated.Client, instanceID string, minCount int, timeout time.Duration) ([]*generated.WorkflowAssignment, error) {
	return pollUntil(ctx, timeout,
		func() ([]*generated.WorkflowAssignment, error) {
			return client.WorkflowAssignment.Query().Where(workflowassignment.WorkflowInstanceIDEQ(instanceID)).All(ctx)
		},
		func(assignments []*generated.WorkflowAssignment) bool {
			return len(assignments) >= minCount
		},
	)
}
