package hooks

import (
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// IntegrationCleanupListeners returns the listeners that cancel queued integration jobs
// when an installation is removed or leaves the connected state
func IntegrationCleanupListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entgen.TypeIntegration,
			Label:      "removed",
			Operations: []string{entityops.OpSoftDelete},
			Handle:     cancelInstallationJobs,
		},
		entityops.MutationListener{
			Schema:     entgen.TypeIntegration,
			Label:      "updated",
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{integration.FieldStatus},
			Handle:     handleIntegrationUpdated,
		},
	}
}

// handleIntegrationUpdated cancels queued jobs when an update moves the installation out of the connected state
func handleIntegrationUpdated(inv entityops.Invocation, payload entityops.MutationPayload) error {
	rawStatus, _ := payload.Value(integration.FieldStatus)

	status, ok := entityops.ParseEnum(
		rawStatus,
		enums.ToIntegrationStatus,
		enums.IntegrationStatusInvalid,
	)
	if !ok || status == enums.IntegrationStatusConnected {
		return nil
	}

	return cancelInstallationJobs(inv, payload)
}

// cancelInstallationJobs cancels every queued River job bound to the mutated installation
func cancelInstallationJobs(inv entityops.Invocation, _ entityops.MutationPayload) error {
	rt, ok := gala.Resolve[*intruntime.Runtime](inv.Context, inv.Injector, "integration_cleanup")
	if !ok {
		return nil
	}

	cancelled, err := rt.CancelInstallationJobs(inv.Context, inv.EntityID)
	if err != nil {
		return err
	}

	if cancelled > 0 {
		logx.FromContext(inv.Context).Info().Int("cancelled", cancelled).Msg("cancelled queued integration jobs")
	}

	return nil
}
