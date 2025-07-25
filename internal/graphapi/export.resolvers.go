package graphapi

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/export"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/utils/rout"
)

// CreateExport is the resolver for the createExport field.
func (r *mutationResolver) CreateExport(ctx context.Context, input generated.CreateExportInput) (*model.ExportCreatePayload, error) {
	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, input.OwnerID); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	res, err := withTransactionalMutation(ctx).Export.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "export"})
	}

	return &model.ExportCreatePayload{
		Export: res,
	}, nil
}

// UpdateExport is the resolver for the updateExport field.
func (r *mutationResolver) UpdateExport(ctx context.Context, id string, input generated.UpdateExportInput, exportFiles []*graphql.Upload) (*model.ExportUpdatePayload, error) {
	res, err := withTransactionalMutation(ctx).Export.Get(ctx, id)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "export"})
	}

	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, &res.OwnerID); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.ErrPermissionDenied
	}

	// setup update request
	req := res.Update().SetInput(input)

	res, err = req.Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "export"})
	}

	return &model.ExportUpdatePayload{
		Export: res,
	}, nil
}

// DeleteExport is the resolver for the deleteExport field.
func (r *mutationResolver) DeleteExport(ctx context.Context, id string) (*model.ExportDeletePayload, error) {
	if err := withTransactionalMutation(ctx).Export.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, parseRequestError(err, action{action: ActionDelete, object: "export"})
	}

	if err := generated.ExportEdgeCleanup(ctx, id); err != nil {
		return nil, newCascadeDeleteError(err)
	}

	return &model.ExportDeletePayload{
		DeletedID: id,
	}, nil
}

// DeleteBulkExport is the resolver for the deleteBulkExport field.
func (r *mutationResolver) DeleteBulkExport(ctx context.Context, ids []string) (*model.ExportBulkDeletePayload, error) {
	if len(ids) == 0 {
		return nil, rout.NewMissingRequiredFieldError("ids")
	}

	return r.bulkDeleteExport(ctx, ids)
}

// Export is the resolver for the export field.
func (r *queryResolver) Export(ctx context.Context, id string) (*generated.Export, error) {
	query, err := withTransactionalMutation(ctx).Export.Query().Where(export.ID(id)).CollectFields(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "export"})
	}

	res, err := query.Only(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "export"})
	}

	return res, nil
}
