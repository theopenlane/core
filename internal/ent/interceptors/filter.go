package interceptors

import (
	"context"
	"encoding/json"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// FilterListQuery filters any list query to only include the objects that the user has access to
// This is automatically added to all schemas using the ObjectOwnedMixin, so should not be added
// directly if that mixin is used
// This function is intended to filter the query before it is run using the ListObjectsRequest
// and should not be used for large lists
func FilterListQuery() ent.Interceptor {
	return intercept.TraverseFunc(AddIDPredicate)
}

// AddIDPredicate adds a predicate to the query to only include the objects that the user has access to
// This should only be used for queries where we are not directly filtering on the `id` field of the object
// e.g. memberships and history tables, and when there are a limited number of objects to filter
// the FilterQueryResults function should be used in most cases due to performance issues of ListObjectsRequest
func AddIDPredicate(ctx context.Context, q intercept.Query) error {
	// by pass checks on invite or pre-allowed request
	if _, allow := privacy.DecisionFromContext(ctx); allow || rule.IsInternalRequest(ctx) {
		return nil
	}

	// Membership tables should use the object_id field,
	// e.g. GroupMembership should use group_id
	isMembership := strings.Contains(q.Type(), "Membership")

	// types that are filtered by another field than id
	// History uses `ref`
	isHistory := strings.Contains(q.Type(), "History")

	objectType := getFGAObjectType(q)

	objectIDs, err := GetAuthorizedObjectIDs(ctx, objectType)
	if err != nil {
		return err
	}

	switch {
	case isHistory:
		q.WhereP(
			sql.FieldIn("ref", objectIDs...),
		)

	case isMembership:
		column := strings.ToLower(objectType) + "_id"
		q.WhereP(
			sql.FieldIn(column, objectIDs...),
		)
	default:
		q.WhereP(
			sql.FieldIn("id", objectIDs...),
		)
	}

	return nil
}

// GetAuthorizedObjectIDs does a list objects request to pull all ids the current user
// has access to within the FGA system
func GetAuthorizedObjectIDs(ctx context.Context, queryType string) ([]string, error) {
	user, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return []string{}, nil
	}

	// get the type of the query, removing the History suffix
	objectType := strings.Replace(queryType, "History", "", 1)

	req := fgax.ListRequest{
		SubjectID:   user.SubjectID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		ObjectType:  strcase.SnakeCase(objectType),
		// add email domain to satisfy any list requests with organization conditions
		ConditionContext: utils.NewOrganizationContextKey(user.SubjectEmail),
	}

	if strings.Contains(queryType, "History") {
		zerolog.Ctx(ctx).Debug().Msg("adding history relation to list request")

		req.Relation = fgax.CanViewAuditLog
	}

	zerolog.Ctx(ctx).Info().Interface("req", req).Msg("getting authorized object ids")

	resp, err := utils.AuthzClientFromContext(ctx).ListObjectsRequest(ctx, req)
	if err != nil {
		return []string{}, err
	}

	objectIDs := make([]string, 0, len(resp.Objects))

	for _, obj := range resp.Objects {
		entity, err := fgax.ParseEntity(obj)
		if err != nil {
			return []string{}, nil
		}

		objectIDs = append(objectIDs, entity.Identifier)
	}

	return objectIDs, nil
}

// FilterQueryResults filters the results of a query to only include the objects that the user has access to
// This is automatically added to all schemas using the ObjectOwnedMixin, so should not be added
// directly if that mixin is used
// This function is intended to filter results after the query is run using the BatchCheck in FGA
// which is more performant than the ListObjectsRequest, especially for large lists
func FilterQueryResults[V any]() ent.InterceptFunc {
	return func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			return filterQueryResults[V](ctx, query, next)
		})
	}
}

// filterQueryResults filters the results of a query to only include the objects that the user has access to
// using the BatchCheck in FGA and returns the filtered results as the ent.Value based on the provided type
func filterQueryResults[V any](ctx context.Context, query ent.Query, next ent.Querier) (ent.Value, error) {
	// by pass checks on invite or pre-allowed request
	if _, allow := privacy.DecisionFromContext(ctx); allow {
		return next.Query(ctx, query)
	}

	v, err := next.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// convert the query to an intercept query
	q, err := intercept.NewQuery(query)
	if err != nil {
		return nil, err
	}

	ctxQuery := ent.QueryFromContext(ctx)

	switch ctxQuery.Op {
	case ent.OpQueryCount:
		// nothing to filter if we're just counting
		return v, nil
	case ent.OpQueryIDs, ent.OpQueryFirstID:
		ids, ok := v.([]string)
		if !ok {
			log.Error().Str("query_type", q.Type()).Msgf("failed to cast query results to expected slice %T", v)

			return nil, ErrRetrievingObjects
		}

		return filterIDList(ctx, ids, getFGAObjectType(q))
	case ent.OpQueryOnlyID:
		allow, err := singleIDCheck(ctx, v, getFGAObjectType(q))
		if err != nil {
			return nil, err
		}

		if !allow {
			return nil, nil
		}

		return v, nil

	default:
		switch t := v.(type) {
		case []*V:
			return filterListObjects[V](ctx, t, q)
		case *V:
			return singleObjectCheck[V](ctx, t, q)
		default:
			// non-standard query results, return as is
			return v, nil
		}
	}
}

