package hooks

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/pkg/gala"
)

// registerMutationListeners compiles mutation listeners to their gala definitions and
// registers them; registration always flows through gala.Register
func registerMutationListeners(g *gala.Gala, listeners ...entityops.MutationListener) ([]gala.ListenerID, error) {
	definitions := make([]gala.Definition[entityops.MutationPayload], 0, len(listeners))

	for _, listener := range listeners {
		definitions = append(definitions, listener.Definition())
	}

	return gala.Register(g, definitions...)
}
