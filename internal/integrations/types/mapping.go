package types

// MappingSchema identifies one normalized ingest schema
type MappingSchema string

// MappingOverride is one definition-level mapping customization
type MappingOverride struct {
	// Version is the schema version for the mapping
	Version string `json:"version,omitempty"`
	// FilterExpr is the CEL expression used to filter source records
	FilterExpr string `json:"filterExpr,omitempty"`
	// MapExpr is the CEL expression used to map source records into target shape
	MapExpr string `json:"mapExpr,omitempty"`
}

// MappingRegistration declares one default mapping shipped with a definition
type MappingRegistration struct {
	// Schema is the normalized target schema for the mapping
	Schema MappingSchema `json:"schema"`
	// Variant is the optional variant name within the schema
	Variant string `json:"variant,omitempty"`
	// Spec contains the mapping expressions for the schema and variant
	Spec MappingOverride `json:"spec"`
}
