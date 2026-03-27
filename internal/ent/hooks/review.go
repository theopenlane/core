package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

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

			for _, id := range entityIDs {
				q := m.Client().Entity.UpdateOneID(id)

				if exists {
					q = q.SetReviewedBy(userID)
				}

				if createdAtExists {
					q = q.SetLastReviewedAt(models.DateTime(createdAt))
				}

				err := q.Exec(ctx)
				if err != nil {
					logx.FromContext(ctx).Err(err).
						Str("entity_id", id).
						Msg("could not update entity reviewer")
					return nil, err
				}

			}

			return v, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}
