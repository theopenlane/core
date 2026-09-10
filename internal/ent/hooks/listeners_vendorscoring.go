package hooks

import (
	"errors"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	entgen "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/vendorriskscore"
	"github.com/theopenlane/core/v2/internal/ent/generated/vendorscoringconfig"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// VendorScoringListeners recomputes entity risk aggregates when vendor scoring configuration changes
func VendorScoringListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema: entityops.SchemaVendorScoringConfig,
			Operations: []string{
				entityops.OpUpdate,
				entityops.OpUpdateOne,
			},
			Fields: []string{
				vendorscoringconfig.FieldScoringMode,
				vendorscoringconfig.FieldRiskThresholds,
			},
			Handle: handleVendorScoringConfigMutationGala,
		},
	}
}

// handleVendorScoringConfigMutationGala recomputes entity risk aggregates when
// scoring_mode or risk_thresholds change on a VendorScoringConfig
func handleVendorScoringConfigMutationGala(inv entityops.Invocation, _ entityops.MutationPayload) error {
	scores, err := inv.Client.VendorRiskScore.Query().
		Where(vendorriskscore.VendorScoringConfigID(inv.EntityID)).
		Select(vendorriskscore.FieldEntityID).
		All(inv.Context)
	if err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to query entities for scoring mode recomputation")

		return err
	}

	entityIDs := lo.Uniq(lo.Map(scores, func(s *entgen.VendorRiskScore, _ int) string {
		return s.EntityID
	}))

	return errors.Join(lo.Map(entityIDs, func(entityID string, _ int) error {
		err := RecomputeEntityRiskAggregate(inv.Context, inv.Client, entityID)
		if err != nil {
			logx.FromContext(inv.Context).Error().Err(err).Str("entity_id", entityID).Msg("failed to recompute entity risk aggregate")
		}

		return err
	})...)
}
