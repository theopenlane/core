package graphapi

import (
	"context"

	"github.com/99designs/gqlgen/complexity"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/theopenlane/core/pkg/metrics"
)

const complexityExtension = "ComplexityLimit"

// ComplexityStats represents the complexity calculation results
type ComplexityStats struct {
	// The calculated complexity for this request
	Complexity int

	// The complexity limit for this request returned by the extension func
	ComplexityLimit int
}

// complexityLimitWithMetrics is a custom complexity limit extension that records metrics
type complexityLimitWithMetrics struct {
	limitFunc func(ctx context.Context, rc *graphql.OperationContext) int
	es        graphql.ExecutableSchema
}

// newComplexityLimitWithMetrics creates a new complexity limit extension with metrics
func newComplexityLimitWithMetrics(limitFunc func(ctx context.Context, rc *graphql.OperationContext) int) *complexityLimitWithMetrics {
	return &complexityLimitWithMetrics{
		limitFunc: limitFunc,
	}
}

// ExtensionName returns the extension name
func (c *complexityLimitWithMetrics) ExtensionName() string {
	return complexityExtension
}

// Validate validates the schema and stores it for complexity calculations
func (c *complexityLimitWithMetrics) Validate(schema graphql.ExecutableSchema) error {
	c.es = schema
	return nil
}

// MutateOperationContext calculates complexity and enforces the limit
func (c *complexityLimitWithMetrics) MutateOperationContext(ctx context.Context, opCtx *graphql.OperationContext) *gqlerror.Error {
	op := opCtx.Doc.Operations.ForName(opCtx.OperationName)
	complexityCalc := complexity.Calculate(ctx, c.es, op, opCtx.Variables)

	limit := c.limitFunc(ctx, opCtx)

	// Store complexity in stats for later retrieval by metrics
	opCtx.Stats.SetExtension(complexityExtension, &ComplexityStats{
		Complexity:      complexityCalc,
		ComplexityLimit: limit,
	})

	// Check if complexity exceeds the limit
	if complexityCalc > limit {
		// Record the rejection
		metrics.RecordGraphQLRejection("complexity")

		return gqlerror.Errorf("operation has complexity %d, which exceeds the limit of %d", complexityCalc, limit)
	}

	return nil
}
