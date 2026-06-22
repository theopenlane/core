package schema

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/pkg/logx"
)

// parentBlockedGroupsInterceptor returns an interceptor that filters out objects
// where the user's group is in the parent type's blocked_groups join table.
// This allows skipping per-object FGA checks on list queries when access is
// controlled by the parent's blocked groups rather than the object's own
func parentBlockedGroupsInterceptor(parentType string) ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if _, ok := auth.ActiveTrustCenterIDKey.Get(ctx); ok {
			return nil
		}

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			return auth.ErrNoAuthUser
		}

		if skip := groupPermissionInterceptorSkipper(ctx, caller); skip {
			return nil
		}

		groupIDs, err := generated.FromContext(ctx).Group.Query().Where(
			group.HasMembersWith(
				groupmembership.UserID(caller.SubjectID),
			),
		).IDs(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to get group IDs for user")

			return err
		}

		addParentBlockedGroupPredicate(q, groupIDs, parentType)

		return nil
	})
}

// addParentBlockedGroupPredicate adds a predicate that excludes objects whose parent
// has the user's group in the parent type's blocked_groups join table.
// Uses NOT EXISTS instead of LEFT JOIN to avoid introducing the join table's columns
// into the outer query scope, which would cause ambiguous column errors when the
// same FK column name appears in both the child table and the blocked_groups table
// (e.g. control_id in subcontrols and control_blocked_groups).
func addParentBlockedGroupPredicate(q intercept.Query, groupIDs []string, parentType string) {
	parentSnakeCase := strcase.SnakeCase(parentType)
	parentFKField := fmt.Sprintf("%s_id", parentSnakeCase)
	joinTableName := fmt.Sprintf("%s_blocked_groups", parentSnakeCase)

	if len(groupIDs) == 0 {
		return
	}

	q.WhereP(func(s *sql.Selector) {
		t := sql.Table(joinTableName).As(joinTableName)
		subquery := sql.SelectExpr(sql.Raw("1")).From(t).Where(
			sql.And(
				sql.EQ(t.C(parentFKField), s.C(parentFKField)),
				sql.In(t.C("group_id"), lo.ToAnySlice(groupIDs)...),
			),
		)
		s.Where(sql.Not(sql.Exists(subquery)))
	})
}
