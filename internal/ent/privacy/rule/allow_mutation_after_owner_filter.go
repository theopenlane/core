package rule

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/iam/auth"
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
				return privacy.Deny
			}

			viewerID, err := auth.GetUserIDFromContext(ctx)
			if err != nil {
				return privacy.Skip
			}

			ownerFilter.WhereHasOwnerWith(user.ID(viewerID))

			return privacy.Allowf("applied owner filter")
		},
	)
}
