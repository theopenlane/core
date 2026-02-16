package graphapi

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/workflows/engine"
)

const (
	defaultPollInterval = 50 * time.Millisecond
	defaultPollTimeout  = 5 * time.Second
)

var (
	workflowTestSetupMu            sync.Mutex
	workflowTestSetups             = map[*generated.Client]*WorkflowTestSetup{}
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

// WorkflowTestSetup contains the workflow engine and eventer for tests
type WorkflowTestSetup struct {
	Engine  *engine.WorkflowEngine
	Eventer *hooks.Eventer
}

// SetupWorkflowEngine creates a workflow engine with a real eventer for integration tests.
// This wires up event listeners so that assignment completions and other workflow events
// are processed through the actual event-driven flow used in production.
func SetupWorkflowEngine(client *generated.Client) (*WorkflowTestSetup, error) {
	if client == nil {
		return nil, ErrClientRequired
	}

	workflowTestSetupMu.Lock()
	defer workflowTestSetupMu.Unlock()

	if existing, ok := workflowTestSetups[client]; ok && existing != nil {
		client.WorkflowEngine = existing.Engine

		return existing, nil
	}

	eventer := hooks.NewEventerPool(client)
	hooks.RegisterGlobalHooks(client, eventer)

	if err := hooks.RegisterListeners(eventer); err != nil {
		return nil, err
	}

	wfEngine, err := engine.NewWorkflowEngine(client, eventer.Emitter)
	if err != nil {
		return nil, err
	}

	client.WorkflowEngine = wfEngine

	setup := &WorkflowTestSetup{
		Engine:  wfEngine,
		Eventer: eventer,
	}

	workflowTestSetups[client] = setup

	return setup, nil
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
