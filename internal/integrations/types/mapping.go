package types

import "github.com/theopenlane/core/internal/integrations/schema"

// MappingSchema identifies one normalized ingest schema
type MappingSchema string

// MappingOverride is one mapping customization; canonical definition is in the schema package
type MappingOverride = schema.MappingOverride

// MappingRegistration declares one default mapping shipped with a definition
type MappingRegistration struct {
	// Schema is the normalized target schema for the mapping
	Schema MappingSchema `json:"schema"`
	// Variant is the optional variant name within the schema
	Variant string `json:"variant,omitempty"`
	// Spec contains the mapping expressions for the schema and variant
	Spec MappingOverride `json:"spec"`
}
