package graphapi

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/utils/rout"
)

// CreateGroup is the resolver for the createGroup field.
func (r *mutationResolver) CreateGroup(ctx context.Context, input generated.CreateGroupInput) (*GroupCreatePayload, error) {
	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, input.OwnerID); err != nil {
		r.logger.Errorw("failed to set organization in auth context", "error", err)

		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	res, err := withTransactionalMutation(ctx).Group.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "group"}, r.logger)
	}

	return &GroupCreatePayload{
		Group: res,
	}, nil
}

// CreateBulkGroup is the resolver for the createBulkGroup field.
func (r *mutationResolver) CreateBulkGroup(ctx context.Context, input []*generated.CreateGroupInput) (*GroupBulkCreatePayload, error) {
	return r.bulkCreateGroup(ctx, input)
}

// CreateBulkCSVGroup is the resolver for the createBulkCSVGroup field.
func (r *mutationResolver) CreateBulkCSVGroup(ctx context.Context, input graphql.Upload) (*GroupBulkCreatePayload, error) {
	data, err := unmarshalBulkData[generated.CreateGroupInput](input)
	if err != nil {
		r.logger.Errorw("failed to unmarshal bulk data", "error", err)

		return nil, err
	}

	return r.bulkCreateGroup(ctx, data)
}

// UpdateGroup is the resolver for the updateGroup field.
func (r *mutationResolver) UpdateGroup(ctx context.Context, id string, input generated.UpdateGroupInput) (*GroupUpdatePayload, error) {
	res, err := withTransactionalMutation(ctx).Group.Get(ctx, id)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "group"}, r.logger)
	}
	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, &res.OwnerID); err != nil {
		r.logger.Errorw("failed to set organization in auth context", "error", err)

		return nil, ErrPermissionDenied
	}

	// setup update request
	req := res.Update().SetInput(input).AppendTags(input.AppendTags)

	res, err = req.Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "group"}, r.logger)
	}

	return &GroupUpdatePayload{
		Group: res,
	}, nil
}

// DeleteGroup is the resolver for the deleteGroup field.
func (r *mutationResolver) DeleteGroup(ctx context.Context, id string) (*GroupDeletePayload, error) {
	if err := withTransactionalMutation(ctx).Group.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, parseRequestError(err, action{action: ActionDelete, object: "group"}, r.logger)
	}

	if err := generated.GroupEdgeCleanup(ctx, id); err != nil {
		return nil, newCascadeDeleteError(err)
	}

	return &GroupDeletePayload{
		DeletedID: id,
	}, nil
}

// Group is the resolver for the group field.
func (r *queryResolver) Group(ctx context.Context, id string) (*generated.Group, error) {
	res, err := withTransactionalMutation(ctx).Group.Get(ctx, id)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "group"}, r.logger)
	}

	return res, nil
}