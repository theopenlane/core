package graphapi

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/utils/rout"
)

// CreateSubscriber is the resolver for the createSubscriber field.
func (r *mutationResolver) CreateSubscriber(ctx context.Context, input generated.CreateSubscriberInput) (*model.SubscriberCreatePayload, error) {
	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, input.OwnerID); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")
		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	sub, err := withTransactionalMutation(ctx).Subscriber.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "subscriber"})
	}

	return &model.SubscriberCreatePayload{Subscriber: sub}, nil
}

// CreateBulkSubscriber is the resolver for the createBulkSubscriber field.
func (r *mutationResolver) CreateBulkSubscriber(ctx context.Context, input []*generated.CreateSubscriberInput) (*model.SubscriberBulkCreatePayload, error) {
	return r.bulkCreateSubscriber(ctx, input)
}

// CreateBulkCSVSubscriber is the resolver for the createBulkCSVSubscriber field.
func (r *mutationResolver) CreateBulkCSVSubscriber(ctx context.Context, input graphql.Upload) (*model.SubscriberBulkCreatePayload, error) {
	subscriberInput, err := unmarshalBulkData[generated.CreateSubscriberInput](input)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal bulk data")

		return nil, err
	}

	return r.bulkCreateSubscriber(ctx, subscriberInput)
}

// UpdateSubscriber is the resolver for the updateSubscriber field.
func (r *mutationResolver) UpdateSubscriber(ctx context.Context, email string, input generated.UpdateSubscriberInput) (*model.SubscriberUpdatePayload, error) {
	subscriber, err := withTransactionalMutation(ctx).Subscriber.Query().
		Where(
			subscriber.EmailEQ(email),
		).Only(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "subscriber"})
	}

	if err := setOrganizationInAuthContext(ctx, &subscriber.OwnerID); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")
		return nil, rout.ErrPermissionDenied
	}

	subscriber, err = subscriber.Update().SetInput(input).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "subscriber"})
	}

	return &model.SubscriberUpdatePayload{Subscriber: subscriber}, nil
}

// DeleteSubscriber is the resolver for the deleteSubscriber field.
func (r *mutationResolver) DeleteSubscriber(ctx context.Context, email string, ownerID *string) (*model.SubscriberDeletePayload, error) {
	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, ownerID); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")
		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	num, err := withTransactionalMutation(ctx).Subscriber.Delete().
		Where(
			subscriber.EmailEQ(email),
		).Exec(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionDelete, object: "subscriber"})
	}

	if num == 0 {
		return nil, newNotFoundError("subscriber")
	}

	return &model.SubscriberDeletePayload{Email: email}, nil
}

// Subscriber is the resolver for the subscriber field.
func (r *queryResolver) Subscriber(ctx context.Context, email string) (*generated.Subscriber, error) {
	subscriber, err := withTransactionalMutation(ctx).Subscriber.Query().
		Where(
			subscriber.EmailEQ(email),
		).Only(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "subscriber"})
	}

	return subscriber, nil
}
