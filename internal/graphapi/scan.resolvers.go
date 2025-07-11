package graphapi

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/scan"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/utils/rout"
)

// CreateScan is the resolver for the createScan field.
func (r *mutationResolver) CreateScan(ctx context.Context, input generated.CreateScanInput) (*model.ScanCreatePayload, error) {
	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, input.OwnerID); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	res, err := withTransactionalMutation(ctx).Scan.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "scan"})
	}

	return &model.ScanCreatePayload{
		Scan: res,
	}, nil
}

// CreateBulkScan is the resolver for the createBulkScan field.
func (r *mutationResolver) CreateBulkScan(ctx context.Context, input []*generated.CreateScanInput) (*model.ScanBulkCreatePayload, error) {
	if len(input) == 0 {
		return nil, rout.NewMissingRequiredFieldError("input")
	}

	// set the organization in the auth context if its not done for us
	// this will choose the first input OwnerID when using a personal access token
	if err := setOrganizationInAuthContextBulkRequest(ctx, input); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	return r.bulkCreateScan(ctx, input)
}

// CreateBulkCSVScan is the resolver for the createBulkCSVScan field.
func (r *mutationResolver) CreateBulkCSVScan(ctx context.Context, input graphql.Upload) (*model.ScanBulkCreatePayload, error) {
	data, err := unmarshalBulkData[generated.CreateScanInput](input)
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

	return r.bulkCreateScan(ctx, data)
}

// UpdateScan is the resolver for the updateScan field.
func (r *mutationResolver) UpdateScan(ctx context.Context, id string, input generated.UpdateScanInput) (*model.ScanUpdatePayload, error) {
	res, err := withTransactionalMutation(ctx).Scan.Get(ctx, id)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "scan"})
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
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "scan"})
	}

	return &model.ScanUpdatePayload{
		Scan: res,
	}, nil
}

// DeleteScan is the resolver for the deleteScan field.
func (r *mutationResolver) DeleteScan(ctx context.Context, id string) (*model.ScanDeletePayload, error) {
	if err := withTransactionalMutation(ctx).Scan.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, parseRequestError(err, action{action: ActionDelete, object: "scan"})
	}

	if err := generated.ScanEdgeCleanup(ctx, id); err != nil {
		return nil, newCascadeDeleteError(err)
	}

	return &model.ScanDeletePayload{
		DeletedID: id,
	}, nil
}

// Scan is the resolver for the scan field.
func (r *queryResolver) Scan(ctx context.Context, id string) (*generated.Scan, error) {
	query, err := withTransactionalMutation(ctx).Scan.Query().Where(scan.ID(id)).CollectFields(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "scan"})
	}

	res, err := query.Only(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "scan"})
	}

	return res, nil
}
