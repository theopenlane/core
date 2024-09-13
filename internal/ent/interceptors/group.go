package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
)

// InterceptorGroup is middleware to change the Group query
func InterceptorGroup() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.GroupFunc(func(ctx context.Context, q *generated.GroupQuery) (generated.Value, error) {
			// run the query
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			return filterGroupsByAccess(ctx, q, v)
		})
	})
}

// filterGroupsByAccess checks fga, using ListObjects, and ensure user has view access to a group before it is returned
func filterGroupsByAccess(ctx context.Context, q *generated.GroupQuery, v ent.Value) ([]*generated.Group, error) {
	log.Debug().Msg("intercepting list group query")

	// return early if no groups
	if v == nil {
		return nil, nil
	}

	qc := ent.QueryFromContext(ctx)

	var groups []*generated.Group

	// check if query is for an exists query, which returns a slice of group ids
	// instead of the group objects
	switch qc.Op {
	case ExistOperation, IDsOperation:
		groupIDs, ok := v.([]string)
		if !ok {
			log.Error().Msg("unexpected type for group query")

			return nil, ErrInternalServerError
		}

		for _, g := range groupIDs {
			groups = append(groups, &generated.Group{ID: g})
		}
	default:
		var ok bool

		groups, ok = v.([]*generated.Group)
		if !ok {
			log.Error().Msg("unexpected type for group query")

			return nil, ErrInternalServerError
		}
	}

	// get userID for tuple checks
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Msg("unable to get user id from context")
		return nil, err
	}

	// See all groups user has view access
	req := fgax.ListRequest{
		SubjectID:   userID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		ObjectType:  "group",
	}

	groupList, err := q.Authz.ListObjectsRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	userGroups := groupList.GetObjects()

	var accessibleGroups []*generated.Group

	for _, g := range groups {
		entityType := "group"

		if !fgax.ListContains(entityType, userGroups, g.ID) {
			log.Info().Str("group_id", g.ID).
				Str("relation", fgax.CanView).
				Str("type", entityType).
				Msg("access denied to group")

			continue
		}

		// add group to accessible group
		accessibleGroups = append(accessibleGroups, g)
	}

	// return updated groups
	return accessibleGroups, nil
}
