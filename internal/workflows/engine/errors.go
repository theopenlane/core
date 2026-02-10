package engine

import "errors"

var (
	// ErrWorkflowNotFound is returned when a workflow definition cannot be found
	ErrWorkflowNotFound = errors.New("workflow definition not found")
	// ErrInstanceNotFound is returned when a workflow instance cannot be found
	ErrInstanceNotFound = errors.New("workflow instance not found")
	// ErrAssignmentNotFound is returned when a workflow assignment cannot be found
	ErrAssignmentNotFound = errors.New("workflow assignment not found")
	// ErrInvalidState is returned when a workflow instance is in an invalid state for the operation
	ErrInvalidState = errors.New("invalid workflow state")
	// ErrConditionFailed is returned when a CEL condition evaluation fails
	ErrConditionFailed = errors.New("condition evaluation failed")
	// ErrNoMatchingDefinitions is returned when no workflow definitions match the trigger criteria
	ErrNoMatchingDefinitions = errors.New("no matching workflow definitions")
	// ErrTargetResolutionFailed is returned when dynamic target resolution fails
	ErrTargetResolutionFailed = errors.New("target resolution failed")
	// ErrActionExecutionFailed is returned when an action execution fails
	ErrActionExecutionFailed = errors.New("action execution failed")
	// ErrInvalidTargetType is returned when an unknown target type is encountered
	ErrInvalidTargetType = errors.New("invalid target type")
	// ErrInvalidActionType is returned when an unknown action type is encountered
	ErrInvalidActionType = errors.New("invalid action type")
	// ErrMissingRequiredField is returned when a required field is missing
	ErrMissingRequiredField = errors.New("missing required field")
	// ErrInvalidObjectType is returned when an unknown object type is encountered
	ErrInvalidObjectType = errors.New("invalid object type")
	// ErrObjectRefMissingID is returned when a workflow object ref has no object ID set
	ErrObjectRefMissingID = errors.New("workflow object ref has no object ID set")
	// ErrFieldNotWorkflowEligible is returned when a field cannot be updated by workflow
	ErrFieldNotWorkflowEligible = errors.New("field is not eligible for workflow modification")
	// ErrCELTypeMismatch is returned when a CEL expression returns a non-boolean type
	ErrCELTypeMismatch = errors.New("CEL expression must return boolean")
	// ErrCELValueExtraction is returned when extracting a boolean value from CEL fails
	ErrCELValueExtraction = errors.New("failed to extract boolean value from CEL result")
	// ErrEvaluationTimeout is returned when CEL evaluation exceeds the timeout
	ErrEvaluationTimeout = errors.New("CEL evaluation timeout")
	// ErrCELPanic is returned when CEL evaluation panics
	ErrCELPanic = errors.New("CEL evaluation panic")
	// ErrCELNilOutput is returned when CEL evaluation returns nil output
	ErrCELNilOutput = errors.New("CEL evaluation returned nil output")
	// ErrWebhookFailed is returned when a webhook request fails
	ErrWebhookFailed = errors.New("webhook request failed")
	// ErrWebhookPayloadUnsupported is returned when legacy webhook payloads are provided
	ErrWebhookPayloadUnsupported = errors.New("webhook payload is not supported; use payload_expr")
	// ErrWebhookPayloadExpressionFailed is returned when webhook payload CEL evaluation fails
	ErrWebhookPayloadExpressionFailed = errors.New("webhook payload expression failed")
	// ErrWebhookPayloadExpressionInvalid is returned when webhook payload CEL does not produce a JSON object
	ErrWebhookPayloadExpressionInvalid = errors.New("webhook payload expression invalid")
	// ErrIntegrationFailed is returned when an integration operation fails
	ErrIntegrationFailed = errors.New("integration operation failed")
	// ErrExecutorNotAvailable is returned when the executor is not available
	ErrExecutorNotAvailable = errors.New("executor is nil")
	// ErrIntegrationManagerNotAvailable is returned when the integration manager is not available
	ErrIntegrationManagerNotAvailable = errors.New("integration operations manager not available")
	// ErrIntegrationStoreRequired is returned when an integration store dependency is missing
	ErrIntegrationStoreRequired = errors.New("integration store required")
	// ErrIntegrationOperationsRequired is returned when integration operations are not configured
	ErrIntegrationOperationsRequired = errors.New("integration operations required")
	// ErrIntegrationEmitterRequired is returned when integration event emitter is missing
	ErrIntegrationEmitterRequired = errors.New("integration emitter required")
	// ErrIntegrationRunIDRequired indicates the integration run identifier is missing
	ErrIntegrationRunIDRequired = errors.New("integration run id required")
	// ErrIntegrationRecordMissing indicates the integration record is missing for a run
	ErrIntegrationRecordMissing = errors.New("integration record missing for run")
	// ErrIntegrationProviderUnknown indicates the integration provider could not be resolved
	ErrIntegrationProviderUnknown = errors.New("integration provider unknown")
	// ErrIntegrationOperationNameRequired indicates the run operation name is missing
	ErrIntegrationOperationNameRequired = errors.New("integration operation name required")
	// ErrIntegrationOperationFailed indicates the operation failed to execute successfully
	ErrIntegrationOperationFailed = errors.New("integration operation failed")
	// ErrIntegrationAlertPayloadsMissing indicates alert payloads are missing from operation output
	ErrIntegrationAlertPayloadsMissing = errors.New("integration alert payloads missing")
	// ErrIntegrationActionQueued indicates the integration action was queued for async processing
	ErrIntegrationActionQueued = errors.New("integration action queued")
	// ErrUnsupportedTimeFormat is returned when a time format is not supported
	ErrUnsupportedTimeFormat = errors.New("unsupported time format")
	// ErrObjectNil is returned when the workflow object is nil
	ErrObjectNil = errors.New("object is nil")
	// ErrProposalChangesModified is returned when proposal changes are modified after approval
	ErrProposalChangesModified = errors.New("proposal changes modified after approval")
	// ErrUnmarshalParams is returned when action params cannot be unmarshaled
	ErrUnmarshalParams = errors.New("failed to unmarshal action params")
	// ErrMarshalPayload is returned when a payload cannot be marshaled
	ErrMarshalPayload = errors.New("failed to marshal payload")
	// ErrIntegrationProviderRequired is returned when integration action is missing provider
	ErrIntegrationProviderRequired = errors.New("integration action requires provider")
	// ErrIntegrationOwnerRequired is returned when integration action is missing owner
	ErrIntegrationOwnerRequired = errors.New("integration action requires instance owner_id")
	// ErrAssignmentCreationFailed is returned when workflow assignment creation fails
	ErrAssignmentCreationFailed = errors.New("failed to create workflow assignment")
	// ErrNotificationCreationFailed is returned when notification creation fails
	ErrNotificationCreationFailed = errors.New("failed to create notification")
	// ErrNotificationTemplateNotFound is returned when a notification template cannot be found
	ErrNotificationTemplateNotFound = errors.New("notification template not found")
	// ErrNotificationTemplateReferenceConflict is returned when both template_id and template_key are provided
	ErrNotificationTemplateReferenceConflict = errors.New("notification template reference conflict")
	// ErrNotificationTemplateDataInvalid is returned when template data fails schema validation
	ErrNotificationTemplateDataInvalid = errors.New("notification template data invalid")
	// ErrNotificationTemplateChannelMismatch is returned when template channel does not match requested channel
	ErrNotificationTemplateChannelMismatch = errors.New("notification template channel mismatch")
	// ErrNotificationChannelUnsupported is returned when a notification channel lacks integration support
	ErrNotificationChannelUnsupported = errors.New("notification channel unsupported")
	// ErrWebhookURLRequired is returned when webhook action is missing URL
	ErrWebhookURLRequired = errors.New("webhook action requires url")
	// ErrAssignmentUpdateFailed is returned when assignment update fails
	ErrAssignmentUpdateFailed = errors.New("failed to update workflow assignment")
	// ErrActionIndexOutOfBounds is returned when an action index is outside the workflow definition range
	ErrActionIndexOutOfBounds = errors.New("workflow action index out of bounds")
	// ErrAssignmentActionNotFound is returned when a workflow assignment cannot be mapped to an action
	ErrAssignmentActionNotFound = errors.New("workflow assignment action not found")
	// ErrNilClient is returned when a nil database client is provided
	ErrNilClient = errors.New("ent client is required to initialize workflow engine")
	// ErrCELProgramCreationFailed is returned when CEL program creation fails
	ErrCELProgramCreationFailed = errors.New("failed to create CEL program")
	// ErrCELCompilationFailed is returned when CEL fails to compile
	ErrCELCompilationFailed = errors.New("failed to compile CEL")
	// ErrFailedToComputeProposalHash is returned when a proposal hash cannot be computed
	ErrFailedToComputeProposalHash = errors.New("failed to compute proposal hash")
	// ErrFailedToCreateProposal is returned when a workflow proposal cannot be created
	ErrFailedToCreateProposal = errors.New("failed to create workflow proposal inside of proposal manager")
	// ErrFailedToQueryObjectRefs is returned when workflow object refs cannot be queried
	ErrFailedToQueryObjectRefs = errors.New("failed to query object refs")
	// ErrFailedToQueryProposals is returned when workflow proposals cannot be queried
	ErrFailedToQueryProposals = errors.New("failed to query proposals")
	// ErrFailedToLoadProposal is returned when a workflow proposal cannot be loaded
	ErrFailedToLoadProposal = errors.New("failed to load proposal")
	// ErrFailedToApplyFieldUpdates is returned when proposal field updates cannot be applied
	ErrFailedToApplyFieldUpdates = errors.New("failed to apply field updates")
	// ErrFailedToCreateAssignmentTarget is returned when an assignment target cannot be created
	ErrFailedToCreateAssignmentTarget = errors.New("failed to create assignment target")
	// ErrFailedToEnrichWebhookPayload is returned when webhook payload enrichment fails
	ErrFailedToEnrichWebhookPayload = errors.New("failed to enrich webhook payload")
	// ErrFailedToQueryDefinitions is returned when workflow definitions cannot be queried
	ErrFailedToQueryDefinitions = errors.New("failed to query workflow definitions")
	// ErrFailedToResolveTarget is returned when a target cannot be resolved
	ErrFailedToResolveTarget = errors.New("failed to resolve target")
	// ErrFailedToResolveNotificationTarget is returned when a notification target cannot be resolved
	ErrFailedToResolveNotificationTarget = errors.New("failed to resolve notification target")
	// ErrApprovalNoTargets indicates an approval action resolved no targets and should be skipped
	ErrApprovalNoTargets = errors.New("approval action has no resolved targets")
	// ErrMissingObjectRef is returned when a workflow object ref is nil
	ErrMissingObjectRef = errors.New("workflow object ref is required")
	// ErrReviewNoTargets indicates a review action resolved no targets and should be skipped
	ErrReviewNoTargets = errors.New("review action has no resolved targets")
)
