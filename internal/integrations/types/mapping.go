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
	// Links are the cross-object link rules applied when a record of this schema is ingested,
	// declared on the definition's mapping and validated at registration; installations do not
	// configure them
	Links []LinkRule `json:"links,omitempty"`
}

// LinkRule describes one cross-object link: which target object type to link the ingested record to
// and how to match candidates — either a field match (target field equals a source field/list value,
// pushed into the query as an indexed predicate) or a CEL expression evaluated per candidate.
// Field names are the entityops catalog identifiers surfaced in the mapping's LinkTargets inventory:
// TargetField is a match-key column on the target, SourceField/SourceList are mapping input keys on
// the source. Registration validates every rule against the catalog so typos fail at startup
type LinkRule struct {
	// TargetSchema is the entityops object type to link to (e.g. "Control")
	TargetSchema string `json:"targetSchema" jsonschema:"title=Target Object,description=The object type to cross-link the ingested record to"`
	// Edge selects the edge to link through when the source schema has more than one edge to the
	// target type (e.g. "editors" vs "viewers"); optional when exactly one edge targets the type
	Edge string `json:"edge,omitempty" jsonschema:"title=Edge,description=Edge to link through when multiple edges reach the target object type"`
	// TargetField is the target match-key field to match against for a field match (e.g. "ref_code")
	TargetField string `json:"targetField,omitempty" jsonschema:"title=Target Field,description=Field on the target object to match"`
	// SourceField is the source scalar input key whose value must equal the target field
	SourceField string `json:"sourceField,omitempty" jsonschema:"title=Source Field,description=Field on the ingested record to match against the target field"`
	// SourceList is the source list input key whose elements are additional match values
	SourceList string `json:"sourceList,omitempty" jsonschema:"title=Source List Field,description=List field on the ingested record providing additional match values"`
	// Expression is a CEL match expression evaluated per candidate; "target" is the candidate and
	// "source" is the ingested record. Used instead of a field match for non-equality conditions
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
	// LinkTargets is the cross-link inventory for this schema (the object types it can link to and the
	// match fields on each side), populated at registration so the UI can render the dropdown + pickers
	LinkTargets []LinkTargetInfo `json:"linkTargets,omitempty"`
}

// LinkTargetInfo describes one edge an ingested record of a schema can be cross-linked through,
// with the fields available on each side for composing a LinkRule. One entry is emitted per edge,
// so a schema with several edges to the same target type surfaces each as a distinct choice
type LinkTargetInfo struct {
	// Edge is the edge name the link applies to (e.g. "controls"); it disambiguates targets when
	// the source schema has multiple edges to the same object type
	Edge string `json:"edge"`
	// TargetType is the object type that can be linked to (e.g. "Control")
	TargetType string `json:"targetType"`
	// Label is the human-readable label for the edge
	Label string `json:"label"`
	// TargetFields are the match-key fields on the target object valid as LinkRule.TargetField
	TargetFields []LinkFieldInfo `json:"targetFields,omitempty"`
	// SourceFields are the mapped input keys on the ingested record valid as LinkRule.SourceField
	// (scalar types) or LinkRule.SourceList (list types)
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

// IngestPayloadSet groups mapping envelopes by normalized target schema
type IngestPayloadSet struct {
	// Schema is the normalized target schema emitted by the operation
	Schema string `json:"schema"`
	// Envelopes are the raw provider payloads to map and ingest
	Envelopes []MappingEnvelope `json:"envelopes,omitempty"`
}
