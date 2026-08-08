package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/vendorriskscore"
	"github.com/theopenlane/core/internal/ent/generated/vendorscoringconfig"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaVendorScoringListeners registers vendor scoring mutation listeners on Gala
func RegisterGalaVendorScoringListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return registerMutationListeners(g,
		entityops.MutationListener{
			Schema: entgen.TypeVendorScoringConfig,
			Operations: []string{
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Fields: []string{
				vendorscoringconfig.FieldScoringMode,
				vendorscoringconfig.FieldRiskThresholds,
			},
			Enrich: func(ctx context.Context, payload entityops.MutationPayload) context.Context {
				return logx.WithFields(ctx, map[string]any{"config_id": payload.EntityID})
			},
			Handle: handleVendorScoringConfigMutationGala,
		},
	)
}

// handleVendorScoringConfigMutationGala recomputes entity risk aggregates when
// scoring_mode or risk_thresholds change on a VendorScoringConfig
func handleVendorScoringConfigMutationGala(inv entityops.Invocation, _ entityops.MutationPayload) error {
	configID := inv.EntityID
	ctx := inv.Context

	// Find all distinct entity IDs that have risk scores under this config
	scores, err := inv.Client.VendorRiskScore.Query().
		Where(vendorriskscore.VendorScoringConfigID(configID)).
		Select(vendorriskscore.FieldEntityID).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to query entities for scoring mode recomputation")
		return err
	}

	entityIDs := lo.Uniq(lo.Map(scores, func(s *entgen.VendorRiskScore, _ int) string {
		return s.EntityID
	}))

	return errors.Join(lo.Map(entityIDs, func(entityID string, _ int) error {
		err := RecomputeEntityRiskAggregate(ctx, inv.Client, entityID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("entity_id", entityID).Msg("failed to recompute entity risk aggregate")
		}

		return err
	})...)
}
