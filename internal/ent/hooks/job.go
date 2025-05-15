package hooks

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/iam/auth"
)

// HookJobRunnerRegistrationToken auto deletes registration tokens once they have been
// used to create a runner
func HookJobRunnerRegistrationToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.JobRunnerRegistrationTokenFunc(
			func(ctx context.Context, m *generated.JobRunnerRegistrationTokenMutation) (generated.Value, error) {

				runnerID, ok := m.JobRunnerID()
				if !ok || runnerID == "" {
					return next.Mutate(ctx, m)
				}

				userID, err := auth.GetSubjectIDFromContext(ctx)
				if err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("unable to determine requestor")
					return m, err
				}

				m.SetDeletedAt(time.Now())
				m.SetDeletedBy(userID)

				return next.Mutate(ctx, m)
			})
	}, ent.OpUpdate|ent.OpUpdateOne)
}
