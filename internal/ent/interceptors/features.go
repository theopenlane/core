package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
)

// InterceptorRequireFeature ensures the organization has the given feature
// enabled before executing the query.
func InterceptorRequireFeature(feature models.OrgModule, schema string) ent.Interceptor {
	return InterceptorRequireAnyFeature(schema, feature)
}

// InterceptorRequireAnyFeature ensures the organization has at least one of the
// provided features enabled before executing the query. It can be used with any
// query type and does not rely on the query value.
func InterceptorRequireAnyFeature(schema string, features ...models.OrgModule) ent.Interceptor {
	return interceptorRequireFeatures(schema, false, features...)
}

// InterceptorRequireAllFeatures ensures the organization has all of the
// provided features enabled before executing the query. It can be used with any
// query type and does not rely on the query value.
func InterceptorRequireAllFeatures(schema string, features ...models.OrgModule) ent.Interceptor {
	return interceptorRequireFeatures(schema, true, features...)
}

// interceptorRequireFeatures is a helper function that creates an interceptor
// based on the requireAll flag. If requireAll is true, all features must be enabled.
// If false, at least one must be enabled.
func interceptorRequireFeatures(schema string, requireAll bool, features ...models.OrgModule) ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {

		if len(features) == 0 {
			return nil
		}

		// check for bypass
		// For unauthenticated users, this interceptor
		// will still run when a query is done to fetch the data such as an api
		// token or personal access token
		// And would lead to a situation where the features cannot be
		// retrieved from the database and a failure occurrs
		if _, allowCtx := privacy.DecisionFromContext(ctx); allowCtx {
			return nil
		}

		if _, ok := contextx.From[auth.OrgSubscriptionContextKey](ctx); ok {
			return nil
		}

		if _, ok := contextx.From[auth.OrganizationCreationContextKey](ctx); ok {
			return nil
		}

		if tok := token.EmailSignUpTokenFromContext(ctx); tok != nil {
			return nil
		}

		if tok := token.ResetTokenFromContext(ctx); tok != nil {
			return nil
		}

		if tok := token.VerifyTokenFromContext(ctx); tok != nil {
			return nil
		}

		if tok := token.JobRunnerRegistrationTokenFromContext(ctx); tok != nil {
			return nil
		}

		var ok bool
		var err error

		switch requireAll {
		case true:
			ok, err = rule.HasAllFeatures(ctx, features...)

		default:
			ok, err = rule.HasAnyFeature(ctx, features...)
		}

		if err != nil {
			return err
		}

		if !ok {
			return ErrFeatureNotEnabled
		}

		return nil
	})
}

func InterceptorFeatures(features ...models.OrgModule) ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {

			ok, err := rule.HasAllFeatures(ctx, features...)
			if err != nil && !ok {
				return nil, nil
			}

			return next.Query(ctx, query)
		})
	})
}
