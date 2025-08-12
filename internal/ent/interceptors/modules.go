package interceptors

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gertd/go-pluralize"
	entintercept "github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	features "github.com/theopenlane/core/internal/entitlements/features"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// InterceptorModules usese the query type to automatically validate the modules
// from the auto generated pipeline
func InterceptorModules() ent.Interceptor {
	return entintercept.TraverseFunc(func(ctx context.Context, q entintercept.Query) error {

		schemaFeatures, _ := features.FeatureOfType[q.Type()]
		// if !exists {
		// 	return nil
		// }

		ok, module, err := rule.HasAllFeatures(ctx, schemaFeatures...)
		if err != nil || !ok {

			if err == nil {
				err = ErrFeatureNotEnabled
			}

			// force an evaluation to false always
			// so the data to be returned will always be empty or not found
			q.WhereP(func(s *sql.Selector) {
				s.Where(sql.ExprP("1=0"))
			})

			entity := pluralize.NewClient().Plural(strings.ToLower(q.Type()))

			path := graphql.GetPath(ctx)

			if len(path) == 0 {
				path = ast.Path{ast.PathName(entity)}
			} else if lastName, ok := path[len(path)-1].(ast.PathName); !ok || string(lastName) != entity {
				path = append(path, ast.PathName(entity))
			}

			graphql.AddError(ctx, &gqlerror.Error{
				Err:     gqlerrors.NewCustomErrorWithModule(gqlerrors.NoAccessToModule, ErrFeatureNotEnabled.Error(), err, module),
				Message: ErrFeatureNotEnabled.Error(),
				Path:    path,
			})
		}

		return nil
	})
}
