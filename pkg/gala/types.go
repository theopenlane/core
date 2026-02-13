package gala

import (
	"encoding/json"
	"time"

	"github.com/theopenlane/utils/ulids"
)

const (
	defaultSchemaVersion = 1
)

// EventID is a stable identifier used for idempotency and traceability.
type EventID string

// NewEventID creates a new event identifier.
func NewEventID() EventID {
	return EventID(ulids.New().String())
}

// TopicName is the stable string identifier for a topic.
type TopicName string

// Topic defines a strongly typed topic contract.
type Topic[T any] struct {
	// Name is the stable topic identifier.
	Name TopicName
	// SchemaVersion is the payload schema version for this topic.
	SchemaVersion int
}

// EffectiveSchemaVersion returns the configured schema version or the default.
func (t Topic[T]) EffectiveSchemaVersion() int {
	if t.SchemaVersion > 0 {
		return t.SchemaVersion
	}

	return defaultSchemaVersion
}

// EmitMode controls dispatch behavior for a topic.
type EmitMode string

const (
	// EmitModeInline dispatches in-process only.
	EmitModeInline EmitMode = "inline"
	// EmitModeDurable dispatches through a durable queue.
	EmitModeDurable EmitMode = "durable"
	// EmitModeDual dispatches to both inline and durable paths.
	EmitModeDual EmitMode = "dual"
)

// QueueClass identifies workload queue families.
type QueueClass string

const (
	// QueueClassWorkflow is the queue class for workflow events.
	QueueClassWorkflow QueueClass = "workflow"
	// QueueClassIntegration is the queue class for integration workloads.
	QueueClassIntegration QueueClass = "integration"
	// QueueClassGeneral is the queue class for general event workloads.
	QueueClassGeneral QueueClass = "general"
)

// TopicPolicy defines dispatch policy for a topic.
type TopicPolicy struct {
	// EmitMode controls inline versus durable behavior.
	EmitMode EmitMode
	// QueueClass chooses queue partitioning for durable workers.
	QueueClass QueueClass
	// QueueName optionally overrides queue selection for durable dispatch.
	QueueName string
	// Durable reports whether durable dispatch is required.
	Durable bool
}

// EffectiveEmitMode returns the effective emit mode for the topic policy.
func (p TopicPolicy) EffectiveEmitMode() EmitMode {
	if p.EmitMode != "" {
		return p.EmitMode
	}

	if p.Durable {
		return EmitModeDurable
	}

	return EmitModeInline
}

// Headers defines operational metadata for an envelope.
type Headers struct {
	// CorrelationID links events within the same high-level flow.
	CorrelationID string `json:"correlation_id,omitempty"`
	// CausationID points to the triggering upstream event ID.
	CausationID string `json:"causation_id,omitempty"`
	// IdempotencyKey identifies duplicate-safe processing scope.
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	// Properties stores additional typed metadata projected to string values.
	Properties map[string]string `json:"properties,omitempty"`
}

// ContextKey identifies a restorable context value key.
type ContextKey string

// ContextFlag identifies a boolean context flag.
type ContextFlag string

const (
	// ContextFlagWorkflowBypass marks workflow bypass behavior.
	ContextFlagWorkflowBypass ContextFlag = "workflow_bypass"
	// ContextFlagWorkflowAllowEventEmission allows workflow listener execution while bypass is set.
	ContextFlagWorkflowAllowEventEmission ContextFlag = "workflow_allow_event_emission"
)

// ContextSnapshot captures context data that can be restored after durable hops.
type ContextSnapshot struct {
	// Values contains codec-managed context values.
	Values map[ContextKey]json.RawMessage `json:"values,omitempty"`
	// Flags contains boolean context flags.
	Flags map[ContextFlag]bool `json:"flags,omitempty"`
}

// Envelope is the durable event envelope.
type Envelope struct {
	// ID is the unique event identifier.
	ID EventID `json:"id"`
	// Topic is the destination topic.
	Topic TopicName `json:"topic"`
	// SchemaVersion is the payload schema version.
	SchemaVersion int `json:"schema_version"`
	// OccurredAt is the emit timestamp in UTC.
	OccurredAt time.Time `json:"occurred_at"`
	// Headers holds operational metadata.
	Headers Headers `json:"headers"`
	// Payload is encoded topic payload data.
	Payload json.RawMessage `json:"payload"`
	// ContextSnapshot holds restorable context metadata.
	ContextSnapshot ContextSnapshot `json:"context_snapshot"`
}

// EmitReceipt captures synchronous dispatch results.
type EmitReceipt struct {
	// EventID is the emitted event identifier.
	EventID EventID
	// Accepted reports whether the event was accepted for processing.
	Accepted bool
	// Err contains any terminal emit error.
	Err error
}
