package interceptors

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/vektah/gqlparser/v2/ast"
)

// InterceptorTrustCenter is middleware to change the TrustCenter query
func InterceptorTrustCenter() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		logx.FromContext(ctx).Debug().Msg("InterceptorTrustCenter")

		if anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx); ok {
			q.WhereP(trustcenter.IDEQ(anon.TrustCenterID))
		}

		return nil
	})
}

// InterceptorTrustCenterChild is middleware to change the TrustCenterChild query.
// Should be used by schemas that are owned by a trust center
func InterceptorTrustCenterChild() ent.Interceptor {
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

		if auth.IsSystemAdminFromContext(ctx) {
			return nil
		}

		if anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx); ok {
			if anon.TrustCenterID != "" && anon.OrganizationID != "" {
				q.WhereP(sql.FieldEQ("trust_center_id", anon.TrustCenterID))
				return nil
			}
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to get organization IDs in InterceptorTrustCenterChild")
			return err
		}

		logx.FromContext(ctx).Debug().
			Strs("org_ids", orgIDs).
			Str("query_type", q.Type()).
			Msg("InterceptorTrustCenterChild filtering by org IDs")

		q.WhereP(func(s *sql.Selector) {
			t := sql.Table(trustcenter.Table)

			anys := make([]any, len(orgIDs))
			for i, s := range orgIDs {
				anys[i] = s
			}

			s.Where(
				sql.In(
					s.C("trust_center_id"),
					sql.Select(t.C(trustcenter.FieldID)).From(t).Where(
						sql.In(
							t.C(trustcenter.FieldOwnerID), anys...,
						),
					),
				),
			)
		})

		return nil
	})
}
