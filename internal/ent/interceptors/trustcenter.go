package interceptors

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/iam/auth"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/pkg/logx"
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
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, q ent.Query) (ent.Value, error) {
			query, err := intercept.NewQuery(q)
			if err != nil {
				return nil, err
			}

			if err := applyTrustCenterChildFilters(ctx, query); err != nil {
				return nil, err
			}

			v, err := next.Query(ctx, q)
			if err != nil {
				logx.FromContext(ctx).Err(err).Msg("trust center child query failed")

				if graphql.HasOperationContext(ctx) {
					graphql.AddError(ctx, &gqlerror.Error{
						Err:     gqlerrors.NewCustomError(gqlerrors.InternalServerErrorCode, "failed to load trust center data", err),
						Message: "failed to load trust center data",
						Path:    graphql.GetPath(ctx),
					})
				}
			}

			return v, nil
		})
	})
}

func applyTrustCenterChildFilters(ctx context.Context, q intercept.Query) error {
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
		return err
	}

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
}
