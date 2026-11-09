package graphapi

import (
	"context"
	"strings"

	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/logx"
)

const (
	bulkCreateFieldPrefix    = "createBulk"
	bulkCreateCSVFieldPrefix = "createBulkCSV"
)

// WithTransactions adds the transactioner to the ent db client
func WithTransactions(h *handler.Server, d *ent.Client) {
	// setup transactional db client
	h.AroundOperations(injectClient(d))

	h.Use(entgql.Transactioner{TxOpener: d})

	h.AroundFields(batchBulkAuthz)
}

// batchBulkAuthz flushes batched tuples after the resolver, before payload fields and the commit
func batchBulkAuthz(ctx context.Context, next graphql.Resolver) (any, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil || fc.Object != "Mutation" {
		return next(ctx)
	}

	objectType, ok := bulkAuthzObjectType(fc.Field.Name)
	if !ok {
		return next(ctx)
	}

	ctx, batch := utils.WithBulkAuthz(ctx, objectType)

	res, err := next(ctx)
	if err != nil {
		return res, err
	}

	if err := batch.Flush(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to write batched relationship tuples")

		return nil, hooks.ErrInternalServerError
	}

	return res, nil
}

// bulkAuthzObjectType returns the object type of a bulk create field
func bulkAuthzObjectType(field string) (string, bool) {
	if objectType, ok := strings.CutPrefix(field, bulkCreateCSVFieldPrefix); ok {
		return objectType, true
	}

	return strings.CutPrefix(field, bulkCreateFieldPrefix)
}
