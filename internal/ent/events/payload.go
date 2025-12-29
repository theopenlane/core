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
	// Operation is the string representation of the mutation operation
	Operation string
	// EntityID is the ID of the entity that was mutated
	EntityID string
	// Client is the ent client that can be used to perform additional queries or mutations
	Client *generated.Client
}
