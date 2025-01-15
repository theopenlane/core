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

	isHistory := strings.Contains(q.Type(), "History")

	objectIDs, err := GetAuthorizedObjectIDs(ctx, q.Type())
	if err != nil {
		return err
	}

	// if the query is a history query, we need to filter by the ref field
	if isHistory {
		q.WhereP(
			sql.FieldIn("ref", objectIDs...),
		)

		return nil
	}

	// filter the query to only include the files that the user has access to
	q.WhereP(
		sql.FieldIn("id", objectIDs...),
	)

	return nil
}

// GetAuthorizedObjectIDs does a list objects request to pull all ids the current user
// has access to within the FGA system
func GetAuthorizedObjectIDs(ctx context.Context, queryType string) ([]string, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return []string{}, nil
	}

	// get the type of the query, removing the History suffix
	objectType := strings.Replace(queryType, "History", "", 1)

	req := fgax.ListRequest{
		SubjectID:   userID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		ObjectType:  strcase.SnakeCase(objectType),
	}

	if strings.Contains(queryType, "History") {
		log.Debug().Msg("adding history relation to list request")

		req.Relation = fgax.CanViewAuditLog
	}

	log.Info().Interface("req", req).Msg("getting authorized object ids")

	resp, err := utils.AuthzClientFromContext(ctx).ListObjectsRequest(ctx, req)
	if err != nil {
		return []string{}, nil
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
