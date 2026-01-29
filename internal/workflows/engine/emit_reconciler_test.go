//go:build test

package engine_test

import (
	"context"
	"errors"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/internal/workflows/reconciler"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/utils/ulids"
)

type failingQueueStore struct {
	err error
}

// SaveEvent simulates an enqueue failure.
func (s *failingQueueStore) SaveEvent(event soiree.Event) error {
	return s.err
}

// SaveHandlerResult is a no-op for tests.
func (s *failingQueueStore) SaveHandlerResult(event soiree.Event, handlerID string, err error) error {
	return nil
}

// DequeueEvent blocks until the context is cancelled.
func (s *failingQueueStore) DequeueEvent(ctx context.Context) (soiree.Event, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

// clearEmitFailedEvents removes emit failure events to isolate test cases.
func (s *WorkflowEngineTestSuite) clearEmitFailedEvents(ctx context.Context) {
	allowCtx := workflows.AllowContext(ctx)
	_, err := s.client.WorkflowEvent.Delete().
		Where(workflowevent.EventTypeEQ(enums.WorkflowEventTypeEmitFailed)).
		Exec(allowCtx)
	s.Require().NoError(err)
}

// TestEmitFailureRecorded verifies that when event emission fails (e.g., queue unavailable),
// the failure is recorded as an EMIT_FAILED WorkflowEvent with full diagnostic details.
//
// Test Flow:
//  1. Creates a workflow engine with a broken emitter (simulates queue failure)
//  2. Triggers a workflow (instance created, but event emission fails)
//  3. Verifies an EMIT_FAILED event was recorded with:
//     - Topic: The intended destination topic
//     - OriginalEventType: The event that failed to emit
//     - ObjectID/ObjectType: Context about the trigger object
//     - Attempts: Number of emit attempts (starts at 1)
//     - EventID: Unique identifier for correlation
//     - Payload: The original event payload
//     - LastError: The error message from the failed emit
//
// Why This Matters:
//   Emit failures should not be silently lost. Recording them enables monitoring, alerting,
//   and automated recovery through the reconciliation process.
func (s *WorkflowEngineTestSuite) TestEmitFailureRecorded() {
	s.ClearWorkflowDefinitions()

	_, orgID, userCtx := s.SetupTestUser()
	s.clearEmitFailedEvents(userCtx)
	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-TEST-" + ulids.New().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	failBus := soiree.New(soiree.EventStore(&failingQueueStore{err: errors.New("enqueue failed")}))
	defer func() { _ = failBus.Close() }()

	wfEngine := s.NewIsolatedEngine(failBus)

	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType: "UPDATE",
	})
	s.Require().NoError(err)
	s.Require().NotNil(instance)

	allowCtx := workflows.AllowContext(userCtx)
	event, err := s.client.WorkflowEvent.Query().
		Where(
			workflowevent.WorkflowInstanceIDEQ(instance.ID),
			workflowevent.EventTypeEQ(enums.WorkflowEventTypeEmitFailed),
		).
		Only(allowCtx)
	s.Require().NoError(err)

	s.Equal(enums.WorkflowEventTypeEmitFailed, event.EventType)
	s.Equal(enums.WorkflowEventTypeEmitFailed, event.Payload.EventType)

	details, err := workflows.ParseEmitFailureDetails(event.Payload.Details)
	s.Require().NoError(err)
	s.Equal(soiree.WorkflowTriggeredTopic.Name(), details.Topic)
	s.Equal(enums.WorkflowEventTypeInstanceTriggered, details.OriginalEventType)
	s.Equal(obj.ID, details.ObjectID)
	s.Equal(obj.Type, details.ObjectType)
	s.Equal(1, details.Attempts)
	s.NotEmpty(details.EventID)
	s.NotEmpty(details.Payload)
	s.Contains(details.LastError, "enqueue failed")
}

