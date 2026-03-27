package hooks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

func getNextReviewDate(frequency enums.Frequency, lastReviewedAt models.DateTime) models.DateTime {
	lastReviewDate := time.Time(lastReviewedAt)

	switch frequency {
	case enums.FrequencyYearly:
		return models.DateTime(lastReviewDate.AddDate(1, 0, 0)) //nolint:mnd
	case enums.FrequencyBiAnnually:
		return models.DateTime(lastReviewDate.AddDate(0, 6, 0)) //nolint:mnd
	case enums.FrequencyQuarterly:
		return models.DateTime(lastReviewDate.AddDate(0, 3, 0)) //nolint:mnd
	case enums.FrequencyMonthly:
		return models.DateTime(lastReviewDate.AddDate(0, 1, 0)) //nolint:mnd
	default:
		return models.DateTime{}
	}
}

// HookReviewFiles runs on review mutations to check for uploaded files
func HookReviewFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ReviewFunc(func(ctx context.Context, m *generated.ReviewMutation) (generated.Value, error) {
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = pkgobjects.ProcessFilesForMutation(ctx, m, "reviewFiles")
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			entityIDs := m.EntitiesIDs()
			if len(entityIDs) == 0 {
				return v, nil
			}

			userID, exists := m.CreatedBy()
			createdAt, createdAtExists := m.CreatedAt()

			var errs []string
			var mu sync.Mutex

			funcs := make([]func(), 0, len(entityIDs))
			for _, id := range entityIDs {
				funcs = append(funcs, func() {
					ent, err := m.Client().Entity.
						Query().Select(entity.FieldReviewFrequency).
						Where(entity.ID(id)).
						Only(ctx)
					if err != nil {
						logx.FromContext(ctx).Err(err).
							Str("entity_id", id).
							Msg("could not fetch entity for review")
						mu.Lock()
						errs = append(errs, err.Error())
						mu.Unlock()
						return
					}

					q := m.Client().Entity.UpdateOneID(id)

					if exists {
						q = q.SetReviewedBy(userID)
					}

					if createdAtExists {
						q = q.SetLastReviewedAt(models.DateTime(createdAt))

						if frequency := ent.ReviewFrequency; frequency != enums.FrequencyNone {
							q = q.SetNextReviewAt(getNextReviewDate(frequency, models.DateTime(createdAt)))
						}
					}

					err = q.Exec(ctx)
					if err != nil {
						logx.FromContext(ctx).Err(err).
							Str("entity_id", id).Msg("could not update entity reviewer")
						mu.Lock()
						errs = append(errs, err.Error())
						mu.Unlock()
					}
				})
			}

			if err := m.Client().Pool.SubmitMultipleAndWait(funcs); err != nil {
				return nil, err
			}

			if len(errs) > 0 {
				logx.FromContext(ctx).Error().
					Int("error_count", len(errs)).
					Strs("errors", errs).
					Msg("review file hook: entity updates failed")

				return nil, fmt.Errorf("%d entities could not be updated", len(errs)) //nolint:err113
			}

			return v, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}
