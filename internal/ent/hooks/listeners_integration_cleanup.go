package hooks

import (
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// IntegrationCleanupListeners cancels queued integration jobs on removal or disconnect
// and reseeds them on reconnect
func IntegrationCleanupListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaIntegration,
			Operations: []string{entityops.OpSoftDelete, entityops.OpDelete, entityops.OpDeleteOne},
			Handle:     entityops.RequireDep(cancelInstallationJobs),
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaIntegration,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{integration.FieldStatus},
			Handle:     entityops.RequireDep(handleIntegrationUpdated),
		},
	}
}

// handleIntegrationUpdated cancels queued jobs when an update leaves the connected state
// and reseeds recurring loops when it returns to connected
func handleIntegrationUpdated(inv entityops.Invocation, payload entityops.MutationPayload, rt *intruntime.Runtime) error {
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
		return reseedInstallationJobs(inv, rt)
	}

	return cancelInstallationJobs(inv, payload, rt)
}

// reseedInstallationJobs restores recurring loops for a reconnected installation;
// ResetReconcileLoops keeps paths that already seeded collapsed to one loop
func reseedInstallationJobs(inv entityops.Invocation, rt *intruntime.Runtime) error {
	installation, found, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Integration.Get)
	if err != nil || !found {
		return err
	}

	return rt.ResetReconcileLoops(inv.Context, installation)
}

// cancelInstallationJobs cancels every queued River job bound to the mutated installation
func cancelInstallationJobs(inv entityops.Invocation, _ entityops.MutationPayload, rt *intruntime.Runtime) error {
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
// organization owns; failures only log because orphaned loops self-cancel on not-found
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
