package interceptors

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
)

func QueryLogger() ent.InterceptFunc {
	return func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query generated.Query) (ent.Value, error) {
			q, err := intercept.NewQuery(query)
			if err != nil {
				return nil, err
			}

			start := time.Now()
			defer func() {
				log.Info().
					Str("duration", time.Since(start).String()).
					Str("schema", q.Type()).
					Msg("query duration")
			}()

			return next.Query(ctx, query)
		})
	}
}
