package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
)

// AllowMutationAfterApplyingUserOwnerFilter defines a privacy rule for mutations in the context of an owner filter
func AllowMutationAfterApplyingUserOwnerFilter() privacy.MutationRule {
	type OwnerFilter interface {
		WhereHasOwnerWith(predicates ...predicate.User)
	}

	return privacy.FilterFunc(
		func(ctx context.Context, f privacy.Filter) error {
			ownerFilter, ok := f.(OwnerFilter)
			if !ok {
				return privacy.Denyf("unable to cast to owner filter")
			}

			subjectID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil || subjectID == "" {
				return privacy.Skip
			}

			ownerFilter.WhereHasOwnerWith(user.ID(subjectID))

			return privacy.Allowf("applied owner filter")
		},
	)
}
