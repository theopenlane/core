package serveropts

import (
	"context"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/subcontrol"
)

func backfillOrigins(ctx context.Context, deps backfillDeps) error {

	client := deps.Client

	type retrievedItem struct {
		id         string
		ownerID    string
		sourceName string
	}

	backfills := []struct {
		schemaName string
		queryFn    func(context.Context, string) ([]retrievedItem, error)
		updateFn   func(context.Context, retrievedItem) error
	}{
		{
			schemaName: "controls",
			queryFn: func(ctx context.Context, lastID string) ([]retrievedItem, error) {
				items, err := client.Control.Query().
					Where(
						control.IDGT(lastID),
						control.OwnerIDNotNil(),
						control.SourceNameNEQ("")).
					Order(control.ByID()).
					Limit(backfillBatchSize).
					Select(control.FieldID, control.FieldOwnerID, control.FieldSourceName).
					All(ctx)
				if err != nil {
					return nil, err
				}

				return lo.Map(items, func(control *generated.Control, _ int) retrievedItem {
					return retrievedItem{
						id:         control.ID,
						ownerID:    control.OwnerID,
						sourceName: lo.FromPtr(control.SourceName),
					}
				}), nil
			},
			updateFn: func(ctx context.Context, item retrievedItem) error {
				return client.Control.UpdateOneID(item.id).
					SetControlOriginName(item.sourceName).
					Exec(ctx)
			},
		},
		{
			schemaName: "subcontrols",
			queryFn: func(ctx context.Context, lastID string) ([]retrievedItem, error) {
				items, err := client.Subcontrol.Query().
					Where(
						subcontrol.IDGT(lastID),
						subcontrol.OwnerIDNotNil(),
						subcontrol.SourceNameNEQ("")).
					Order(subcontrol.ByID()).
					Limit(backfillBatchSize).
					Select(subcontrol.FieldID, subcontrol.FieldOwnerID, subcontrol.FieldSourceName).
					All(ctx)
				if err != nil {
					return nil, err
				}

				return lo.Map(items, func(control *generated.Subcontrol, _ int) retrievedItem {
					return retrievedItem{
						id:         control.ID,
						ownerID:    control.OwnerID,
						sourceName: lo.FromPtr(control.SourceName),
					}
				}), nil
			},
			updateFn: func(ctx context.Context, item retrievedItem) error {
				return client.Subcontrol.UpdateOneID(item.id).
					SetControlOriginName(item.sourceName).
					Exec(ctx)
			},
		},
	}

	for _, backfill := range backfills {
		var lastKnownID string

		for {
			items, err := backfill.queryFn(ctx, lastKnownID)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).
					Str("schema", backfill.schemaName).
					Str("ordered_from_id", lastKnownID).
					Msg("backfill: failed to query control origins")

				return err
			}

			if len(items) == 0 {
				break
			}

			for _, item := range items {
				caller := &auth.Caller{
					OrganizationID: item.ownerID,
					Capabilities:   backfillBypassCaps,
				}

				callerCtx := auth.WithCaller(ctx, caller)

				if err := backfill.updateFn(callerCtx, item); err != nil {
					logx.FromContext(ctx).Error().Err(err).
						Str("schema", backfill.schemaName).
						Str("item_id", item.id).
						Msg("backfill: failed to migrate origin")

					return err
				}
			}

			lastKnownID = items[len(items)-1].id
		}
	}

	return nil
}
