package runtime

import (
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveDefinitionForInstallation resolves the definition for one installation, returning sentinels for nil installation or missing definition
func (r *Runtime) resolveDefinitionForInstallation(installation *ent.Integration) (types.Definition, error) {
	if installation == nil {
		return types.Definition{}, ErrInstallationRequired
	}

	def, ok := r.Registry().Definition(installation.DefinitionID)
	if !ok {
		return types.Definition{}, registry.ErrDefinitionNotFound
	}

	return def, nil
}

// ResolvePersistedConnection resolves the active connection mode for one installation
func (r *Runtime) ResolvePersistedConnection(installation *ent.Integration) (types.ConnectionRegistration, error) {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return types.ConnectionRegistration{}, err
	}

	return r.resolvePersistedConnection(def, installation)
}
