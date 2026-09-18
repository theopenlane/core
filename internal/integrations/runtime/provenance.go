package runtime

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// BackfillInstallationProvenance stamps every ingest schema's fill-only provenance columns for one
// installation whose instance id is already resolved, reporting the rows stamped and skipping an
// installation that still has no instance id
func (r *Runtime) BackfillInstallationProvenance(ctx context.Context, installation *ent.Integration) (int, error) {
	if installation.InstallationMetadata.Display.ExternalID == "" {
		return 0, nil
	}

	orgCtx := entityops.WithEmissionVetoed(auth.EnsureIntegrationCaller(privacy.DecisionContext(ctx, privacy.Allow), installation.OwnerID))
	instCtx := intobvs.WithInstallation(orgCtx, installation)

	return stampInstallationProvenance(instCtx, r.DB(), installation)
}

// BackfillInstallationInstanceID resolves and overwrites one installation's instance id from its
// persisted credential during the startup backfill, so provenance stamps the current external
// instance and a later reinstall correlates records on it
func (r *Runtime) BackfillInstallationInstanceID(ctx context.Context, installation *ent.Integration) error {
	orgCtx := entityops.WithEmissionVetoed(auth.EnsureIntegrationCaller(privacy.DecisionContext(ctx, privacy.Allow), installation.OwnerID))
	instCtx := intobvs.WithInstallation(orgCtx, installation)

	return r.RefreshInstallationMetadata(instCtx, installation)
}

// stampInstallationProvenance runs every provenance stamp against one installation's records,
// running every stamp regardless of earlier failures, and reports the total rows stamped and the
// first stamp error encountered, if any
func stampInstallationProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration) (int, error) {
	var stamped int

	var firstErr error

	for _, schema := range entityops.AllSchemas() {
		if schema.FillProvenance == nil {
			continue
		}

		affected, err := schema.FillProvenance(ctx, db, installation)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("schema", schema.Snake).Msg("provenance conversion: stamp failed")

			if firstErr == nil {
				firstErr = err
			}

			continue
		}

		stamped += affected
	}

	return stamped, firstErr
}
