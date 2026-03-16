package types

import "encoding/json"

// MappingOverride is one mapping customization
type MappingOverride struct {
	// Version is the optional version of the mapping spec, used for migration and compatibility checks
	Version string `json:"version,omitempty"`
	// FilterExpr is the optional CEL expression used to filter provider payloads before mapping
	FilterExpr string `json:"filterExpr,omitempty"`
	// MapExpr is the CEL expression used to map provider payloads to the normalized schema
	MapExpr string `json:"mapExpr,omitempty"`
}

// MappingRegistration declares one default mapping shipped with a definition
type MappingRegistration struct {
	// Schema is the normalized target schema for the mapping
	Schema string `json:"schema"`
	// Variant is the optional variant name within the schema
	Variant string `json:"variant,omitempty"`
	// Spec contains the mapping expressions for the schema and variant
	Spec MappingOverride `json:"spec"`
}

// MappingEnvelope wraps one provider payload for CEL filter and map evaluation
type MappingEnvelope struct {
	// Variant selects which mapping variant should be applied
	Variant string `json:"variant,omitempty"`
	// Resource identifies the provider resource associated with the payload
	Resource string `json:"resource,omitempty"`
	// Action identifies the provider event or collection action associated with the payload
	Action string `json:"action,omitempty"`
	// Payload is the raw provider payload
	Payload json.RawMessage `json:"payload,omitempty"`
}

// IngestPayloadSet groups mapping envelopes by normalized target schema
type IngestPayloadSet struct {
	// Schema is the normalized target schema emitted by the operation
	Schema string `json:"schema"`
	// Envelopes are the raw provider payloads to map and ingest
	Envelopes []MappingEnvelope `json:"envelopes,omitempty"`
}
