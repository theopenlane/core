package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/iam/auth"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/theopenlane/ent/generated/group"
	"github.com/theopenlane/ent/generated/groupsetting"
	"github.com/theopenlane/ent/generated/intercept"
	"github.com/theopenlane/ent/generated/organization"
)

// InterceptorGroupSetting is middleware to change the GroupSetting query
// to only include the objects that the user has access to
// by filtering the group settings with groups from the authorized organization only
func InterceptorGroupSetting() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// skip if not a graph ctx
		if hasOpCtx := graphql.HasOperationContext(ctx); !hasOpCtx {
			return nil
		}

		// skip if we are creating an object, no need to filter
		opCtx := graphql.GetOperationContext(ctx)
		if opCtx.Operation.Operation == ast.Mutation {
			return nil
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return err
		}

		q.WhereP(
			groupsetting.Or(
				groupsetting.HasGroupWith(
					group.HasOwnerWith(organization.IDIn(orgIDs...)),
				),
			),
		)

		return nil
	})
}
