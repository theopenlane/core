package hooks

import (
	"errors"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/vendorriskscore"
	"github.com/theopenlane/core/internal/ent/generated/vendorscoringconfig"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaVendorScoringListeners registers vendor scoring mutation listeners on Gala
func RegisterGalaVendorScoringListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return eventqueue.RegisterMutationListeners(g,
		eventqueue.MutationListener{
			Schema: entgen.TypeVendorScoringConfig,
			Name:   "vendorscoring.config_mode_change",
			Operations: []string{
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Fields: []string{
				vendorscoringconfig.FieldScoringMode,
				vendorscoringconfig.FieldRiskThresholds,
			},
			Handle: handleVendorScoringConfigMutationGala,
		},
	)
}

// handleVendorScoringConfigMutationGala recomputes entity risk aggregates when
// scoring_mode or risk_thresholds change on a VendorScoringConfig
func handleVendorScoringConfigMutationGala(inv eventqueue.Invocation, _ eventqueue.MutationGalaPayload) error {
	configID := inv.EntityID

	// Find all distinct entity IDs that have risk scores under this config
	scores, err := inv.Client.VendorRiskScore.Query().
		Where(vendorriskscore.VendorScoringConfigID(configID)).
		Select(vendorriskscore.FieldEntityID).
		All(inv.Context)
	if err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Str("config_id", configID).Msg("failed to query entities for scoring mode recomputation")
		return err
	}

	entityIDs := lo.Uniq(lo.Map(scores, func(s *entgen.VendorRiskScore, _ int) string {
		return s.EntityID
	}))

	return errors.Join(lo.Map(entityIDs, func(entityID string, _ int) error {
		err := RecomputeEntityRiskAggregate(inv.Context, inv.Client, entityID)
		if err != nil {
			logx.FromContext(inv.Context).Error().Err(err).Str("entity_id", entityID).Str("config_id", configID).Msg("failed to recompute entity risk aggregate")
		}

		return err
	})...)
}
