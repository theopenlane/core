package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/intercept"
)

const (
	redacted = "*****************************"
)

// InterceptorPat is middleware to change the PAT query
func InterceptorPat() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.PersonalAccessTokenFunc(func(ctx context.Context, q *generated.PersonalAccessTokenQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			pats, ok := v.([]*generated.PersonalAccessToken)
			if !ok {
				return v, err
			}

			// redact the token on get
			for _, pat := range pats {
				pat.Token = redacted
			}

			return pats, nil
		})
	})
}
