package hooks

import (
	"entgo.io/ent"
	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaIntegrationCleanupListeners cancels queued integration jobs when an
// installation is removed or leaves the connected state, so scheduled loops and queued
// record jobs stop eagerly instead of each discovering the removal one attempt at a time
func RegisterGalaIntegrationCleanupListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return registerMutationListeners(g,
		entityops.MutationListener{
			Schema: entgen.TypeIntegration,
			Label:  "removed",
			Operations: []string{
				ent.OpDelete.String(),
				ent.OpDeleteOne.String(),
				gala.SoftDeleteOne,
			},
			Handle: handleIntegrationRemoved,
		},
		entityops.MutationListener{
			Schema: entgen.TypeIntegration,
			Label:  "updated",
			Operations: []string{
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Fields: []string{integration.FieldDeletedAt, integration.FieldStatus},
			Handle: handleIntegrationUpdated,
		},
	)
}

// handleIntegrationRemoved cancels queued jobs for a deleted installation
func handleIntegrationRemoved(inv entityops.Invocation, _ entityops.MutationPayload) error {
	return cancelInstallationJobs(inv)
}

// handleIntegrationUpdated cancels queued jobs when an update soft-deletes the installation
// (soft deletes surface as updates setting deleted_at) or moves it out of the connected state
func handleIntegrationUpdated(inv entityops.Invocation, payload entityops.MutationPayload) error {
	if payload.FieldChanged(integration.FieldDeletedAt) {
		return cancelInstallationJobs(inv)
	}

	rawStatus, _ := payload.Value(integration.FieldStatus)

	status, ok := entityops.ParseEnum(
		rawStatus,
		enums.ToIntegrationStatus,
		enums.IntegrationStatusInvalid,
	)
	if !ok || status == enums.IntegrationStatusConnected {
		return nil
	}

	return cancelInstallationJobs(inv)
}

// cancelInstallationJobs cancels every queued River job bound to the mutated installation
func cancelInstallationJobs(inv entityops.Invocation) error {
	rt, err := do.Invoke[*intruntime.Runtime](inv.Injector)
	if err != nil || rt == nil {
		return nil
	}

	cancelled, err := rt.CancelInstallationJobs(inv.Context, inv.EntityID)
	if err != nil {
		return err
	}

	if cancelled > 0 {
		logx.FromContext(inv.Context).Info().Int("cancelled", cancelled).Str("integration_id", inv.EntityID).Msg("cancelled queued integration jobs")
	}

	return nil
}
