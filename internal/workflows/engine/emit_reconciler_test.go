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

// setupEmitterWithListeners wires the workflow engine to a synchronous event bus for tests.
func (s *WorkflowEngineTestSuite) setupEmitterWithListeners() (*engine.WorkflowEngine, *syncBusEmitter, *soiree.EventBus) {
	bus := soiree.New()
	emitter := &syncBusEmitter{bus: bus}
	wfEngine := s.NewTestEngine(emitter)

	listeners := engine.NewWorkflowListeners(s.client, wfEngine, emitter)
	bindings := []soiree.ListenerBinding{
		soiree.BindListener(soiree.WorkflowTriggeredTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowTriggeredPayload) error {
			return listeners.HandleWorkflowTriggered(ctx, payload)
		}),
		soiree.BindListener(soiree.WorkflowActionStartedTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowActionStartedPayload) error {
			return listeners.HandleActionStarted(ctx, payload)
		}),
		soiree.BindListener(soiree.WorkflowActionCompletedTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowActionCompletedPayload) error {
			return listeners.HandleActionCompleted(ctx, payload)
		}),
		soiree.BindListener(soiree.WorkflowAssignmentCompletedTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowAssignmentCompletedPayload) error {
			return listeners.HandleAssignmentCompleted(ctx, payload)
		}),
		soiree.BindListener(soiree.WorkflowInstanceCompletedTopic, func(ctx *soiree.EventContext, payload soiree.WorkflowInstanceCompletedPayload) error {
			return listeners.HandleInstanceCompleted(ctx, payload)
		}),
	}

	for _, binding := range bindings {
		_, err := binding.Register(bus)
		s.Require().NoError(err)
	}

	return wfEngine, emitter, bus
}

// clearEmitFailedEvents removes emit failure events to isolate test cases.
func (s *WorkflowEngineTestSuite) clearEmitFailedEvents(ctx context.Context) {
	allowCtx := workflows.AllowContext(ctx)
	_, err := s.client.WorkflowEvent.Delete().
		Where(workflowevent.EventTypeEQ(enums.WorkflowEventTypeEmitFailed)).
		Exec(allowCtx)
	s.Require().NoError(err)
}

// TestEmitFailureRecorded verifies emit failures are recorded
func (s *WorkflowEngineTestSuite) TestEmitFailureRecorded() {
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

	wfEngine := s.NewTestEngine(failBus)

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

// TestReconcileEmitFailureRecovers verifies reconciler recovers emit failures
func (s *WorkflowEngineTestSuite) TestReconcileEmitFailureRecovers() {
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

	wfEngine := s.NewTestEngine(failBus)
	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType: "UPDATE",
	})
	s.Require().NoError(err)
	s.Require().NotNil(instance)

	_, emitter, bus := s.setupEmitterWithListeners()
	defer func() { _ = bus.Close() }()

	rec, err := reconciler.New(s.client, emitter)
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

// TestReconcileEmitFailureTerminalAfterMaxAttempts verifies terminal behavior after max attempts
func (s *WorkflowEngineTestSuite) TestReconcileEmitFailureTerminalAfterMaxAttempts() {
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

	wfEngine := s.NewTestEngine(failBus)
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
