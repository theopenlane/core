package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/intercept"
)

// InterceptorAPIToken is middleware to change the api token query
func InterceptorAPIToken() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.APITokenFunc(func(ctx context.Context, q *generated.APITokenQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			apiTokens, ok := v.([]*generated.APIToken)
			if !ok {
				return v, err
			}

			// redact the token on get
			for _, t := range apiTokens {
				t.Token = redacted
			}

			return apiTokens, nil
		})
	})
}
