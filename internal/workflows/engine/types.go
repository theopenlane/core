package engine

// TriggerInput captures the trigger metadata passed to workflow execution
type TriggerInput struct {
	// EventType is the trigger event name
	EventType string
	// ChangedFields lists updated fields on the target object
	ChangedFields []string
	// ChangedEdges lists updated edges on the target object
	ChangedEdges []string
	// AddedIDs captures added edge IDs keyed by edge name
	AddedIDs map[string][]string
	// RemovedIDs captures removed edge IDs keyed by edge name
	RemovedIDs map[string][]string
	// ProposedChanges contains proposed field updates for approval workflows
	ProposedChanges map[string]any
}
