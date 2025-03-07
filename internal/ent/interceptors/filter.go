package interceptors

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// FilterListQuery filters any list query to only include the objects that the user has access to
// This is automatically added to all schemas using the ObjectOwnedMixin, so should not be added
// directly if that mixin is used
func FilterListQuery() ent.Interceptor {
	return intercept.TraverseFunc(AddIDPredicate)
}

// AddIDPredicate adds a predicate to the query to only include the objects that the user has access to
// This is generally used by object owned setups with the ObjectOwnedMixin
func AddIDPredicate(ctx context.Context, q intercept.Query) error {
	// by pass checks on invite or pre-allowed request
	if _, allow := privacy.DecisionFromContext(ctx); allow {
		return nil
	}

	// Membership tables should use the object_id field,
	// e.g. GroupMembership should use group_id
	isMembership := strings.Contains(q.Type(), "Membership")

	// types that are filtered by another field than id
	// History uses `ref`
	isHistory := strings.Contains(q.Type(), "History")

	objectType := q.Type()
	if isMembership {
		objectType = strings.ReplaceAll(q.Type(), "Membership", "")
	}

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
		log.Debug().Msg("adding history relation to list request")

		req.Relation = fgax.CanViewAuditLog
	}

	log.Info().Interface("req", req).Msg("getting authorized object ids")

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
