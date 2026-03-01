package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentercompliance"
)

// TraverseStandard only returns public standards and standards owned by the organization
func TraverseStandard() ent.Interceptor {
	return intercept.TraverseStandard(func(ctx context.Context, q *generated.StandardQuery) error {
		caller, ok := auth.CallerFromContext(ctx)
		if ok && caller != nil && caller.IsAnonymous() {
			q.Where(
				standard.HasTrustCenterCompliancesWith(
					trustcentercompliance.HasTrustCenterWith(
						trustcenter.OwnerID(caller.OrganizationID),
					),
				),
			)

			return nil
		}

		if !ok || caller == nil {
			return auth.ErrNoAuthUser
		}

		orgIDs := caller.OrgIDs()

		systemStandardPredicates := []predicate.Standard{
			standard.OwnerIDIsNil(),
			standard.SystemOwned(true),
		}

		if !auth.IsSystemAdminFromContext(ctx) {
			// if the user is a not-system admin, restrict to only public standards
			systemStandardPredicates = append(systemStandardPredicates, standard.IsPublic(true))
		}

		// filter to return system owned standards and standards owned by the organization
		q.Where(
			standard.Or(
				standard.And(
					systemStandardPredicates...,
				),
				standard.OwnerIDIn(orgIDs...),
			),
		)

		return nil
	})
}
