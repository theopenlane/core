package common //nolint:revive

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
)

// WithMetrics adds Prometheus instrumentation around GraphQL operations.
func WithMetrics(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		// call the next middleware to get the response
		resp := next(ctx)

		// get the operation context
		opCtx := graphql.GetOperationContext(ctx)

		opName := getOpName(ctx)

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

// getOpName retrieves the operation name from the context or returns "unknown" if not found
func getOpName(ctx context.Context) string {
	if !graphql.HasOperationContext(ctx) {
		logx.FromContext(ctx).Info().Msg("graphql operation context not found; unable to determine operation name")

		return "unknown"
	}

	opCtx := graphql.GetOperationContext(ctx)
	if opCtx.OperationName != "" {
		return opCtx.OperationName
	}

	if opCtx.Operation != nil {
		if opCtx.Operation.Name != "" {
			return opCtx.Operation.Name
		}

		if opCtx.Operation.Operation == ast.Mutation {
			return "unnamed_mutation"
		}

		if opCtx.Operation.Operation == ast.Query {
			return "unnamed_query"
		}
	}

	logx.FromContext(ctx).Info().Str("raw_query", opCtx.RawQuery).Msg("GraphQL operation name is empty; metrics may be obscured")

	return "unknown"
}
