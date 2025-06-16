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
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		// call the next middleware to get the response
		resp := next(ctx)

		// get the operation context
		opCtx := graphql.GetOperationContext(ctx)

		opName := opCtx.OperationName
		if opName == "" && opCtx.Operation != nil {
			opName = opCtx.Operation.Name
		}

		start := opCtx.Stats.OperationStart

		success := "false"
		if resp != nil && len(resp.Errors) == 0 {
			success = "true"
		}

		metrics.GraphQLOperationTotal.WithLabelValues(opName, success).Inc()
		metrics.GraphQLOperationDuration.WithLabelValues(opName).Observe(time.Since(start).Seconds())

		return resp
	})
}
