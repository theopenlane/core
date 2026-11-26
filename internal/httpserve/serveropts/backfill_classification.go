package serveropts

import (
	"context"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated/review"
)

func backfillClassifications(ctx context.Context, deps backfillDeps) error {
	var lastKnownID string

	for {
		reviews, err := deps.Client.Review.Query().
			Where(
				review.IDGT(lastKnownID),
				review.OwnerIDNotNil(),
				review.ClassificationNEQ(""),
				review.ReviewClassificationIDIsNil(),
				review.ReviewClassificationNameIsNil(),
			).
			Order(review.ByID()).
			Limit(backfillBatchSize).
			Select(review.FieldID, review.FieldOwnerID, review.FieldClassification).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).
				Str("ordered_from_id", lastKnownID).
				Msg("backfill: failed to query review classifications")

			return err
		}

		if len(reviews) == 0 {
			return nil
		}

		for _, item := range reviews {
			caller := &auth.Caller{
				OrganizationID: item.OwnerID,
				Capabilities:   backfillBypassCaps,
			}

			err := deps.Client.Review.UpdateOneID(item.ID).
				SetReviewClassificationName(item.Classification).
				Exec(auth.WithCaller(ctx, caller))
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).
					Str("review_id", item.ID).
					Msg("backfill: failed to migrate review classification")

				return err
			}
		}

		lastKnownID = reviews[len(reviews)-1].ID
	}
}
