package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersubprocessor"
	"github.com/theopenlane/iam/auth"
)

// TraverseSubprocessor only returns public subprocessors and subprocessors owned by the organization
func TraverseSubprocessor() ent.Interceptor {
	return intercept.TraverseSubprocessor(func(ctx context.Context, q *generated.SubprocessorQuery) error {
		// allow anonymous access to subprocessors this will only allow view
		// access to the trust center-owned subprocessors
		if tcID, ok := auth.ActiveTrustCenterIDKey.Get(ctx); ok && tcID != "" {
			q.Where(
				subprocessor.Or(
					subprocessor.And(
						subprocessor.OwnerIDIsNil(),
						subprocessor.SystemOwned(true),
					),
					subprocessor.HasTrustCenterSubprocessorsWith(
						trustcentersubprocessor.TrustCenterID(tcID),
					),
				),
			)

			return nil
		}

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			return auth.ErrNoAuthUser
		}

		orgIDs := caller.OrgIDs()

		// filter to return system owned subprocessors and subprocessors owned by the organization
		q.Where(
			subprocessor.Or(
				subprocessor.And(
					subprocessor.OwnerIDIsNil(),
					subprocessor.SystemOwned(true),
				),
				subprocessor.OwnerIDIn(orgIDs...),
			),
		)

		return nil
	})
}
