package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/standard"
)

// TraverseStandard only returns public standards and standards owned by the organization
func TraverseStandard() ent.Interceptor {
	return intercept.TraverseStandard(func(ctx context.Context, q *generated.StandardQuery) error {
		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return err
		}

		// filter to return public standards and standards owned by the organization
		q.Where(
			standard.Or(
				standard.And(
					standard.OwnerIDIsNil(),
					standard.IsPublic(true),
					standard.SystemOwned(true),
				),
				standard.OwnerIDIn(orgIDs...),
			),
		)

		return nil
	})
}
