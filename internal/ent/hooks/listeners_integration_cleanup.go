package hooks

import (
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// IntegrationCleanupListeners returns the listeners that cancel queued integration jobs
// when an installation is removed or leaves connected, and reseed them on reconnect
func IntegrationCleanupListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaIntegration,
			Label:      "removed",
			Operations: []string{entityops.OpSoftDelete, entityops.OpDelete, entityops.OpDeleteOne},
			Handle:     cancelInstallationJobs,
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaIntegration,
			Label:      "updated",
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{integration.FieldStatus},
			Handle:     handleIntegrationUpdated,
		},
	}
}

// handleIntegrationUpdated cancels queued jobs when an update leaves the connected state
// and reseeds recurring loops when it returns to connected
func handleIntegrationUpdated(inv entityops.Invocation, payload entityops.MutationPayload) error {
	rawStatus, _ := payload.Value(integration.FieldStatus)

	status, ok := entityops.ParseEnum(
		rawStatus,
		enums.ToIntegrationStatus,
		enums.IntegrationStatusInvalid,
	)
	if !ok {
		return nil
	}

	if status == enums.IntegrationStatusConnected {
		return reseedInstallationJobs(inv)
	}

	return cancelInstallationJobs(inv, payload)
}

// reseedInstallationJobs restores recurring loops for a reconnected installation;
// ResetReconcileLoops keeps paths that already seeded collapsed to one loop
func reseedInstallationJobs(inv entityops.Invocation) error {
	rt, ok := gala.Resolve[*intruntime.Runtime](inv.Context, inv.Injector, "integration_cleanup")
	if !ok {
		return nil
	}

	installation, found, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Integration.Get)
	if err != nil || !found {
		return err
	}

	return rt.ResetReconcileLoops(inv.Context, installation)
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

// cancelOrganizationIntegrationJobs cancels queued jobs for every installation the
// organization owns; the cascade's vetoed hard deletes never reach the per-integration
// listeners, and failures only log because orphaned loops self-cancel on not-found
func cancelOrganizationIntegrationJobs(inv entityops.Invocation) {
	rt, ok := gala.Resolve[*intruntime.Runtime](inv.Context, inv.Injector, "organization_cleanup")
	if !ok {
		return
	}

	ids, err := inv.Client.Integration.Query().Where(integration.OwnerID(inv.EntityID)).IDs(inv.Context)
	if err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed listing organization integrations for job cancellation")

		return
	}

	for _, id := range ids {
		if _, err := rt.CancelInstallationJobs(inv.Context, id); err != nil {
			logx.FromContext(inv.Context).Error().Err(err).Str("integration_id", id).Msg("failed cancelling queued integration jobs")
		}
	}
}
