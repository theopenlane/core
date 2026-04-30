package models

import "io"

// TemplateProjectionTarget is the destination object type for projected document data.
type TemplateProjectionTarget string

const (
	// TemplateProjectionTargetEntity projects document data into an entity record.
	TemplateProjectionTargetEntity TemplateProjectionTarget = "Entity"
	// TemplateProjectionTargetAsset projects document data into an asset record.
	TemplateProjectionTargetAsset TemplateProjectionTarget = "Asset"
)

// TemplateProjectionOperation is the persistence operation used by a projection.
type TemplateProjectionOperation string

const (
	// TemplateProjectionOperationCreate creates a new target record.
	TemplateProjectionOperationCreate TemplateProjectionOperation = "create"
	// TemplateProjectionOperationUpdate updates an existing target record.
	TemplateProjectionOperationUpdate TemplateProjectionOperation = "update"
	// TemplateProjectionOperationUpsert creates or updates the target record.
	TemplateProjectionOperationUpsert TemplateProjectionOperation = "upsert"
)

// TemplateProjectionTrigger is the event that runs a projection.
type TemplateProjectionTrigger string

const (
	// TemplateProjectionTriggerCompleted runs the projection after the questionnaire is completed.
	TemplateProjectionTriggerCompleted TemplateProjectionTrigger = "completed"
)

// TemplateProjectionTransform converts a source value before it is assigned.
type TemplateProjectionTransform string

const (
	// TemplateProjectionTransformSlugify converts text to a URL-safe slug.
	TemplateProjectionTransformSlugify TemplateProjectionTransform = "slugify"
	// TemplateProjectionTransformDate converts input into the target date/time representation.
	TemplateProjectionTransformDate TemplateProjectionTransform = "date"
	// TemplateProjectionTransformString normalizes input into a string.
	TemplateProjectionTransformString TemplateProjectionTransform = "string"
)

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
