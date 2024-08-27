package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/pkg/auth"
)

// InterceptorOrganizationSetting is middleware to change the org setting query
func InterceptorOrganizationSetting() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// Organization list queries should not be filtered by organization id
		// Same with OrganizationSetting queries with the Only operation
		ctxQuery := ent.QueryFromContext(ctx)
		if ctxQuery.Type == "Organization" || ctxQuery.Op == "Only" {
			return nil
		}

		orgID, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil {
			return err
		}

		// sets the organization id on the query for the current organization
		q.WhereP(organizationsetting.OrganizationID(orgID))

		return nil
	})
}
