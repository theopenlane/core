package graphapi

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/utils/rout"
)

// CreateMappedControl is the resolver for the createMappedControl field.
func (r *mutationResolver) CreateMappedControl(ctx context.Context, input generated.CreateMappedControlInput) (*model.MappedControlCreatePayload, error) {
	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, input.OwnerID); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	res, err := withTransactionalMutation(ctx).MappedControl.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "mappedcontrol"})
	}

	return &model.MappedControlCreatePayload{
		MappedControl: res,
	}, nil
}

// CreateBulkMappedControl is the resolver for the createBulkMappedControl field.
func (r *mutationResolver) CreateBulkMappedControl(ctx context.Context, input []*generated.CreateMappedControlInput) (*model.MappedControlBulkCreatePayload, error) {
	if len(input) == 0 {
		return nil, rout.NewMissingRequiredFieldError("input")
	}

	// set the organization in the auth context if its not done for us
	// this will choose the first input OwnerID when using a personal access token
	if err := setOrganizationInAuthContextBulkRequest(ctx, input); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	return r.bulkCreateMappedControl(ctx, input)
}

// CreateBulkCSVMappedControl is the resolver for the createBulkCSVMappedControl field.
func (r *mutationResolver) CreateBulkCSVMappedControl(ctx context.Context, input graphql.Upload) (*model.MappedControlBulkCreatePayload, error) {
	data, err := unmarshalBulkData[generated.CreateMappedControlInput](input)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal bulk data")

		return nil, err
	}

	if len(data) == 0 {
		return nil, rout.NewMissingRequiredFieldError("input")
	}

	// set the organization in the auth context if its not done for us
	// this will choose the first input OwnerID when using a personal access token
	if err := setOrganizationInAuthContextBulkRequest(ctx, data); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	return r.bulkCreateMappedControl(ctx, data)
}

// UpdateMappedControl is the resolver for the updateMappedControl field.
func (r *mutationResolver) UpdateMappedControl(ctx context.Context, id string, input generated.UpdateMappedControlInput) (*model.MappedControlUpdatePayload, error) {
	res, err := withTransactionalMutation(ctx).MappedControl.Get(ctx, id)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "mappedcontrol"})
	}

	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, &res.OwnerID); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.ErrPermissionDenied
	}

	// setup update request
	req := res.Update().SetInput(input).AppendTags(input.AppendTags)

	res, err = req.Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "mappedcontrol"})
	}

	return &model.MappedControlUpdatePayload{
		MappedControl: res,
	}, nil
}

// DeleteMappedControl is the resolver for the deleteMappedControl field.
func (r *mutationResolver) DeleteMappedControl(ctx context.Context, id string) (*model.MappedControlDeletePayload, error) {
	if err := withTransactionalMutation(ctx).MappedControl.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, parseRequestError(err, action{action: ActionDelete, object: "mappedcontrol"})
	}

	if err := generated.MappedControlEdgeCleanup(ctx, id); err != nil {
		return nil, newCascadeDeleteError(err)
	}

	return &model.MappedControlDeletePayload{
		DeletedID: id,
	}, nil
}

// MappedControl is the resolver for the mappedControl field.
func (r *queryResolver) MappedControl(ctx context.Context, id string) (*generated.MappedControl, error) {
	query, err := withTransactionalMutation(ctx).MappedControl.Query().Where(mappedcontrol.ID(id)).CollectFields(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "mappedcontrol"})
	}

	res, err := query.Only(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "mappedcontrol"})
	}

	return res, nil
}
