package interceptors

import (
	"context"
	"time"

	"entgo.io/ent"
	"go.uber.org/zap"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
)

func QueryLogger(l *zap.SugaredLogger) ent.InterceptFunc {
	return func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query generated.Query) (ent.Value, error) {
			q, err := intercept.NewQuery(query)
			if err != nil {
				return nil, err
			}

			start := time.Now()
			defer func() {
				l.Infow("query duration", "duration", time.Since(start), "schema", q.Type())
			}()

			return next.Query(ctx, query)
		})
	}
}
