package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// InterceptorOrganizationSetting is middleware to change the org setting query
func InterceptorOrganizationSetting() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return nil
		}
		// Organization list queries should not be filtered by organization id
		// Same with OrganizationSetting queries with the Only operation
		ctxQuery := ent.QueryFromContext(ctx)
		if ctxQuery.Type == generated.TypeOrganization || ctxQuery.Op == OnlyOperation {
			return nil
		}

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			return auth.ErrNoAuthUser
		}

		orgIDs := caller.OrgIDs()

		// sets the organization id on the query for the current organization
		q.WhereP(organizationsetting.OrganizationIDIn(orgIDs...))

		return nil
	})
}
