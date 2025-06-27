package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// InterceptorRequireFeature ensures the organization has the given feature
// enabled before executing the query.
func InterceptorRequireFeature(feature string) ent.Interceptor {
	return InterceptorRequireAnyFeature(feature)
}

// InterceptorRequireAnyFeature ensures the organization has at least one of the
// provided features enabled before executing the query. It can be used with any
// query type and does not rely on the query value.
func InterceptorRequireAnyFeature(features ...string) ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, _ intercept.Query) error {
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
