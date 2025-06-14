package graphapi

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"

	"github.com/theopenlane/core/pkg/metrics"
)

// WithMetrics adds Prometheus instrumentation around GraphQL operations.
func WithMetrics(h *handler.Server) {
	h.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		opCtx := graphql.GetOperationContext(ctx)
		opName := "unknown"
		if opCtx != nil {
			opName = opCtx.OperationName
			if opName == "" && opCtx.Operation != nil {
				opName = string(opCtx.Operation.Operation)
			}
		}

		start := time.Now()

		return func(ctx context.Context) *graphql.Response {
			resp := next(ctx)(ctx)

			success := "true"
			if resp != nil && len(resp.Errors) > 0 {
				success = "false"
			}

			metrics.GraphQLOperationTotal.WithLabelValues(opName, success).Inc()
			metrics.GraphQLOperationDuration.WithLabelValues(opName).Observe(time.Since(start).Seconds())

			return resp
		}
	})
}
