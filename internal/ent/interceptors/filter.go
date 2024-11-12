package interceptors

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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

	objectIDs, err := GetAuthorizedObjectIDs(ctx, q.Type())
	if err != nil {
		return err
	}

	// filter the query to only include the files that the user has access to
	q.WhereP(
		sql.FieldIn("id", objectIDs...),
	)

	return nil
}

// GetAuthorizedObjectIDs does a list objects request to pull all ids the current user
// has access to within the FGA system
func GetAuthorizedObjectIDs(ctx context.Context, objectType string) ([]string, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return []string{}, nil
	}

	req := fgax.ListRequest{
		SubjectID:   userID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		ObjectType:  strings.ToLower(objectType),
	}

	resp, err := generated.FromContext(ctx).Authz.ListObjectsRequest(ctx, req)
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
