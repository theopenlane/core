package ingest

import (
	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// Runtime wires ingest dependencies that must be injected explicitly (mapping index, listeners).
type Runtime struct {
	mappingIndex integrationtypes.MappingIndex
}

// NewRuntime constructs an ingest runtime with an explicit mapping index source.
func NewRuntime(mappingIndex integrationtypes.MappingIndex) *Runtime {
	return &Runtime{mappingIndex: mappingIndex}
}

// MappingIndex returns the runtime mapping index dependency.
func (r *Runtime) MappingIndex() integrationtypes.MappingIndex {
	if r == nil {
		return nil
	}

	return r.mappingIndex
}

// RegisterListeners registers ingest listeners using this runtime's mapping index dependency.
func (r *Runtime) RegisterListeners(registry *gala.Registry, db *ent.Client) ([]gala.ListenerID, error) {
	return RegisterIngestListeners(registry, db, r.MappingIndex())
}
