package graphapi

import (
	"context"
	"errors"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
)

const (
	defaultPollInterval = 50 * time.Millisecond
	defaultPollTimeout  = 5 * time.Second
)

var ErrTimedOutWaitingForCondition = errors.New("timed out waiting for condition")

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

// workflowsEnabled reports whether the process-wide workflow engine is registered
func workflowsEnabled() bool {
	return engine.Enabled()
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
