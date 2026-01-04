package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
)

// AllowMutationAfterApplyingOwnerFilter defines a privacy rule for mutations in the context of an owner filter
func AllowMutationAfterApplyingOwnerFilter() privacy.MutationRule {
	type OwnerFilter interface {
		WhereHasOwnerWith(predicates ...predicate.User)
	}

	return privacy.FilterFunc(
		func(ctx context.Context, f privacy.Filter) error {
			ownerFilter, ok := f.(OwnerFilter)
			if !ok {
				return privacy.Denyf("unable to cast to owner filter")
			}

			viewerID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil {
				return privacy.Skip
			}

			ownerFilter.WhereHasOwnerWith(user.ID(viewerID))

			return privacy.Allowf("applied owner filter")
		},
	)
}
