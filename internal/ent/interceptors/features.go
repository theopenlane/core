package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// InterceptorRequireFeatureProgram ensures the organization has the given feature enabled before executing the query
func InterceptorRequireFeatureProgram(feature string) ent.Interceptor {
	return InterceptorRequireAnyFeatureProgram(feature)
}

// InterceptorRequireAnyFeatureProgram ensures the organization has at least one of the provided features enabled before executing the query
func InterceptorRequireAnyFeatureProgram(features ...string) ent.Interceptor {
	return intercept.TraverseProgram(func(ctx context.Context, _ *generated.ProgramQuery) error {
		ok, err := rule.HasAnyFeature(ctx, features...)
		if err != nil {
			return err
		}

		if !ok {
			return ErrFeatureNotEnabled
		}

		return nil
	})
}
