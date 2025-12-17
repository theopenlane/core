package graphapi

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	ent "github.com/theopenlane/core/internal/ent/generated"
)

// common.WithTransactionalMutation automatically wrap the GraphQL mutations with a database transaction.
// This allows the ent.Client to commit at the end, or rollback the transaction in case of a GraphQL error.
func withTransactionalMutation(ctx context.Context) *ent.Client {
	return ent.FromContext(ctx)
}

// injectClient adds the db client to the context to be used with transactional mutations
func injectClient(db *ent.Client) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		ctx = ent.NewContext(ctx, db)
		return next(ctx)
	}
}

// extendedMappedControlInput is an extended version of the CreateMappedControlInput that is used for
// bulk operations and CSV uploads to allow mapping by reference codes
// when using the regular create operations the resolvers will parse the ref codes directly
type ExtendedMappedControlInput struct {
	ent.CreateMappedControlInput

	FromControlRefCodes    []string `json:"fromControlRefCodes"`
	FromSubcontrolRefCodes []string `json:"fromSubcontrolRefCodes"`
	ToControlRefCodes      []string `json:"toControlRefCodes"`
	ToSubcontrolRefCodes   []string `json:"toSubcontrolRefCodes"`
}
