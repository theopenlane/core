package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
)

// InterceptorOrganizationSetting is middleware to change the org setting query
func InterceptorOrganizationSetting() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// Organization list queries should not be filtered by organization id
		// Same with OrganizationSetting queries with the Only operation
		ctxQuery := ent.QueryFromContext(ctx)
		if ctxQuery.Type == generated.TypeOrganization || ctxQuery.Op == OnlyOperation {
			return nil
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return err
		}

		if len(orgIDs) == 0 {
			orgID, err := auth.GetOrganizationIDFromContext(ctx)
			if err != nil {
				return err
			}

			orgIDs = append(orgIDs, orgID)
		}

		// sets the organization id on the query for the current organization
		q.WhereP(organizationsetting.OrganizationIDIn(orgIDs...))

		return nil
	})
}