// TestReconcileEmitFailureRecovers verifies that the reconciler can recover from emit failures
// by re-emitting the event to a healthy emitter and marking the failure as recovered.
//
// Test Flow:
//  1. Creates a workflow engine with a broken emitter (failure recorded)
//  2. Triggers a workflow (EMIT_FAILED event recorded)
//  3. Creates a reconciler using the suite's healthy emitter
//  4. Calls ReconcileEmitFailures
//  5. Verifies: Attempted = 1, Recovered = 1
//  6. Verifies an EMIT_RECOVERED event was recorded
//  7. Verifies the workflow instance eventually completed (event processed)
//
// Why This Matters:
//   The reconciliation process enables self-healing. When infrastructure recovers, the
//   reconciler can retry failed emissions without losing workflow progress.
func (s *WorkflowEngineTestSuite) TestReconcileEmitFailureRecovers() {
	s.ClearWorkflowDefinitions()

	_, orgID, userCtx := s.SetupTestUser()
	s.clearEmitFailedEvents(userCtx)
	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-TEST-" + ulids.New().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	failBus := soiree.New(soiree.EventStore(&failingQueueStore{err: errors.New("enqueue failed")}))
	defer func() { _ = failBus.Close() }()

	wfEngine := s.NewIsolatedEngine(failBus)
	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType: "UPDATE",
	})
	s.Require().NoError(err)
	s.Require().NotNil(instance)

	// Use the suite's real emitter for recovery (listeners already registered)
	rec, err := reconciler.New(s.client, s.eventer.Emitter)
	s.Require().NoError(err)

	result, err := rec.ReconcileEmitFailures(userCtx)
	s.Require().NoError(err)
	s.Equal(1, result.Attempted)
	s.Equal(1, result.Recovered)

	allowCtx := workflows.AllowContext(userCtx)
	recovered, err := s.client.WorkflowEvent.Query().
		Where(
			workflowevent.WorkflowInstanceIDEQ(instance.ID),
			workflowevent.EventTypeEQ(enums.WorkflowEventTypeEmitRecovered),
		).
		Only(allowCtx)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowEventTypeEmitRecovered, recovered.EventType)

	updated, err := s.client.WorkflowInstance.Get(allowCtx, instance.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, updated.State)
}

// TestReconcileEmitFailureTerminalAfterMaxAttempts verifies that after exhausting the maximum
// retry attempts, the reconciler marks the failure as terminal and fails the workflow instance.
//
// Test Setup:
//   - Reconciler configured with MaxAttempts = 3
//   - Emitter that always fails (simulates persistent infrastructure issue)
//
// Test Flow:
//  1. Creates a workflow engine with a broken emitter (EMIT_FAILED recorded, attempts = 1)
//  2. First reconciliation attempt (emitter still broken):
//     - Attempts incremented to 2
//     - EMIT_FAILED event updated
//  3. Second reconciliation attempt:
//     - Attempts = 3 (max reached)
//     - EMIT_FAILED_TERMINAL event recorded
//     - Workflow instance marked as FAILED
//
// Why This Matters:
//   Infinite retries could waste resources on permanently broken workflows. After max
//   attempts, the system must give up gracefully, mark the failure as terminal, and
//   fail the workflow instance for human investigation.
func (s *WorkflowEngineTestSuite) TestReconcileEmitFailureTerminalAfterMaxAttempts() {
	s.ClearWorkflowDefinitions()

	_, orgID, userCtx := s.SetupTestUser()
	s.clearEmitFailedEvents(userCtx)
	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-TEST-" + ulids.New().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	failBus := soiree.New(soiree.EventStore(&failingQueueStore{err: errors.New("enqueue failed")}))
	defer func() { _ = failBus.Close() }()

	wfEngine := s.NewIsolatedEngine(failBus)
	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType: "UPDATE",
	})
	s.Require().NoError(err)
	s.Require().NotNil(instance)

	rec, err := reconciler.New(s.client, failBus, reconciler.WithMaxAttempts(3))
	s.Require().NoError(err)

	allowCtx := workflows.AllowContext(userCtx)

	_, err = rec.ReconcileEmitFailures(userCtx)
	s.Require().NoError(err)

	retryEvent, err := s.client.WorkflowEvent.Query().
		Where(
			workflowevent.WorkflowInstanceIDEQ(instance.ID),
			workflowevent.EventTypeEQ(enums.WorkflowEventTypeEmitFailed),
		).
		Only(allowCtx)
	s.Require().NoError(err)

	details, err := workflows.ParseEmitFailureDetails(retryEvent.Payload.Details)
	s.Require().NoError(err)
	s.Equal(2, details.Attempts)

	_, err = rec.ReconcileEmitFailures(userCtx)
	s.Require().NoError(err)

	terminalEvent, err := s.client.WorkflowEvent.Query().
		Where(
			workflowevent.WorkflowInstanceIDEQ(instance.ID),
			workflowevent.EventTypeEQ(enums.WorkflowEventTypeEmitFailedTerminal),
		).
		Only(allowCtx)
	s.Require().NoError(err)

	terminalDetails, err := workflows.ParseEmitFailureDetails(terminalEvent.Payload.Details)
	s.Require().NoError(err)
	s.Equal(3, terminalDetails.Attempts)

	updated, err := s.client.WorkflowInstance.Get(allowCtx, instance.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateFailed, updated.State)
}
