package events

import (
	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
)

// MutationPayload carries the raw ent mutation, the resolved operation, the entity ID and the ent
// client so listeners can act without additional lookups
type MutationPayload struct {
	// Mutation is the raw ent mutation that triggered the event
	Mutation ent.Mutation
	// MutationType is the ent schema type that emitted the mutation
	MutationType string
	// Operation is the string representation of the mutation operation
	Operation string
	// EntityID is the ID of the entity that was mutated
	EntityID string
	// ChangedFields captures updated/cleared fields for the mutation
	ChangedFields []string
	// ClearedFields captures fields explicitly cleared in the mutation
	ClearedFields []string
	// ChangedEdges captures changed edge names for workflow-eligible edges
	ChangedEdges []string
	// AddedIDs captures edge IDs added by edge name
	AddedIDs map[string][]string
	// RemovedIDs captures edge IDs removed by edge name
	RemovedIDs map[string][]string
	// ProposedChanges captures field-level proposed values (including nil for clears)
	ProposedChanges map[string]any
	// Client is the ent client that can be used to perform additional queries or mutations
	Client *generated.Client
}
