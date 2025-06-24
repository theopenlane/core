package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/iam/auth"
	"github.com/vektah/gqlparser/v2/ast"
)

// InterceptorTrustCenterSetting is middleware to change the TrustCenter query
// to only include the objects that the user has access to
// by filtering the trust center settings by the organization
// TODO (sfunk): this should work on all _settings tables instead
// of having a specific interceptor for each one
func InterceptorTrustCenterSetting() ent.Interceptor {
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
			trustcentersetting.Or(
				trustcentersetting.HasTrustCenterWith(
					trustcenter.HasOwnerWith(organization.IDIn(orgIDs...)),
				),
			),
		)

		return nil
	})
}
