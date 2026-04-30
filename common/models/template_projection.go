package models

import "io"


// TemplateProjectionResolver resolves a source value to another record or field.
type TemplateProjectionResolver string

const (
	// TemplateProjectionResolverInternalOwner resolves an owner value to a user, group, or fallback string.
	TemplateProjectionResolverInternalOwner TemplateProjectionResolver = "internal_owner"
)

// TemplateProjectionRule computes a target value from projection context.
type TemplateProjectionRule string

const (
	// TemplateProjectionRuleActiveIfContractStartedElseUnderReview sets status from contract timing.
	TemplateProjectionRuleActiveIfContractStartedElseUnderReview TemplateProjectionRule = "active_if_contract_started_else_under_review"
)

// TemplateProjectionConfig describes how document data should be projected into a typed schema.
type TemplateProjectionConfig struct {
	// Enabled controls whether projection should run for this template.
	Enabled bool `json:"enabled,omitempty"`
	// Target is the destination object type, e.g. Entity or Asset.
	Target TemplateProjectionTarget `json:"target,omitempty"`
	// Operation is the persistence behavior for the projection.
	Operation TemplateProjectionOperation `json:"operation,omitempty"`
	// Trigger is the questionnaire/document event that runs the projection.
	Trigger TemplateProjectionTrigger `json:"trigger,omitempty"`
	// FieldMappings maps document data fields to target schema fields.
	FieldMappings []TemplateProjectionFieldMapping `json:"fieldMappings,omitempty"`
}

// TemplateProjectionFieldMapping maps one document value or computed rule to a target field.
type TemplateProjectionFieldMapping struct {
	// From is a JSON pointer into the submitted document data.
	From string `json:"from,omitempty"`
	// To is the target schema field name.
	To string `json:"to,omitempty"`
	// Transform converts the source value before assignment.
	Transform TemplateProjectionTransform `json:"transform,omitempty"`
	// Resolver resolves a source value to one or more target fields.
	Resolver TemplateProjectionResolver `json:"resolver,omitempty"`
	// Rule computes a value from projection context instead of directly copying From.
	Rule TemplateProjectionRule `json:"rule,omitempty"`
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
