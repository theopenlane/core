package workflows

import "errors"

var (
	// ErrWorkflowNotFound is returned when a workflow instance cannot be found
	ErrWorkflowNotFound = errors.New("workflow instance not found")
	// ErrNoProposedChanges is returned when attempting to apply changes but none exist
	ErrNoProposedChanges = errors.New("no proposed changes to apply")
	// ErrWorkflowAlreadyActive is returned when trying to start a workflow on an object that already has one
	ErrWorkflowAlreadyActive = errors.New("workflow already active for this object")
	// ErrFieldNotWorkflowEligible is returned when proposed changes include non-eligible fields
	ErrFieldNotWorkflowEligible = errors.New("field is not eligible for workflow modification")
	// ErrUnsupportedObjectType is returned when a workflow is triggered for an unsupported object type
	ErrUnsupportedObjectType = errors.New("object type does not support workflows")
	// ErrMissingObjectID is returned when a workflow object is missing an ID
	ErrMissingObjectID = errors.New("workflow object is missing an ID")
	// ErrFailedToBuildCELEnv is returned when the CEL environment cannot be built
	ErrFailedToBuildCELEnv = errors.New("failed to build CEL environment")
	// ErrEmitFailureDetailsMissing is returned when emit failure details are empty
	ErrEmitFailureDetailsMissing = errors.New("emit failure details are missing")
	// ErrEmitNoEmitter is returned when an emitter is required but missing
	ErrEmitNoEmitter = errors.New("workflow emit requires an emitter")
	// ErrNilClient is returned when a workflow helper requires a client
	ErrNilClient = errors.New("workflow client is required")
	// ErrMissingOrganizationID is returned when an organization id is required
	ErrMissingOrganizationID = errors.New("organization id is required")
	// ErrApprovalActionParamsInvalid is returned when approval action params are invalid
	ErrApprovalActionParamsInvalid = errors.New("approval action params are invalid")
	// ErrStringFieldMarshal is returned when string field extraction cannot marshal the node
	ErrStringFieldMarshal = errors.New("failed to marshal node for string field extraction")
	// ErrStringFieldUnmarshal is returned when string field extraction cannot unmarshal the node
	ErrStringFieldUnmarshal = errors.New("failed to unmarshal node for string field extraction")
	// ErrStringFieldNil is returned when string field extraction receives a nil node
	ErrStringFieldNil = errors.New("string field node is nil")
	// ErrStringSliceFieldInvalid is returned when a string slice field contains non-string values
	ErrStringSliceFieldInvalid = errors.New("string slice field contains non-string values")
)
