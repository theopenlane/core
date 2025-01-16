package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/hooks"
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

	switch v := v.(type) {
	case []string:
		for _, g := range v {
			groups = append(groups, &generated.Group{ID: g})
		}
	case string:
		groups = append(groups, &generated.Group{ID: v})
	case []*generated.Group:
		groups = v
	case *generated.Group:
		groups = append(groups, v)
	default:
		log.Error().Msg("unexpected type for group query")

		return nil, ErrInternalServerError
	}

	_, managedGroup := contextx.From[hooks.ManagedContextKey](ctx)

	if qc.Op == OnlyOperation && managedGroup {
		return v.([]*generated.Group), nil
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
