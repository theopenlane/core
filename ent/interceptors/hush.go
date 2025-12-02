package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/intercept"
	"github.com/theopenlane/ent/hooks"
)

// InterceptorHush keeps it secret, keeps it safe
func InterceptorHush() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.HushFunc(func(ctx context.Context, q *generated.HushQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			hush, ok := v.([]*generated.Hush)
			// Skip all query types besides node queries (e.g., Count, Scan, GroupBy).
			if !ok {
				return v, nil
			}

			for _, u := range hush {
				if err := hooks.DecryptHush(u); err != nil {
					return nil, err
				}
			}

			return hush, nil
		})
	})
}
