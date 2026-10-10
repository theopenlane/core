package runtime

import (
	"context"
	"errors"
	"slices"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// UpgradeInstallations upgrades every connected, degraded, or errored installation whose stored definition version predates the registry version
func (r *Runtime) UpgradeInstallations(ctx context.Context) error {
	definitionIDs := lo.Map(r.Registry().Definitions(), func(def types.Definition, _ int) string {
		return def.ID
	})
	if len(definitionIDs) == 0 {
		return nil
	}

	installations, err := r.DB().Integration.Query().
		Where(
			integration.StatusIn(append(slices.Clone(enums.IntegrationOperationalStatuses), enums.IntegrationStatusErrored)...),
			integration.DefinitionIDIn(definitionIDs...),
		).
		All(auth.EnsureIntegrationCaller(ctx, ""))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed querying installations")

		return err
	}

	stale := staleInstallations(r.Registry(), installations)

	failures := lo.Filter(lo.Map(stale, func(inst *ent.Integration, _ int) error {
		return r.ensureCurrentVersion(auth.EnsureIntegrationCaller(ctx, inst.OwnerID), inst)
	}), func(err error, _ int) bool {
		return err != nil
	})

	return errors.Join(failures...)
}

// staleInstallations keeps the installations whose stored definition version predates the registry version
func staleInstallations(reg *registry.Registry, installations []*ent.Integration) []*ent.Integration {
	return lo.Filter(installations, func(inst *ent.Integration, _ int) bool {
		return outdated(inst.DefinitionVersion, reg.Version(inst.DefinitionID))
	})
}
