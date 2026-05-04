package models

import (
	"io"

	"github.com/theopenlane/core/common/enums"
)

// TemplateProjectionResolver resolves a source value to another record or field.
type TemplateProjectionResolver string

const (
	// TemplateProjectionResolverInternalOwner resolves an owner value to a user, group, or fallback string.
	TemplateProjectionResolverInternalOwner TemplateProjectionResolver = "internal_owner"
)

// TemplateProjectionConfig describes how document data should be projected into a typed schema.
type TemplateProjectionConfig struct {
	// Enabled controls whether projection should run for this template.
	Enabled bool `json:"enabled,omitempty"`
	// Target is the destination object type, e.g. Entity or Asset.
	Target enums.TemplateProjectionTarget `json:"target,omitempty"`
	// Operation is the persistence behavior for the projection.
	Operation enums.TemplateProjectionOperation `json:"operation,omitempty"`
	// Mappings maps document data fields to target schema fields.
	Mappings []TemplateProjectionFieldMapping `json:"mappings,omitempty"`
}

// TemplateProjectionFieldMapping maps one document value to a target field.
type TemplateProjectionFieldMapping struct {
	// From is a submitted document data field path, e.g. "vendorName" or "vendor.name".
	From string `json:"from,omitempty"`
	// To is the target schema field name.
	To string `json:"to,omitempty"`
	// Transform converts the source value before assignment.
	Transform enums.TemplateProjectionTransform `json:"transform,omitempty"`
	// Resolver resolves a source value to one or more target fields.
	Resolver TemplateProjectionResolver `json:"resolver,omitempty"`
	// Required marks the source value as required before projection can run.
	Required bool `json:"required,omitempty"`
}

// MarshalGQL implements the Marshaler interface for gqlgen.
func (c TemplateProjectionConfig) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, c)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen.
func (c *TemplateProjectionConfig) UnmarshalGQL(v interface{}) error {
	return unmarshalGQLJSON(v, c)
}
