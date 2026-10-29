package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/intercept"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationbrandingsetting"
)

// InterceptorOrganizationBrandingSetting is middleware to change the org branding setting query
func InterceptorOrganizationBrandingSetting() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if auth.HasCrossOrgCapabilities(ctx) {
			return nil
		}

		ctxQuery := ent.QueryFromContext(ctx)
		if ctxQuery.Type == generated.TypeOrganization || ctxQuery.Op == OnlyOperation {
			return nil
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Msg("unable to get authenticated user context while traversing organization settings")
			return auth.ErrNoAuthUser
		}

		q.WhereP(organizationbrandingsetting.OrganizationIDIn(orgIDs...))
		return nil
	})
}
