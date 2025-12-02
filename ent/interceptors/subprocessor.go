package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/intercept"
	"github.com/theopenlane/ent/generated/subprocessor"
	"github.com/theopenlane/iam/auth"
)

// TraverseSubprocessor only returns public subprocessors and subprocessors owned by the organization
func TraverseSubprocessor() ent.Interceptor {
	return intercept.TraverseSubprocessor(func(ctx context.Context, q *generated.SubprocessorQuery) error {
		var (
			orgIDs []string
			err    error
		)

		// allow anonymous access to subprocessors this will only allow view
		// access to the trust center-owned subprocessors we could probably go
		// further here and check that the subprocessor is referenced by a
		// trustcentersubprocessor, but i don't think that is necessary here

		if anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx); ok {
			orgIDs = []string{anon.OrganizationID}
		} else {
			orgIDs, err = auth.GetOrganizationIDsFromContext(ctx)
			if err != nil {
				return err
			}
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
