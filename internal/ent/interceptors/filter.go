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
	return intercept.TraverseFunc(addIDPredicate)
}

func addIDPredicate(ctx context.Context, q intercept.Query) error {
	// by pass checks on invite or pre-allowed request
	if _, allow := privacy.DecisionFromContext(ctx); allow {
		return nil
	}

	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	req := fgax.ListRequest{
		SubjectID:  userID,
		ObjectType: strings.ToLower(q.Type()),
	}

	resp, err := generated.FromContext(ctx).Authz.ListObjectsRequest(ctx, req)
	if err != nil {
		return err
	}

	objectIDs := make([]string, 0, len(resp.Objects))

	for _, obj := range resp.Objects {
		entity, err := fgax.ParseEntity(obj)
		if err != nil {
			return err
		}

		objectIDs = append(objectIDs, entity.Identifier)
	}

	// filter the query to only include the files that the user has access to
	q.WhereP(
		sql.FieldIn("id", objectIDs...),
	)

	return nil
}
