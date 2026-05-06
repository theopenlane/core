package gala

import (
	"encoding/json"
	"time"
)

// Headers defines operational metadata for an envelope
type Headers struct {
	// IdempotencyKey identifies duplicate-safe processing scope
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	// Properties stores additional metadata for UI visibility
	Properties map[string]string `json:"properties,omitempty"`
	// Tags are low-cardinality labels forwarded to the transport layer (e.g. River job tags)
	Tags []string `json:"tags,omitempty"`
	// Listeners are the registered listener names for the topic, populated at dispatch time
	Listeners []string `json:"listeners,omitempty"`
	// Queue optionally overrides the River queue used for dispatch
	Queue string `json:"queue,omitempty"`
	// MaxAttempts optionally overrides River max attempts for this envelope
	MaxAttempts int `json:"max_attempts,omitempty"`
	// ScheduledAt defers execution until the specified time; nil means immediate
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

// NewHeaders returns Headers with the given tags and the input marshaled as JSON in the "input" property
func NewHeaders(tags []string, input any) Headers {
	raw, err := json.Marshal(input)
	if err != nil {
		return Headers{Tags: tags}
	}

	return Headers{
		Tags:       tags,
		Properties: map[string]string{"input": string(raw)},
	}
}

// Envelope is the durable event envelope
type Envelope struct {
	// ID is the unique event identifier
	ID EventID `json:"id"`
	// Topic is the destination topic
	Topic TopicName `json:"topic"`
	// OccurredAt is the emit timestamp in UTC
	OccurredAt time.Time `json:"occurred_at"`
	// Headers holds operational metadata
	Headers Headers `json:"headers"`
	// Payload is encoded topic payload data
	Payload json.RawMessage `json:"payload"`
	// ContextSnapshot holds restorable context metadata
	ContextSnapshot ContextSnapshot `json:"context_snapshot"`
}

// EmitReceipt captures synchronous dispatch results
type EmitReceipt struct {
	// EventID is the emitted event identifier
	EventID EventID
	// Accepted reports whether the event was accepted for processing
	Accepted bool
	// Err contains any terminal emit error
	Err error
}
