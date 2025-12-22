package common //nolint:revive

import (
	"context"

	"ariga.io/entcache"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/vektah/gqlparser/v2/ast"
)

// WithFileUploader adds the file uploader to the graphql handler
// this will handle the file upload process for the multipart form
func WithFileUploader(h *handler.Server, u *objects.Service) {
	h.AroundFields(injectFileUploader(u))
}

// WithContextLevelCache adds a context level cache to the handler
func WithContextLevelCache(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		if op := graphql.GetOperationContext(ctx).Operation; op != nil && op.Operation == ast.Query {
			ctx = entcache.NewContext(ctx)
		}

		return next(ctx)
	})
}

// WithResultLimit adds a max result limit to the handler in order to set limits on
// all nested edges in the graphql request
func WithResultLimit(h *handler.Server, limit *int) {
	h.AroundFields(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
		if limit == nil {
			return next(ctx)
		}

		// grab preloads to set max result limits
		graphutils.GetPreloads(ctx, limit)

		return next(ctx)
	})
}

// WithSkipCache adds a skip cache middleware to the handler
// This is useful for testing, where you don't want to cache responses
// so you can see the changes immediately
func WithSkipCache(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		return next(entcache.Skip(ctx))
	})
}
