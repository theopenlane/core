package engine

import "errors"

var (
	// ErrInstanceNotFound is returned when a workflow instance cannot be found
	ErrInstanceNotFound = errors.New("workflow instance not found")
	// ErrScopeExpressionRequired is returned when a scope expression is required but empty
	ErrScopeExpressionRequired = errors.New("scope expression is required")
	// ErrAssignmentNotFound is returned when a workflow assignment cannot be found
	ErrAssignmentNotFound = errors.New("workflow assignment not found")
	// ErrInvalidState is returned when a workflow instance is in an invalid state for the operation
	ErrInvalidState = errors.New("invalid workflow state")
	// ErrConditionFailed is returned when a CEL condition evaluation fails
	ErrConditionFailed = errors.New("condition evaluation failed")
	// ErrActionExecutionFailed is returned when an action execution fails
	ErrActionExecutionFailed = errors.New("action execution failed")
	// ErrInvalidTargetType is returned when an unknown target type is encountered
	ErrInvalidTargetType = errors.New("invalid target type")
	// ErrInvalidActionType is returned when an unknown action type is encountered
	ErrInvalidActionType = errors.New("invalid action type")
	// ErrMissingRequiredField is returned when a required field is missing
	ErrMissingRequiredField = errors.New("missing required field")
	// ErrObjectRefMissingID is returned when a workflow object ref has no object ID set
	ErrObjectRefMissingID = errors.New("workflow object ref has no object ID set")
	// ErrCELTypeMismatch is returned when a CEL expression returns a non-boolean type
	ErrCELTypeMismatch = errors.New("CEL expression must return boolean")
	// ErrEvaluationTimeout is returned when CEL evaluation exceeds the timeout
	ErrEvaluationTimeout = errors.New("CEL evaluation timeout")
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
	// ErrExecutorNotAvailable is returned when the executor is not available
	ErrExecutorNotAvailable = errors.New("executor is nil")
	// ErrIntegrationRegistryRequired is returned when an integration registry dependency is missing
	ErrIntegrationRegistryRequired = errors.New("integration registry required")
	// ErrIntegrationOperationsRequired is returned when integration dispatcher is not configured
	ErrIntegrationOperationsRequired = errors.New("integration dispatcher required")
// ErrIntegrationOperationCriteriaRequired indicates operation name is required
	ErrIntegrationOperationCriteriaRequired = errors.New("integration operation name required")
	// ErrIntegrationScopeConditionFalse indicates integration scope evaluation rejected the operation
	ErrIntegrationScopeConditionFalse = errors.New("integration scope condition false")
	// ErrObjectNil is returned when the workflow object is nil
	ErrObjectNil = errors.New("object is nil")
	// ErrUnmarshalParams is returned when action params cannot be unmarshaled
	ErrUnmarshalParams = errors.New("failed to unmarshal action params")
	// ErrMarshalPayload is returned when a payload cannot be marshaled
	ErrMarshalPayload = errors.New("failed to marshal payload")
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
	// ErrNotificationChannelTargetChannelRequired is returned when a CHANNEL target omits channel
	ErrNotificationChannelTargetChannelRequired = errors.New("notification channel target requires channel")
	// ErrNotificationChannelTargetDestinationRequired is returned when a CHANNEL target omits destination
	ErrNotificationChannelTargetDestinationRequired = errors.New("notification channel target requires destination")
	// ErrWebhookURLRequired is returned when webhook action is missing URL
	ErrWebhookURLRequired = errors.New("webhook action requires url")
	// ErrAssignmentUpdateFailed is returned when assignment update fails
	ErrAssignmentUpdateFailed = errors.New("failed to update workflow assignment")
	// ErrActionIndexOutOfBounds is returned when an action index is outside the workflow definition range
	ErrActionIndexOutOfBounds = errors.New("workflow action index out of bounds")
	// ErrAssignmentActionNotFound is returned when a workflow assignment cannot be mapped to an action
	ErrAssignmentActionNotFound = errors.New("workflow assignment action not found")
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
	// ErrTemplateRenderDepthExceeded is returned when template rendering exceeds the maximum depth
	ErrTemplateRenderDepthExceeded = errors.New("template render depth exceeded")
	// ErrNotificationTemplateBlocksInvalid is returned when rendered template blocks are not a valid block list
	ErrNotificationTemplateBlocksInvalid = errors.New("notification template blocks invalid")
)
