package types //nolint:revive

import (
	"encoding/json"
)

// MappingOverride is one mapping customization
type MappingOverride struct {
	// FilterExpr is the optional CEL expression used to filter provider payloads before mapping
	FilterExpr string `json:"filterExpr,omitempty"`
	// MapExpr is the CEL expression used to map provider payloads to the normalized schema
	MapExpr string `json:"mapExpr,omitempty"`
	// Links are the cross-object link rules applied when a record of this schema is ingested
	Links []LinkRule `json:"links,omitempty"`
}

// LinkRule describes one cross-object link and how to match candidates for it
type LinkRule struct {
	// TargetSchema is the entityops object type to link to (e.g. "Control")
	TargetSchema string `json:"targetSchema" jsonschema:"title=Target Object,description=The object type to cross-link the ingested record to"`
	// Edge selects the edge to link through when several edges reach the target type
	Edge string `json:"edge,omitempty" jsonschema:"title=Edge,description=Edge to link through when multiple edges reach the target object type"`
	// TargetField is the target match-key field to match against for a field match (e.g. "ref_code")
	TargetField string `json:"targetField,omitempty" jsonschema:"title=Target Field,description=Field on the target object to match"`
	// SourceField is the source scalar input key whose value must equal the target field
	SourceField string `json:"sourceField,omitempty" jsonschema:"title=Source Field,description=Field on the ingested record to match against the target field"`
	// SourceList is the source list input key whose elements are additional match values
	SourceList string `json:"sourceList,omitempty" jsonschema:"title=Source List Field,description=List field on the ingested record providing additional match values"`
	// Expression is a CEL match expression evaluated per candidate against the source
	Expression string `json:"expression,omitempty" jsonschema:"title=Match Expression,description=CEL expression matching target to source for non-equality conditions"`
}

// MappingRegistration declares one default mapping shipped with a definition
type MappingRegistration struct {
	// Schema is the normalized target schema for the mapping
	Schema string `json:"schema"`
	// Variant is the optional variant name within the schema
	Variant string `json:"variant,omitempty"`
	// Spec contains the mapping expressions for the schema and variant
	Spec MappingOverride `json:"spec"`
	// LinkTargets is the cross-link inventory for this schema, populated at registration
	LinkTargets []LinkTargetInfo `json:"linkTargets,omitempty"`
}

// LinkTargetInfo describes one edge an ingested record can be cross-linked through
type LinkTargetInfo struct {
	// Edge is the edge name the link applies to, disambiguating multiple edges to one type
	Edge string `json:"edge"`
	// TargetType is the object type that can be linked to (e.g. "Control")
	TargetType string `json:"targetType"`
	// Label is the human-readable label for the edge
	Label string `json:"label"`
	// TargetFields are the match-key fields on the target object valid as LinkRule.TargetField
	TargetFields []LinkFieldInfo `json:"targetFields,omitempty"`
	// SourceFields are the mapped input keys valid as LinkRule.SourceField or SourceList
	SourceFields []LinkFieldInfo `json:"sourceFields,omitempty"`
}

// LinkFieldInfo is one field available for cross-link matching
type LinkFieldInfo struct {
	// Name is the snake_case field name
	Name string `json:"name"`
	// Label is the human-readable label
	Label string `json:"label"`
	// Type is the field type
	Type string `json:"type"`
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

// MappingInstallation is the writing installation exposed to map expressions
type MappingInstallation struct {
	// ID is the installation id
	ID string `json:"id"`
	// Name is the installation's user-facing name
	Name string `json:"name"`
	// DefinitionID is the canonical id of the installed definition
	DefinitionID string `json:"definition_id"`
	// DefinitionName is the installed definition's display name
	DefinitionName string `json:"definition_name"`
	// InstanceID is the external system instance the installation connects to
	InstanceID string `json:"instance_id"`
	// PrimaryDirectory reports whether the installation is the org's authoritative directory source
	PrimaryDirectory bool `json:"primary_directory"`
}

// IngestPayloadSet groups mapping envelopes by normalized target schema
type IngestPayloadSet struct {
	// Schema is the normalized target schema emitted by the operation
	Schema string `json:"schema"`
	// Envelopes are the raw provider payloads to map and ingest
	Envelopes []MappingEnvelope `json:"envelopes,omitempty"`
	// SnapshotComplete marks this payload set as the provider's complete record set for its schema
	SnapshotComplete bool `json:"snapshotComplete,omitempty"`
}
