package hooks

import (
	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
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
	return eventqueue.RegisterMutationListeners(g,
		eventqueue.MutationListener{
			Schema: entgen.TypeIntegration,
			Name:   "integrations.job_cleanup_removed",
			Operations: []string{
				ent.OpDelete.String(),
				ent.OpDeleteOne.String(),
				eventqueue.SoftDeleteOne,
			},
			Handle: handleIntegrationRemoved,
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeIntegration,
			Name:   "integrations.job_cleanup_updated",
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
func handleIntegrationRemoved(inv eventqueue.Invocation, _ eventqueue.MutationGalaPayload) error {
	return cancelInstallationJobs(inv)
}

// handleIntegrationUpdated cancels queued jobs when an update soft-deletes the installation
// (soft deletes surface as updates setting deleted_at) or moves it out of the connected state
func handleIntegrationUpdated(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	if eventqueue.MutationFieldChanged(payload, integration.FieldDeletedAt) {
		return cancelInstallationJobs(inv)
	}

	status, ok := eventqueue.ParseEnum(
		payload.ProposedChanges[integration.FieldStatus],
		enums.ToIntegrationStatus,
		enums.IntegrationStatusInvalid,
	)
	if !ok || status == enums.IntegrationStatusConnected {
		return nil
	}

	return cancelInstallationJobs(inv)
}

// cancelInstallationJobs cancels every queued River job bound to the mutated installation
func cancelInstallationJobs(inv eventqueue.Invocation) error {
	rt := intruntime.FromClient(inv.Context, inv.Client)
	if rt == nil {
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
