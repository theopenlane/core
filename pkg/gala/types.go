package gala

import (
	"encoding/json"
	"time"
)

// Headers defines operational metadata for an envelope
type Headers struct {
	// Properties stores additional metadata for UI visibility
	Properties map[string]string `json:"properties,omitempty"`
	// Tags are low-cardinality labels forwarded to the transport layer (e.g. River job tags)
	Tags []string `json:"tags,omitempty"`
	// Listeners are the registered listener names for the topic, populated at dispatch time
	Listeners []string `json:"listeners,omitempty"`
	// Kind optionally routes dispatch to a registered job kind
	Kind string `json:"kind,omitempty"`
	// ScheduledAt defers execution until the specified time; nil means immediate
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	// UniqueKey enforces at most one live job per key at insert time
	UniqueKey string `json:"unique_key,omitempty"`
	// SkipUniqueKey suppresses the topic's UniqueKey derivation
	SkipUniqueKey bool `json:"skip_unique_key,omitempty"`
	// UniqueOnce extends UniqueKey matching to terminal job states
	UniqueOnce bool `json:"unique_once,omitempty"`
	// Metadata carries structured operation context as opaque JSON
	Metadata json.RawMessage `json:"metadata,omitempty"`
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
