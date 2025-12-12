package events

import (
	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
)

// MutationPayload carries the raw ent mutation, the resolved operation, the entity ID and the ent
// client so listeners can act without additional lookups
type MutationPayload struct {
	Mutation  ent.Mutation
	Operation string
	EntityID  string
	Client    *generated.Client
}
