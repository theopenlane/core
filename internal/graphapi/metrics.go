package graphapi

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"

	"github.com/theopenlane/shared/metrics"
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
		duration := time.Since(start).Seconds()

		// Get complexity from extensions if available, otherwise default to 0
		complexity := 0
		if complexityExt := opCtx.Stats.GetExtension("ComplexityLimit"); complexityExt != nil {
			if stats, ok := complexityExt.(*ComplexityStats); ok {
				complexity = stats.Complexity
			}
		}

		var err error
		if resp != nil && len(resp.Errors) > 0 {
			err = resp.Errors[0]
		}

		// Use the helper function for consistency
		metrics.RecordGraphQLOperation(opName, duration, complexity, err)

		return resp
	})
}
