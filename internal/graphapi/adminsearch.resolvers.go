package graphapi

// THIS CODE IS REGENERATED BY github.com/theopenlane/gqlgen-plugins. DO NOT EDIT.

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
)

// Search is the resolver for the search field.
func (r *queryResolver) AdminSearch(ctx context.Context, query string) (*SearchResultConnection, error) {
	var (
		errors              []error
		contactResults      []*generated.Contact
		entityResults       []*generated.Entity
		groupResults        []*generated.Group
		organizationResults []*generated.Organization
		subscriberResults   []*generated.Subscriber
	)

	r.withPool().SubmitMultipleAndWait([]func(){
		func() {
			var err error
			contactResults, err = searchContacts(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			entityResults, err = searchEntities(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			groupResults, err = searchGroups(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			organizationResults, err = searchOrganizations(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			subscriberResults, err = searchSubscribers(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
	})

	// Check all errors and return a single error if any of the searches failed
	if len(errors) > 0 {
		r.logger.Errorw("search failed", "errors", errors)

		return nil, ErrSearchFailed
	}

	// return the results
	return &SearchResultConnection{
		Nodes: []SearchResult{
			ContactSearchResult{
				Contacts: contactResults,
			},
			EntitySearchResult{
				Entities: entityResults,
			},
			GroupSearchResult{
				Groups: groupResults,
			},
			OrganizationSearchResult{
				Organizations: organizationResults,
			},
			SubscriberSearchResult{
				Subscribers: subscriberResults,
			},
		},
	}, nil
}
func (r *queryResolver) AdminContactSearch(ctx context.Context, query string) (*ContactSearchResult, error) {
	contactResults, err := adminSearchContacts(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &ContactSearchResult{
		Contacts: contactResults,
	}, nil
}
func (r *queryResolver) AdminEntitySearch(ctx context.Context, query string) (*EntitySearchResult, error) {
	entityResults, err := adminSearchEntities(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &EntitySearchResult{
		Entities: entityResults,
	}, nil
}
func (r *queryResolver) AdminGroupSearch(ctx context.Context, query string) (*GroupSearchResult, error) {
	groupResults, err := adminSearchGroups(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &GroupSearchResult{
		Groups: groupResults,
	}, nil
}
func (r *queryResolver) AdminOrganizationSearch(ctx context.Context, query string) (*OrganizationSearchResult, error) {
	organizationResults, err := adminSearchOrganizations(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &OrganizationSearchResult{
		Organizations: organizationResults,
	}, nil
}
func (r *queryResolver) AdminSubscriberSearch(ctx context.Context, query string) (*SubscriberSearchResult, error) {
	subscriberResults, err := adminSearchSubscribers(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &SubscriberSearchResult{
		Subscribers: subscriberResults,
	}, nil
}