// filterIDList filters a list of object ids to only include the objects that the user has access to
func filterIDList(ctx context.Context, ids []string, objectType string) ([]string, error) {
	allowedIDs, err := filterAuthorizedObjectIDs(ctx, objectType, ids)
	if err != nil {
		return nil, err
	}

	return allowedIDs, nil
}

// singleIDCheck checks if a single object id is allowed and returns a boolean
func singleIDCheck(ctx context.Context, v ent.Value, objectType string) (bool, error) {
	id, ok := v.(string)
	if !ok {
		log.Error().Msgf("failed to cast query results to expected single ID %T", v)

		return false, ErrRetrievingObjects
	}

	allowedIDs, err := filterIDList(ctx, []string{id}, objectType)
	if err != nil {
		return false, err
	}

	if len(allowedIDs) == 0 {
		return false, nil
	}

	return true, nil
}

// filterListObjects filters a list of objects to only include the objects that the user has access to
// and returns the filtered list as the ent.Value
func filterListObjects[T any](ctx context.Context, v ent.Value, q intercept.Query) (ent.Value, error) {
	listResults := v.([]*T)
	if len(listResults) == 0 {
		return v, nil
	}

	objectIDs, err := getObjectIDsFromEntValues(v)
	if err != nil {
		return nil, err
	}

	allowedIDs, err := filterAuthorizedObjectIDs(ctx, getFGAObjectType(q), objectIDs)
	if err != nil {
		return nil, err
	}

	// if no results are allowed, return an empty list
	if len(allowedIDs) == 0 {
		return make([]*T, 0), nil
	}

	// if all the results are allowed, return early
	if len(allowedIDs) == len(objectIDs) {
		return v, nil
	}

	// filter the results based on the allowed ids
	filteredResults := make([]*T, 0, len(allowedIDs))

	for _, id := range allowedIDs {
		for _, item := range listResults {
			objID, err := getObjectIDFromEntValue(item)
			if err != nil {
				return nil, err
			}

			if id == objID {
				filteredResults = append(filteredResults, item)
				break
			}
		}
	}

	// return the filtered results
	return filteredResults, nil
}

// singleObjectCheck checks if a single object is allowed and returns the object if it is
func singleObjectCheck[T any](ctx context.Context, v ent.Value, q intercept.Query) (ent.Value, error) {
	objectIDs, err := getObjectIDsFromEntValues(v)
	if err != nil {
		return nil, err
	}

	allowedIDs, err := filterAuthorizedObjectIDs(ctx, getFGAObjectType(q), objectIDs)
	if err != nil {
		return nil, err
	}

	// if the query is a single result query, we don't need to filter, if it was allowed
	// it would have been returned
	if len(allowedIDs) == 0 {
		return nil, nil
	}

	return v, nil
}

// getFGAObjectType returns the object type for the query
// for membership tables, it will return the type with the membership suffix removed
// e.g. GroupMembership -> Group
func getFGAObjectType(q intercept.Query) string {
	// Membership tables should use the object_id field,
	// e.g. GroupMembership should use group_id
	isMembership := strings.Contains(q.Type(), "Membership")

	objectType := q.Type()
	if isMembership {
		objectType = strings.ReplaceAll(q.Type(), "Membership", "")
	}

	return objectType
}

// getObjectIDFromEntValues extracts the object id from a generic ent value (used for list queries)
// this function should be called after the query has been successful to get the returned object ids
func getObjectIDsFromEntValues(m ent.Value) ([]string, error) {
	type objectIDer struct {
		ID string `json:"id"`
	}

	tmp, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	var results []objectIDer
	if err := json.Unmarshal(tmp, &results); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(results))
	for _, d := range results {
		ids = append(ids, d.ID)
	}

	return ids, nil
}

// getObjectIDFromEntValue extracts the object id from a generic ent value (used for single object
// queries) this function should be called after the query has been successful to get the returned object ids from the object
func getObjectIDFromEntValue(m ent.Value) (string, error) {
	type objectIDer struct {
		ID string `json:"id"`
	}

	tmp, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	var res objectIDer
	if err := json.Unmarshal(tmp, &res); err != nil {
		return "", err
	}

	return res.ID, nil
}

// filterAuthorizedObjectIDs takes all the object ids returned from a query and will filter the results
// this is intended to be used in place of GetAuthorizedObjectIDs when you already have the object ids
// and just need to filter them based on the user's permissions
func filterAuthorizedObjectIDs(ctx context.Context, objectType string, objectIDs []string) ([]string, error) {
	user, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return []string{}, nil
	}

	checks := []fgax.AccessCheck{}

	for _, id := range objectIDs {
		ac := fgax.AccessCheck{
			SubjectID:   user.SubjectID,
			SubjectType: auth.GetAuthzSubjectType(ctx),
			ObjectID:    id,
			ObjectType:  fgax.Kind(strcase.SnakeCase(objectType)), // convert to snake case e.g. InternalPolicy -> internal_policy
			Relation:    fgax.CanView,
			Context:     utils.NewOrganizationContextKey(user.SubjectEmail), // required for any check that goes back up to the organization
		}

		checks = append(checks, ac)
	}

	allowedIDs, err := utils.AuthzClientFromContext(ctx).BatchCheckObjectAccess(ctx, checks)
	if err != nil {
		return []string{}, err
	}

	return allowedIDs, nil
}
