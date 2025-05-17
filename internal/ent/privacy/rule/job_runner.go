package rule

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/jobrunner"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// SystemOwnedJobRunner is a privacy rule that checks if the job runner is system owned
// and if the user is a system admin
// For Create operations, it will always be allowed to go through
// For Update/Delete operations, the rule checks if runner is system owned
// and denys if it is and the user is not a system admin
func SystemOwnedJobRunner() privacy.JobRunnerMutationRuleFunc {
	return privacy.JobRunnerMutationRuleFunc(func(ctx context.Context, m *generated.JobRunnerMutation) error {
		isAdmin, err := CheckIsSystemAdmin(ctx, m)
		if err != nil {
			return err
		}

		if isAdmin {
			return privacy.Allow
		}

		if m.Op() == ent.OpCreate {
			return privacy.Allow
		}

		ids, err := m.IDs(ctx)
		if err != nil {
			return err
		}

		runners, err := m.Client().JobRunner.Query().
			Where(jobrunner.IDIn(ids...)).
			Select(jobrunner.FieldSystemOwned).
			All(ctx)
		if err != nil {
			return err
		}

		for _, r := range runners {
			if r.SystemOwned {
				return generated.ErrPermissionDenied
			}
		}

		return privacy.Allow
	})
}
