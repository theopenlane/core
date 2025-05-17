package hooks

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/jobrunnerregistrationtoken"
	"github.com/theopenlane/iam/auth"
)

// HookJobRunnerRegistrationToken auto deletes registration tokens
//
// There can also be only one token available at any given time.
// - If a new token is generated, delete the existing registration tokens
// - If a token has been used to successfully register a job runner node, delete it
func HookJobRunnerRegistrationToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.JobRunnerRegistrationTokenFunc(
			func(ctx context.Context, m *generated.JobRunnerRegistrationTokenMutation) (generated.Value, error) {

				if m.Op().Is(ent.OpCreate) {
					orgID, err := auth.GetOrganizationIDFromContext(ctx)
					if err != nil {
						return nil, err
					}

					_, err = m.Client().JobRunnerRegistrationToken.Delete().
						Where(jobrunnerregistrationtoken.OwnerID(orgID)).
						Exec(ctx)
					if err != nil {
						return nil, err
					}

					return next.Mutate(ctx, m)
				}

				runnerID, ok := m.JobRunnerID()
				if !ok || runnerID == "" {
					return next.Mutate(ctx, m)
				}

				userID, err := auth.GetSubjectIDFromContext(ctx)
				if err != nil {
					return m, err
				}

				m.SetDeletedAt(time.Now())
				m.SetDeletedBy(userID)

				return next.Mutate(ctx, m)
			})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
