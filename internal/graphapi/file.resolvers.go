package graphapi

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/model"
)

// DeleteFile is the resolver for the deleteFile field.
func (r *mutationResolver) DeleteFile(ctx context.Context, id string) (*model.FileDeletePayload, error) {
	if err := withTransactionalMutation(ctx).File.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, parseRequestError(err, action{action: ActionDelete, object: "file"})
	}

	if err := generated.FileEdgeCleanup(ctx, id); err != nil {
		return nil, newCascadeDeleteError(err)
	}

	return &model.FileDeletePayload{
		DeletedID: id,
	}, nil
}

// File is the resolver for the file field.
func (r *queryResolver) File(ctx context.Context, id string) (*generated.File, error) {
	res, err := withTransactionalMutation(ctx).File.Get(ctx, id)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "file"})
	}

	return res, nil
}
