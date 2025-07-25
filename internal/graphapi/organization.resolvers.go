package graphapi

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/iam/auth"
)

// CreateOrganization is the resolver for the createOrganization field.
func (r *mutationResolver) CreateOrganization(ctx context.Context, input generated.CreateOrganizationInput, avatarFile *graphql.Upload) (*model.OrganizationCreatePayload, error) {
	if auth.GetAuthTypeFromContext(ctx) != auth.JWTAuthentication {
		log.Debug().Msg("organization attempted to be created with non-JWT auth type")

		return nil, ErrResourceNotAccessibleWithToken
	}

	// set the parent organization in the auth context, used when creating a sub-organization with a personal access token
	if input.ParentID != nil {
		if err := setOrganizationInAuthContext(ctx, input.ParentID); err != nil {
			log.Error().Err(err).Msg("failed to set organization in auth context")

			return nil, newNotFoundError("parent_id")
		}
	}

	res, err := withTransactionalMutation(ctx).Organization.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "organization"})
	}

	return &model.OrganizationCreatePayload{
		Organization: res,
	}, nil
}

// UpdateOrganization is the resolver for the updateOrganization field.
func (r *mutationResolver) UpdateOrganization(ctx context.Context, id string, input generated.UpdateOrganizationInput, avatarFile *graphql.Upload) (*model.OrganizationUpdatePayload, error) {
	res, err := withTransactionalMutation(ctx).Organization.Get(ctx, id)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "organization"})
	}

	// setup update request
	req := res.Update().SetInput(input).AppendTags(input.AppendTags)

	res, err = req.Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "organization"})
	}

	return &model.OrganizationUpdatePayload{
		Organization: res,
	}, nil
}

// DeleteOrganization is the resolver for the deleteOrganization field.
func (r *mutationResolver) DeleteOrganization(ctx context.Context, id string) (*model.OrganizationDeletePayload, error) {
	if auth.GetAuthTypeFromContext(ctx) != auth.JWTAuthentication {
		log.Info().Msg("organization attempted to be deleted with non-JWT auth type")

		return nil, ErrResourceNotAccessibleWithToken
	}

	if err := withTransactionalMutation(ctx).Organization.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, parseRequestError(err, action{action: ActionDelete, object: "organization"})
	}

	if err := generated.OrganizationEdgeCleanup(ctx, id); err != nil {
		return nil, newCascadeDeleteError(err)
	}

	return &model.OrganizationDeletePayload{
		DeletedID: id,
	}, nil
}

// Organization is the resolver for the organization field.
func (r *queryResolver) Organization(ctx context.Context, id string) (*generated.Organization, error) {
	if err := setOrganizationInAuthContext(ctx, &id); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

		return nil, newNotFoundError("id")
	}

	query, err := withTransactionalMutation(ctx).Organization.Query().Where(organization.ID(id)).CollectFields(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "organization"})
	}

	res, err := query.Only(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "organization"})
	}

	return res, nil
}
