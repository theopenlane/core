package graphapi

import (
	"context"
	"time"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/ent/generated/user"
)

var (
	maxSearchTime = time.Duration(30 * time.Second)
)

// searchResult is a generic struct to hold the result of a search operation
type searchResult[T any] struct {
	result T
	err    error
}

// searchOrganizations searches for organizations based on the query string looking for matches in the name, description and display name
func searchOrganizations(ctx context.Context, query string) ([]*generated.Organization, error) {
	return withTransactionalMutation(ctx).Organization.Query().Where(
		organization.Or(
			organization.NameContains(query),        // search by name
			organization.DescriptionContains(query), // search by description
			organization.DisplayNameContains(query), // search by display name
		),
	).All(ctx)
}

// searchGroups searches for groups based on the query string looking for matches in the name, description and display name
func searchGroups(ctx context.Context, query string) ([]*generated.Group, error) {
	return withTransactionalMutation(ctx).Group.Query().Where(
		group.Or(
			group.NameContains(query),        // search by name
			group.DescriptionContains(query), // search by description
			group.DisplayNameContains(query), // search by display name
		),
	).All(ctx)
}

// searchUsers searches for org members based on the query string looking for matches in the email, display name, first name and last name
func searchUsers(ctx context.Context, query string) ([]*generated.User, error) {
	members, err := withTransactionalMutation(ctx).OrgMembership.Query().Where(
		orgmembership.Or(
			orgmembership.HasUserWith(user.EmailContains(query)),       // search by email
			orgmembership.HasUserWith(user.DisplayNameContains(query)), // search by display name
			orgmembership.HasUserWith(user.FirstNameContains(query)),   // search by first name
			orgmembership.HasUserWith(user.LastNameContains(query)),    // search by last name
		),
	).WithUser().All(ctx)

	if members == nil || err != nil {
		return nil, err
	}

	users := make([]*generated.User, 0, len(members))
	for _, member := range members {
		users = append(users, member.Edges.User)
	}

	return users, err
}

// searchSubscriber searches for subscribers based on the query string looking for matches in the email
func searchSubscriber(ctx context.Context, query string) ([]*generated.Subscriber, error) {
	return withTransactionalMutation(ctx).Subscriber.Query().Where(
		subscriber.Or(
			subscriber.EmailContains(query), // search by email
		),
	).All(ctx)
}
