package reconciler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/events/soiree"
)

const defaultMaxEmitAttempts = 3

// EmitReconcileResult summarizes a reconciliation pass over emit failures
type EmitReconcileResult struct {
	// Attempted is the count of retry attempts
	Attempted int
	// Recovered is the count of recovered emit failures
	Recovered int
	// Failed is the count of emit failures that remain retryable
	Failed int
	// Terminal is the count of emit failures marked terminal
	Terminal int
	// Skipped is the count of events skipped during reconciliation
	Skipped int
}

// Option configures a Reconciler
type Option func(*Reconciler)

// WithMaxAttempts overrides the default max retry attempts
func WithMaxAttempts(maxRetries int) Option {
	return func(r *Reconciler) {
		if maxRetries > 0 {
			r.maxAttempts = maxRetries
		}
	}
}

// Reconciler re-emits workflow events that failed to enqueue
type Reconciler struct {
	client      *generated.Client
	emitter     soiree.Emitter
	maxAttempts int
}

// New constructs a workflow reconciler
func New(client *generated.Client, emitter soiree.Emitter, opts ...Option) (*Reconciler, error) {
	if client == nil {
		return nil, ErrNilClient
	}

	r := &Reconciler{
		client:      client,
		emitter:     emitter,
		maxAttempts: defaultMaxEmitAttempts,
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.maxAttempts <= 0 {
		r.maxAttempts = defaultMaxEmitAttempts
	}

	return r, nil
}

// ReconcileEmitFailures retries deterministic workflow emits that failed to enqueue
func (r *Reconciler) ReconcileEmitFailures(ctx context.Context) (EmitReconcileResult, error) {
	var result EmitReconcileResult

	allowCtx := workflows.AllowContext(ctx)

	events, err := r.client.WorkflowEvent.Query().
		Where(workflowevent.EventTypeEQ(enums.WorkflowEventTypeEmitFailed)).
		All(allowCtx)
	if err != nil {
		return result, err
	}

	var joinedErr error
	for _, evt := range events {
		eventResult, err := r.processEmitFailure(allowCtx, evt)
		result.Attempted += eventResult.Attempted
		result.Recovered += eventResult.Recovered
		result.Failed += eventResult.Failed
		result.Terminal += eventResult.Terminal
		result.Skipped += eventResult.Skipped
		if err != nil {
			joinedErr = errors.Join(joinedErr, err)
		}
	}

	return result, joinedErr
}

// processEmitFailure reconciles a single emit failure event
func (r *Reconciler) processEmitFailure(ctx context.Context, evt *generated.WorkflowEvent) (EmitReconcileResult, error) {
	var result EmitReconcileResult

	details, detailsErr := workflows.ParseEmitFailureDetails(evt.Payload.Details)
	if detailsErr != nil {
		result.Terminal++

		return result, r.markTerminal(ctx, evt, nil, fmt.Errorf("invalid emit failure details: %w", detailsErr))
	}

	detailsUpdated := false
	if details.Attempts <= 0 {
		details.Attempts = 1
		detailsUpdated = true
	}

	if details.Topic == "" || len(details.Payload) == 0 {
		result.Terminal++

		return result, r.markTerminal(ctx, evt, &details, ErrEmitFailureMissingPayload)
	}

	if details.Attempts >= r.maxAttempts {
		result.Terminal++

		return result, r.markTerminal(ctx, evt, &details, nil)
	}

	result.Attempted++
	receipt := r.reemit(ctx, &details)
	if details.EventID == "" && receipt.EventID != "" {
		details.EventID = receipt.EventID
		detailsUpdated = true
	}

	if receipt.Err == nil {
		result.Recovered++

		return result, r.updateWorkflowEvent(ctx, evt, enums.WorkflowEventTypeEmitRecovered, nil, &details, detailsUpdated)
	}

	details.Attempts++
	details.LastError = receipt.Err.Error()

	if details.Attempts >= r.maxAttempts {
		result.Terminal++

		return result, r.markTerminal(ctx, evt, &details, receipt.Err)
	}

	result.Failed++

	return result, r.updateWorkflowEvent(ctx, evt, enums.WorkflowEventTypeEmitFailed, receipt.Err, &details, true)
}

// reemit builds an event from stored details and emits it through the configured emitter
func (r *Reconciler) reemit(ctx context.Context, details *workflows.EmitFailureDetails) workflows.EmitReceipt {
	event := soiree.NewBaseEvent(details.Topic, json.RawMessage(details.Payload))
	props := event.Properties()
	if details.EventID != "" {
		props[soiree.PropertyEventID] = details.EventID
		event.SetProperties(props)
	}

	return workflows.EmitWorkflowEventWithEvent(ctx, r.emitter, details.Topic, event, r.client)
}

// markTerminal marks a failure event as terminal and updates the workflow instance state
func (r *Reconciler) markTerminal(ctx context.Context, evt *generated.WorkflowEvent, details *workflows.EmitFailureDetails, emitErr error) error {
	if details != nil && emitErr != nil {
		details.LastError = emitErr.Error()
		if details.Attempts <= 0 {
			details.Attempts = 1
		}
		if details.Attempts < r.maxAttempts {
			details.Attempts = r.maxAttempts
		}
	}

	if err := r.updateWorkflowEvent(ctx, evt, enums.WorkflowEventTypeEmitFailedTerminal, emitErr, details, details != nil); err != nil {
		return err
	}

	if evt.WorkflowInstanceID == "" {
		return nil
	}

	_, err := r.client.WorkflowInstance.Update().
		Where(
			workflowinstance.IDEQ(evt.WorkflowInstanceID),
			workflowinstance.Not(workflowinstance.StateIn(enums.WorkflowInstanceStateCompleted, enums.WorkflowInstanceStateFailed)),
		).
		SetState(enums.WorkflowInstanceStateFailed).
		Save(ctx)
	if err != nil && !generated.IsNotFound(err) {
		return err
	}

	return nil
}

// updateWorkflowEvent persists updated emit failure state
func (r *Reconciler) updateWorkflowEvent(ctx context.Context, evt *generated.WorkflowEvent, eventType enums.WorkflowEventType, emitErr error, details *workflows.EmitFailureDetails, updateDetails bool) error {
	if evt == nil {
		return nil
	}

	payload := evt.Payload
	payload.EventType = eventType
	if payload.ActionKey == "" && details != nil && details.ActionKey != "" {
		payload.ActionKey = details.ActionKey
	}

	if updateDetails && details != nil {
		if emitErr != nil {
			details.LastError = emitErr.Error()
		}
		encoded, err := json.Marshal(details)
		if err != nil {
			return err
		}
		payload.Details = encoded
	}

	update := r.client.WorkflowEvent.UpdateOneID(evt.ID).
		SetEventType(eventType).
		SetPayload(payload)

	return update.Exec(ctx)
}
