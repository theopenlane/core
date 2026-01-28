package graphapi

import "errors"

var (
	// ErrWorkflowsDisabled is returned when workflows are disabled
	ErrWorkflowsDisabled = errors.New("workflows are disabled")
	// ErrSchemaTypeMismatch is returned when definition schemaType does not match input schemaType
	ErrSchemaTypeMismatch = errors.New("schema type mismatch")
	// ErrNoTriggers is returned when workflow definition has no triggers
	ErrNoTriggers = errors.New("workflow definition must include at least one trigger")
	// ErrTriggerMissingOperation is returned when a trigger is missing an operation
	ErrTriggerMissingOperation = errors.New("trigger missing operation")
	// ErrTriggerUnsupportedOperation is returned when a trigger has an unsupported operation
	ErrTriggerUnsupportedOperation = errors.New("trigger has unsupported operation")
	// ErrTriggerMissingObjectType is returned when a trigger is missing an objectType
	ErrTriggerMissingObjectType = errors.New("trigger missing objectType")
	// ErrTriggerObjectTypeMismatch is returned when trigger objectType does not match schemaType
	ErrTriggerObjectTypeMismatch = errors.New("trigger objectType does not match schemaType")
	// ErrTriggerEmptyFieldName is returned when a trigger has an empty field name
	ErrTriggerEmptyFieldName = errors.New("trigger has empty field name")
	// ErrTriggerEmptyEdgeName is returned when a trigger has an empty edge name
	ErrTriggerEmptyEdgeName = errors.New("trigger has empty edge name")
	// ErrNoActions is returned when workflow definition has no actions
	ErrNoActions = errors.New("workflow definition must include at least one action")
	// ErrActionMissingKey is returned when an action is missing a key
	ErrActionMissingKey = errors.New("action missing key")
	// ErrActionDuplicateKey is returned when an action has a duplicate key
	ErrActionDuplicateKey = errors.New("duplicate action key")
	// ErrActionMissingType is returned when an action is missing a type
	ErrActionMissingType = errors.New("action missing type")
	// ErrActionUnsupportedType is returned when an action has an unsupported type
	ErrActionUnsupportedType = errors.New("action has unsupported type")
	// ErrActionInvalidParams is returned when action params are invalid
	ErrActionInvalidParams = errors.New("invalid action params")
	// ErrDuplicateApprovalFieldSet is returned when multiple approval actions target the same field set
	ErrDuplicateApprovalFieldSet = errors.New("duplicate approval actions for field set")
	// ErrConflictingApprovalDomain is returned when multiple definitions target the same approval domain
	ErrConflictingApprovalDomain = errors.New("conflicting approval domain for schema type")
	// ErrInvalidCELExpression is returned when a CEL expression fails to compile
	ErrInvalidCELExpression = errors.New("invalid CEL expression")
	// ErrDefinitionRequired is returned when definition_json is required but not provided
	ErrDefinitionRequired = errors.New("definition_json is required")
	// ErrApprovalParamsRequired is returned when approval params are required but not provided
	ErrApprovalParamsRequired = errors.New("approval params required")
	// ErrApprovalFieldRequired is returned when at least one approval field or edge is required
	ErrApprovalFieldRequired = errors.New("approval action must specify at least one field or edge")
	// ErrApprovalEdgesNotSupported is returned when approval edges are used but not supported
	ErrApprovalEdgesNotSupported = errors.New("approval action edges are not supported")
	// ErrApprovalFieldNotEligible is returned when an approval field is not workflow-eligible
	ErrApprovalFieldNotEligible = errors.New("approval field is not workflow-eligible")
	// ErrApprovalTargetsRequired is returned when approval targets or assignees are required
	ErrApprovalTargetsRequired = errors.New("approval action requires targets or assignees")
	// ErrRequiredCountNegative is returned when required_count is negative
	ErrRequiredCountNegative = errors.New("required_count must be non-negative")
	// ErrRequiredInvalid is returned when required field has an invalid value
	ErrRequiredInvalid = errors.New("required must be a bool or non-negative integer")
	// ErrTargetMissingID is returned when a target is missing required ID
	ErrTargetMissingID = errors.New("target requires id")
	// ErrTargetInvalidRole is returned when a role target has an invalid role
	ErrTargetInvalidRole = errors.New("unknown role")
	// ErrTargetMissingResolverKey is returned when a resolver target is missing resolver_key
	ErrTargetMissingResolverKey = errors.New("resolver target requires resolver_key")
	// ErrTargetUnknownResolver is returned when a resolver key is unknown
	ErrTargetUnknownResolver = errors.New("unknown resolver key")
	// ErrTargetInvalidType is returned when a target has an invalid type
	ErrTargetInvalidType = errors.New("invalid target type")
	// ErrWebhookParamsRequired is returned when webhook params are required but not provided
	ErrWebhookParamsRequired = errors.New("webhook params required")
	// ErrWebhookURLRequired is returned when webhook URL is required but not provided
	ErrWebhookURLRequired = errors.New("webhook action requires url")
	// ErrWebhookURLInvalid is returned when webhook URL is invalid
	ErrWebhookURLInvalid = errors.New("invalid webhook url")
	// ErrFieldUpdateParamsRequired is returned when field update params are required but not provided
	ErrFieldUpdateParamsRequired = errors.New("field update params required")
	// ErrFieldUpdateUpdatesRequired is returned when field update updates are required
	ErrFieldUpdateUpdatesRequired = errors.New("field_update action requires updates")
	// ErrIntegrationParamsRequired is returned when integration params are required but not provided
	ErrIntegrationParamsRequired = errors.New("integration params required")
	// ErrIntegrationConfigRequired is returned when integration config is incomplete
	ErrIntegrationConfigRequired = errors.New("integration action requires integration or {provider, operation}")
	// ErrApprovalSubmissionModeInvalid is returned when approval submission mode is invalid
	ErrApprovalSubmissionModeInvalid = errors.New("invalid approval submission mode")
	// ErrManualSubmitModeNotSupported is returned when MANUAL_SUBMIT mode is specified but not yet supported
	ErrManualSubmitModeNotSupported = errors.New("manual submit mode is not yet supported; use AUTO_SUBMIT or omit the field")
	// ErrFailedToQueryDefinitions is returned when workflow definitions cannot be queried
	ErrFailedToQueryDefinitions = errors.New("failed to query workflow definitions")
	// ErrInvalidWorkflowSchema is returned when a workflow schema is invalid
	ErrInvalidWorkflowSchema = errors.New("invalid workflow schema")
)
