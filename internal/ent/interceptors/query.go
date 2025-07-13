package interceptors

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/rs/zerolog"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
)

// QueryLogger is an interceptor that logs the duration of each query.
func QueryLogger() ent.InterceptFunc {
	return func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query generated.Query) (ent.Value, error) {
			q, err := intercept.NewQuery(query)
			if err != nil {
				return nil, err
			}

			start := time.Now()
			defer func() {
				zerolog.Ctx(ctx).Info().
					Str("duration", time.Since(start).String()).
					Str("schema", q.Type()).
					Msg("query duration")
			}()

			return next.Query(ctx, query)
		})
	}
}
