package observability

// OperationName identifies a workflow operation name.
type OperationName string

// Origin identifies the component emitting the observation.
type Origin string

const (
	// OriginEngine identifies workflow engine operations.
	OriginEngine Origin = "engine"
	// OriginListeners identifies workflow listener operations.
	OriginListeners Origin = "listeners"
	// OriginResolver identifies workflow resolver operations.
	OriginResolver Origin = "resolver"
)

const (
	// OpTriggerWorkflow identifies trigger workflow operations.
	OpTriggerWorkflow OperationName = "trigger_workflow"
	// OpTriggerExistingInstance identifies trigger existing instance operations.
	OpTriggerExistingInstance OperationName = "trigger_existing_instance"
	// OpCompleteAssignment identifies assignment completion processing.
	OpCompleteAssignment OperationName = "complete_assignment"
	// OpFindMatchingDefinitions identifies definition matching.
	OpFindMatchingDefinitions OperationName = "find_matching_definitions"
	// OpResolveTargets identifies target resolution.
	OpResolveTargets OperationName = "resolve_targets"
	// OpHandleAssignmentCompleted identifies assignment completion handling.
	OpHandleAssignmentCompleted OperationName = "handle_assignment_completed"
	// OpExecuteAction identifies action execution.
	OpExecuteAction OperationName = "execute_action"
)
