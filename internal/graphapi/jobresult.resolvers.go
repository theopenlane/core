package graphapi

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/jobresult"
)

// JobResult is the resolver for the jobResult field.
func (r *queryResolver) JobResult(ctx context.Context, id string) (*generated.JobResult, error) {
	query, err := withTransactionalMutation(ctx).JobResult.Query().Where(jobresult.ID(id)).CollectFields(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "jobresult"})
	}

	res, err := query.Only(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "jobresult"})
	}

	return res, nil
}
