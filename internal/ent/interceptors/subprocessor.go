package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/iam/auth"
)

// TraverseSubprocessor only returns public subprocessors and subprocessors owned by the organization
func TraverseSubprocessor() ent.Interceptor {
	return intercept.TraverseSubprocessor(func(ctx context.Context, q *generated.SubprocessorQuery) error {
		zerolog.Ctx(ctx).Debug().Msg("traversing subprocessor")
		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return err
		}

		systemSubprocessorPredicates := []predicate.Subprocessor{
			subprocessor.OwnerIDIsNil(),
			subprocessor.SystemOwned(true),
		}

		// filter to return system owned subprocessors and subprocessors owned by the organization
		q.Where(
			subprocessor.Or(
				subprocessor.And(
					systemSubprocessorPredicates...,
				),
				subprocessor.OwnerIDIn(orgIDs...),
			),
		)

		return nil
	})
}
