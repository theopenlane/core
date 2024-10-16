package graphapi

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// errorResponse returns a graphql response with a single error
func errorResponse(err error) func(ctx context.Context) *graphql.Response {
	return func(ctx context.Context) *graphql.Response {
		return &graphql.Response{
			Errors: gqlerror.List{
				{
					Message: err.Error(),
				},
			},
		}
	}
}
