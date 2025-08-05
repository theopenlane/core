package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/pkg/models"
)

func InterceptorFeatures(features ...models.OrgModule) ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {

			// _, err := rule.HasAllFeatures(ctx, features...)
			// if err != nil {
			// 	return nil, nil
			// }

			return next.Query(ctx, query)
		})
	})
}
