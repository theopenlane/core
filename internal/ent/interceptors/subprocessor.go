package interceptors

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/intercept"
	"github.com/theopenlane/core/v2/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcentersubprocessor"
	"github.com/theopenlane/iam/auth"
)

// AnonScopeSubprocessor limits anonymous trust center callers to subprocessors linked to the active trust center
func AnonScopeSubprocessor(trustCenterID string) func(*sql.Selector) {
	return subprocessor.HasTrustCenterSubprocessorsWith(
		trustcentersubprocessor.TrustCenterID(trustCenterID),
	)
}

// TraverseSubprocessor only returns public subprocessors and subprocessors owned by the organization
func TraverseSubprocessor() ent.Interceptor {
	return intercept.TraverseSubprocessor(func(ctx context.Context, q *generated.SubprocessorQuery) error {
		// anonymous trust center users are scoped by AnonScopeSubprocessor in the object owned mixin
		if auth.IsTrustCenterFromContext(ctx) {
			return nil
		}

		// trusted internal callers that bypass org scoping (e.g. scheduled pollers) read every
		// organization's subprocessors
		if auth.HasCrossOrgCapabilities(ctx) {
			return nil
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return auth.ErrNoAuthUser
		}

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
