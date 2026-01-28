package observability

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
)

// Fields captures standard workflow log fields.
type Fields map[string]any

// Operation describes a workflow operation being observed.
type Operation struct {
	// Name is the workflow operation name.
	Name OperationName
	// Origin is the component emitting the observation.
	Origin Origin
	// TriggerEvent is the event type that caused the operation.
	TriggerEvent string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler for Operation.
func (op Operation) MarshalZerologObject(e *zerolog.Event) {
	e.Str(FieldOperation, string(op.Name)).
		Str(FieldOrigin, string(op.Origin))

	if op.TriggerEvent != "" {
		e.Str(FieldTriggerEvent, op.TriggerEvent)
	}
}

const (
	// FieldOperation is the log field key for the workflow operation
	FieldOperation = "workflow_op"
	// FieldOrigin is the log field key for the workflow origin
	FieldOrigin = "origin"
	// FieldTriggerEvent is the log field key for the trigger event
	FieldTriggerEvent = "trigger_event"
	// FieldTopic is the log field key for event topics
	FieldTopic = "topic"
	// FieldPayload is the log field key for event payloads
	FieldPayload = "payload"

	// FieldActionKey is the log field key for workflow action keys
	FieldActionKey = "action_key"
	// FieldObjectType is the log field key for object types
	FieldObjectType = "object_type"
	// FieldExpression is the log field key for CEL expressions
	FieldExpression = "expression"
	// FieldSkipReason is the log field key for skip reasons
	FieldSkipReason = "skip_reason"
)

// ActionFields returns an action-aware set of observability fields with optional overrides.
func ActionFields(actionKey string, extra Fields) Fields {
	if len(extra) == 0 {
		return Fields{
			FieldActionKey: actionKey,
		}
	}

	return lo.Assign(Fields{
		FieldActionKey: actionKey,
	}, extra)
}

const (
	msgOpFailed   = "workflow op failed"
	msgOpSkipped  = "workflow op skipped"
	msgOpWarning  = "workflow op warning" //nolint:gosec
	msgOpInfo     = "workflow op info"
	msgEmitFailed = "workflow emit failed"
)

// Observer standardizes workflow logging and metrics
type Observer struct {
	now func() time.Time
}

// New returns a workflow observer with defaults
func New() *Observer {
	return &Observer{now: time.Now}
}

// applyFields updates the logger stored in ctx with the provided fields
func applyFields(ctx context.Context, fields Fields) {
	logx.FromContext(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Fields(fields)
	})
}

// begin initializes a scoped workflow observation
func (o *Observer) begin(ctx context.Context, op Operation, fields Fields) *Scope {
	logger := logx.FromContext(ctx).With().EmbedObject(op).Fields(fields).Logger()

	return &Scope{
		observer: o,
		op:       op,
		start:    o.now(),
		ctx:      logger.WithContext(ctx),
	}
}

// handleEmit handles errors emitted by the event bus using context metadata when available
func (o *Observer) handleEmit(ctx context.Context, op Operation, fields Fields, topic string, errCh <-chan error) error {
	logger := logx.FromContext(ctx).With().EmbedObject(op).Fields(fields).Str(FieldTopic, topic).Logger()

	go func() {
		for err := range errCh {
			if err == nil {
				continue
			}

			logger.Error().Err(err).Msg(msgEmitFailed)
			metrics.RecordWorkflowEmitError(topic, string(op.Origin))
		}
	}()

	return nil
}

// warn logs a standardized workflow warning without starting a scope
func warn(ctx context.Context, op Operation, fields Fields, err error) {
	event := logx.FromContext(ctx).Warn().EmbedObject(op).Fields(fields)
	if err != nil {
		event = event.Err(err)
	}

	event.Msg(msgOpWarning)
}

// info logs a standardized workflow info event without starting a scope
func info(ctx context.Context, op Operation, fields Fields) {
	logx.FromContext(ctx).Info().EmbedObject(op).Fields(fields).Msg(msgOpInfo)
}

// Scope tracks a single observed workflow operation
type Scope struct {
	observer *Observer
	op       Operation
	start    time.Time
	ctx      context.Context

	skipped   bool
	recordErr error
}

// Context returns the scoped context with logger fields applied
func (s *Scope) Context() context.Context {
	return s.ctx
}

// WithFields appends fields to the scope logger and context
func (s *Scope) WithFields(fields Fields) {
	applyFields(s.ctx, fields)
}

// Skip marks the operation as skipped with an optional reason
func (s *Scope) Skip(reason string, fields Fields) {
	s.skipped = true
	s.WithFields(lo.Assign(fields, Fields{FieldSkipReason: reason}))
}

// Warn logs a standardized workflow warning
func (s *Scope) Warn(err error, fields Fields) {
	warn(s.ctx, s.op, fields, err)
}

// Info logs a standardized workflow info event
func (s *Scope) Info(fields Fields) {
	info(s.ctx, s.op, fields)
}

// RecordError captures an error to be logged when the scope ends
func (s *Scope) RecordError(err error, fields Fields) {
	if err == nil {
		return
	}

	s.WithFields(fields)
	if s.recordErr == nil {
		s.recordErr = err
	}
}

// Fail records an error with fields and returns it for the caller
func (s *Scope) Fail(err error, fields Fields) error {
	if err == nil {
		return nil
	}

	s.RecordError(err, fields)

	return err
}

// End finalizes the observation, logging and recording metrics.
func (s *Scope) End(err error, fields Fields) {
	s.WithFields(fields)
	finalErr := lo.CoalesceOrEmpty(err, s.recordErr)
	duration := s.observer.now().Sub(s.start)
	logger := logx.FromContext(s.ctx)
	if finalErr != nil {
		logger.Error().Err(finalErr).Msg(msgOpFailed)
		metrics.RecordWorkflowOperation(string(s.op.Name), string(s.op.Origin), s.op.TriggerEvent, duration, finalErr)

		return
	}

	if s.skipped {
		logger.Debug().Msg(msgOpSkipped)
	}

	metrics.RecordWorkflowOperation(string(s.op.Name), string(s.op.Origin), s.op.TriggerEvent, duration, nil)
}
