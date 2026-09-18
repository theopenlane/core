package serveropts

import (
	"context"

	"github.com/samber/lo"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
	"github.com/theopenlane/core/v2/internal/ent/generated/risk"
	"github.com/theopenlane/core/v2/pkg/logx"
)

func backfillSchemaResponsibilities(ctx context.Context, dbClient *ent.Client) error {
	const batchSize = 50

	type itemObject struct {
		id    string
		owner string
		risk  *ent.Risk
	}

	schemaBackfills := []struct {
		schemaName string
		queryFn    func(context.Context, string) ([]itemObject, error)
		updateFn   func(context.Context, itemObject) error
	}{
		{
			schemaName: "risk",

			queryFn: func(ctx context.Context, mostRecentID string) ([]itemObject, error) {

				query := dbClient.Risk.Query().
					Where(risk.Or(
						risk.And(risk.StakeholderIDNEQ(""), risk.StakeholderGroupIDIsNil()),
						risk.And(risk.DelegateIDNEQ(""), risk.DelegateGroupIDIsNil()),
					)).
					Order(risk.ByID()).
					Limit(batchSize)

				if mostRecentID != "" {
					query = query.Where(risk.IDGT(mostRecentID))
				}

				items, err := query.All(ctx)
				if err != nil {
					return nil, err
				}

				return lo.Map(items, func(item *ent.Risk, _ int) itemObject {
					return itemObject{id: item.ID, risk: item}
				}), nil
			},

			updateFn: func(ctx context.Context, item itemObject) error {
				update := dbClient.Risk.Update().Where(risk.ID(item.id))

				if item.risk.StakeholderID != "" && item.risk.StakeholderGroupID == "" {
					update.SetStakeholderGroupID(item.risk.StakeholderID)
				}

				if item.risk.DelegateID != "" && item.risk.DelegateGroupID == "" {
					update.SetDelegateGroupID(item.risk.DelegateID)
				}

				return update.Exec(ctx)
			},
		},
		{
			schemaName: "asset",

			queryFn: func(ctx context.Context, recentID string) ([]itemObject, error) {
				query := dbClient.Asset.Query().
					Where(
						asset.InternalOwnerNEQ(""),
						asset.InternalOwnerUserIDIsNil(),
						asset.InternalOwnerGroupIDIsNil(),
						asset.InternalOwnerIdentityHolderIDIsNil(),
					).
					Order(asset.ByID()).
					Limit(batchSize)

				if recentID != "" {
					query = query.Where(asset.IDGT(recentID))
				}

				items, err := query.All(ctx)

				return lo.Map(items, func(item *ent.Asset, _ int) itemObject {
					return itemObject{id: item.ID, owner: item.InternalOwner}
				}), err
			},

			updateFn: func(ctx context.Context, item itemObject) error {
				return dbClient.Asset.UpdateOneID(item.id).
					SetInternalOwner(item.owner).
					Exec(ctx)
			},
		},
	}

	for _, backfill := range schemaBackfills {
		recentID := ""

		for {
			items, err := backfill.queryFn(ctx, recentID)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).
					Str("schema", backfill.schemaName).
					Str("filtered_from", recentID).
					Msg("backfill: failed to query responsibility fields")

				return err
			}

			if len(items) == 0 {
				break
			}

			for _, item := range items {

				if err := backfill.updateFn(ctx, item); err != nil {
					logx.FromContext(ctx).Error().Err(err).
						Str("schema", backfill.schemaName).Str("id", item.id).
						Msg("backfill: failed to migrate responsibility fields")

					return err
				}

				recentID = item.id
			}
		}

		logx.FromContext(ctx).Info().
			Str("schema", backfill.schemaName).
			Msg("backfill: migrated responsibility fields")
	}

	return nil
}
