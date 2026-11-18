package serveropts

import (
	"context"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/risk"
)

func backfillRiskEntities(ctx context.Context, client *ent.Client) error {
	batchSize := 50

	var lastID string

	for {
		risks, err := client.Risk.Query().
			Where(
				risk.HasLegacyEntities(),
				risk.IDGT(lastID),
			).
			Order(ent.Asc(risk.FieldID)).
			Limit(batchSize).
			Select(risk.FieldID).
			WithLegacyEntities(func(q *ent.EntityQuery) {
				q.Select(entity.FieldID)
			}).
			All(ctx)
		if err != nil {
			return err
		}

		if len(risks) == 0 {
			return nil
		}

		for _, r := range risks {
			if len(r.Edges.LegacyEntities) == 0 {
				continue
			}

			err := client.Risk.UpdateOneID(r.ID).
				AddEntities(r.Edges.LegacyEntities...).
				Exec(ctx)
			if err != nil {
				return err
			}
		}

		lastID = risks[len(risks)-1].ID
	}
}
