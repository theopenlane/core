package interceptors

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/pkg/gqlerrors"
	"github.com/theopenlane/core/pkg/logx"
	features "github.com/theopenlane/ent/entitlements/features"
	"github.com/theopenlane/ent/generated"
	entintercept "github.com/theopenlane/ent/generated/intercept"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/utils/contextx"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type moduleInterceptorKey struct{}

// InterceptorModules uses the query type to automatically validate the modules
// from the auto generated pipeline
func InterceptorModules(modulesEnabled bool) ent.Interceptor {
	return entintercept.TraverseFunc(func(ctx context.Context, q entintercept.Query) error {
		if !modulesEnabled {
			return nil
		}

		fCtx := graphql.GetFieldContext(ctx)

		fieldCheck := ""

		if fCtx != nil {
			if fCtx.Object == "Query" {
				fieldCheck = fCtx.Field.Name
			}
		}

		if _, ok := contextx.From[moduleInterceptorKey](ctx); ok {
			return nil
		}

		if rule.ShouldSkipFeatureCheck(ctx) {
			return nil
		}

		schemaFeatures, exists := features.FeatureOfType[q.Type()]
		if !exists {
			return nil
		}

		// prevent infinite recursion. HasAllFeatures calls the OrgModule queries in some scenarios.
		// This prevents a scenario where this interceptor is called again when already inside this function
		ctxWithKey := contextx.With(ctx, moduleInterceptorKey{})

		ok, module, err := rule.HasAnyFeature(ctxWithKey, schemaFeatures...)
		if err != nil || !ok {
			if err == nil {
				err = ErrFeatureNotEnabled
			}

			logx.FromContext(ctx).Info().Interface("required_features", schemaFeatures).Str("missing_module", module.String()).Msg("feature not enabled for organization")

			// force an evaluation to false always
			// so the data to be returned will always be empty or not found
			q.WhereP(func(s *sql.Selector) {
				s.Where(sql.ExprP("1=0"))
			})

			entity := pluralize.NewClient().Plural(strings.ToLower(q.Type()))

			path := graphql.GetPath(ctx)

			// {
			//    "message": "task not found",
			//    "path": [
			//      "task"
			//    ],
			//    "extensions": {
			//      "code": "NOT_FOUND",
			//      "message": "task not found"
			//    }
			//  },
			//  {
			//    "message": "feature not enabled for organization",
			//    "path": [
			//      "organizations",
			//      "tasks"
			//    ],
			//    "extensions": {
			//      "code": "MODULE_NO_ACCESS",
			//      "message": "feature not enabled for organization"
			//    }
			//  },
			//  {
			//    "message": "feature not enabled for organization",
			//    "path": [
			//      "organizations",
			//      "programs"
			//    ],
			//    "extensions": {
			//      "code": "MODULE_NO_ACCESS",
			//      "message": "feature not enabled for organization"
			//    }
			//  }
			//
			//  we want to be able to show the user the exact nodes and schemas they don't have
			//  access to. So this constructs the path array in the above json . it could be ["task"] alone
			//  which means the task schema cannot be accessed
			//
			//  or it could be ["organization","tasks"] which means a query like { org { tasks {}}} was
			//  requested and the tasks section could not be retrieved
			if len(path) == 0 {
				path = ast.Path{ast.PathName(entity)}
			} else if lastName, ok := path[len(path)-1].(ast.PathName); !ok || string(lastName) != entity {
				path = append(path, ast.PathName(entity))
			}

			// ignore the error graph enrichment when a global search is done
			if graphql.HasOperationContext(ctx) && fieldCheck != "search" {
				graphql.AddError(ctx, &gqlerror.Error{
					Err:     gqlerrors.NewCustomErrorWithModule(gqlerrors.NoAccessToModule, ErrFeatureNotEnabled.Error(), err, module),
					Message: ErrFeatureNotEnabled.Error(),
					Path:    path,
				})
			} else if fieldCheck != "search" {
				// this shouldn't happen unless a REST request is requesting data that isn't in the base module
				// adding warning here to indicate potential misconfiguration
				logx.FromContext(ctx).Info().Str("field", fieldCheck).Msg("graphql operation not found, unable to set graphql error for missing module")

				return generated.ErrPermissionDenied
			}
		}

		return nil
	})
}
