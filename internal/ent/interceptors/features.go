package interceptors

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	entintercept "github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/pkg/models"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func InterceptorFeatures(features ...models.OrgModule) ent.Interceptor {
	return entintercept.TraverseFunc(func(ctx context.Context, q entintercept.Query) error {
		ok, err := rule.HasAllFeatures(ctx, features...)
		if err != nil || !ok {

			if err == nil {
				err = ErrFeatureNotEnabled
			}

			log.Err(err).Msg("cannot access all modules")

			// force an evaluation to false always
			// so the data to be returned will always be empty or not found
			q.WhereP(func(s *sql.Selector) {
				s.Where(sql.ExprP("1=0"))
			})

			entity := strings.ToLower(q.Type())
			if !strings.HasSuffix(entity, "s") {
				entity += "s"
			}

			path := graphql.GetPath(ctx)

			if len(path) == 0 {
				path = ast.Path{ast.PathName(entity)}
			} else if lastName, ok := path[len(path)-1].(ast.PathName); !ok || string(lastName) != entity {
				path = append(path, ast.PathName(entity))
			}

			graphql.AddError(ctx, &gqlerror.Error{
				Err:     gqlerrors.NewCustomError(gqlerrors.NoAccessToModule, ErrFeatureNotEnabled.Error(), ErrFeatureNotEnabled),
				Message: ErrFeatureNotEnabled.Error(),
				Path:    path,
			})
		}

		return nil
	})
}
