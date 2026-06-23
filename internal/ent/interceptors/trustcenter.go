package interceptors

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/auth"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/pkg/logx"
)

// InterceptorTrustCenter is middleware to change the TrustCenter query
func InterceptorTrustCenter() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if tcID, ok := auth.ActiveTrustCenterIDKey.Get(ctx); ok && tcID != "" {
			q.WhereP(trustcenter.IDEQ(tcID))
		}

		return nil
	})
}

// AnonInterceptorTrustCenterChild filters trust center child queries for anon callers only
// Use this for org-owned schemas (e.g. templates) where regular users rely on the org interceptor
func AnonInterceptorTrustCenterChild() ent.Interceptor {
	return trustCenterInterceptor(false, true)
}

// InterceptorTrustCenterChild filters trust center child queries for all callers
// Use this for Trust Center owned schemas (e.g. docs, FAQs) where anon Trust Center  users can read
// and regular users need a fallback org filter
func InterceptorTrustCenterChild() ent.Interceptor {
	return trustCenterInterceptor(true, true)
}

// InterceptorTrustCenterChildDenyAnon applies the fallback org filter for regular users
// but denies all anonymous trust center callers, even those with an active Trust Center  key
// Use for Trust Center -owned schemas that anon users submit to but must not read back (e.g. NDA requests)
func InterceptorTrustCenterChildDenyAnon() ent.Interceptor {
	return trustCenterInterceptor(true, false)
}

// trustCenterInterceptor builds the shared interceptor body for all filters and should be used with the exported functions
// applyToAllRequests: when true, regular (non-anon) users get a fallback SQL filter scoped to
// trust centers owned by their organizations
// allowAnonAccess: when true, anon Trust Center  callers with an active key get a trust_center_id filter;
// when false they are denied outright from querying data
func trustCenterInterceptor(applyToAllRequests, allowAnonAccess bool) ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, q ent.Query) (ent.Value, error) {
			query, err := intercept.NewQuery(q)
			if err != nil {
				return nil, err
			}

			if err := applyTrustCenterChildFilters(ctx, query, applyToAllRequests, allowAnonAccess); err != nil {
				return nil, err
			}

			v, err := next.Query(ctx, q)
			if err != nil {
				logx.FromContext(ctx).Err(err).Str("type", query.Type()).Msg("trust center child query failed")

				// only do this for anon trust center tokens
				if _, ok := auth.ActiveTrustCenterIDKey.Get(ctx); ok && graphql.HasOperationContext(ctx) {
					entity := pluralize.NewClient().Plural(strings.ToLower(query.Type()))

					path := graphql.GetPath(ctx)

					if len(path) == 0 {
						path = ast.Path{ast.PathName(entity)}
					} else if lastName, ok := path[len(path)-1].(ast.PathName); !ok || string(lastName) != entity {
						path = append(path, ast.PathName(entity))
					}

					msg := fmt.Sprintf("failed to load %s data", query.Type())

					graphql.AddError(ctx, &gqlerror.Error{
						Err:     gqlerrors.NewCustomError(gqlerrors.InternalServerErrorCode, msg, err),
						Message: "failed to load trust center data",
						Path:    path,
					})
				}
			}

			return v, nil
		})
	})
}

func applyTrustCenterChildFilters(ctx context.Context, q intercept.Query, applyToAllRequests, allowAnonAccess bool) error {
	if hasOpCtx := graphql.HasOperationContext(ctx); hasOpCtx {
		// skip if we are creating an object, no need to filter
		opCtx := graphql.GetOperationContext(ctx)
		if opCtx.Operation.Operation == ast.Mutation {
			return nil
		}
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return auth.ErrNoAuthUser
	}

	if caller.Has(auth.CapSystemAdmin) {
		return nil
	}

	// get the trust center key from the context
	tcID, ok := auth.ActiveTrustCenterIDKey.Get(ctx)
	if ok && tcID != "" {
		_, allowRequest := privacy.DecisionFromContext(ctx)
		if !allowAnonAccess && !allowRequest {
			return privacy.Denyf("anonymous trust center access not allowed for this resource: %s", q.Type())
		}

		// filter the request by trust center
		q.WhereP(sql.FieldEQ("trust_center_id", tcID))

		return nil
	}

	// deny all trust center requests that did not have a trust center key
	if caller.Has(auth.CapTrustCenterAnonymous) {
		return privacy.Denyf("trust center request without active trust center key")
	}

	// if the trust center filter should not be applied to all requests (e.g for schemas that are org owned and not solely trust
	// center owned, continue with no filter)
	if !applyToAllRequests {
		return nil
	}

	// filter by trust centers owned by the organization in the caller context
	orgIDs := caller.OrgIDs()

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
