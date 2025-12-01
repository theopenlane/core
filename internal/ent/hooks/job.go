package hooks

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/jobrunnerregistrationtoken"
	"github.com/theopenlane/core/internal/ent/generated/jobrunnertoken"
	"github.com/theopenlane/entx"
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

				m.SetDeletedAt(time.Now())

				return next.Mutate(ctx, m)
			})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// HookJobRunnerCreate makes sure there is always a token for
// the job runner node when a new runner is created
//
// This also deletes the registration token
func HookJobRunnerCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.JobRunnerFunc(
			func(ctx context.Context, m *generated.JobRunnerMutation) (generated.Value, error) {
				v, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				runner := v.(*generated.JobRunner)

				// if system owned, no registration token to delete
				// also no token to create
				if runner.SystemOwned {
					return v, err
				}

				subjectID, err := auth.GetSubjectIDFromContext(ctx)
				if err != nil {
					return nil, err
				}

				// make sure we cannot reuse the registration token
				// for cases where there is no "registration token"
				// like an admin creating a runner via api
				// or even tests.
				// Only check the error and make sure it is not a "not found".
				err = m.Client().JobRunnerRegistrationToken.UpdateOneID(runner.CreatedBy).
					SetJobRunnerID(runner.ID).
					SetDeletedAt(time.Now()).
					SetDeletedBy(subjectID).
					Exec(ctx)
				if err != nil && !generated.IsNotFound(err) {
					return nil, err
				}

				return v, m.Client().JobRunnerToken.Create().
					AddJobRunnerIDs(runner.ID).
					SetOwnerID(runner.OwnerID).
					Exec(ctx)
			})
	}, ent.OpCreate)
}

// HookJobRunnerDelete deletes all token associated with a runner when the runner is deleted
func HookJobRunnerDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.JobRunnerFunc(
			func(ctx context.Context, m *generated.JobRunnerMutation) (generated.Value, error) {
				if !entx.CheckIsSoftDeleteType(ctx, m.Type()) {
					return next.Mutate(ctx, m)
				}

				id, _ := m.ID()

				runner, err := m.Client().JobRunner.Get(ctx, id)
				if err != nil {
					return nil, err
				}

				tokens, err := runner.QueryJobRunnerTokens().All(ctx)
				if err != nil {
					return nil, err
				}

				if len(tokens) == 0 {
					return next.Mutate(ctx, m)
				}

				var ids []string

				for _, token := range tokens {
					ids = append(ids, token.ID)
				}

				//  Then delete the actual tokens
				_, err = m.Client().JobRunnerToken.Delete().
					Where(jobrunnertoken.IDIn(ids...)).
					Exec(ctx)
				if err != nil {
					return nil, err
				}

				return next.Mutate(ctx, m)
			})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne)
}
