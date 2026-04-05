package hooks

import (
	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/vendorriskscore"
	"github.com/theopenlane/core/internal/ent/generated/vendorscoringconfig"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaVendorScoringListeners registers vendor scoring mutation listeners on Gala
func RegisterGalaVendorScoringListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeVendorScoringConfig),
			Name:  "vendorscoring.config_mode_change",
			Operations: []string{
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Handle: handleVendorScoringConfigMutationGala,
		},
	)
}

// handleVendorScoringConfigMutationGala recomputes entity risk aggregates when
// scoring_mode or risk_thresholds change on a VendorScoringConfig
func handleVendorScoringConfigMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if !eventqueue.MutationFieldChanged(payload, vendorscoringconfig.FieldScoringMode) &&
		!eventqueue.MutationFieldChanged(payload, vendorscoringconfig.FieldRiskThresholds) {
		return nil
	}

	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	configID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || configID == "" {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)

	// Find all distinct entity IDs that have risk scores under this config
	scores, err := client.VendorRiskScore.Query().
		Where(vendorriskscore.VendorScoringConfigID(configID)).
		Select(vendorriskscore.FieldEntityID).
		All(allowCtx)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Str("config_id", configID).Msg("failed to query entities for scoring mode recomputation")
		return err
	}

	entityIDs := lo.Uniq(lo.Map(scores, func(s *entgen.VendorRiskScore, _ int) string {
		return s.EntityID
	}))

	errs := lo.FilterMap(entityIDs, func(entityID string, _ int) (error, bool) {
		err := RecomputeEntityRiskAggregate(allowCtx, client, entityID)
		if err != nil {
			logx.FromContext(ctx.Context).Error().Err(err).Str("entity_id", entityID).Str("config_id", configID).Msg("failed to recompute entity risk aggregate")
		}

		return err, err != nil
	})

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}
