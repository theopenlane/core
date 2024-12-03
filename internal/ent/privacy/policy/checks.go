package policy

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

// CheckCreateAccess checks if the user has access to create an object in the org
// for create operations
func CheckCreateAccess() privacy.MutationRule {
	return privacy.OnMutationOperation(
		rule.CheckGroupBasedObjectCreationAccess(),
		ent.OpCreate,
	)
}

// CheckOrgReadAccess checks if the requestor has access to read the organization
func CheckOrgReadAccess() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		return rule.CheckOrgAccess(ctx, fgax.CanView)
	})
}

// CheckOrgWriteAccess checks if the requestor has access to edit the organization
func CheckOrgWriteAccess() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, q ent.Mutation) error {
		return rule.CheckOrgAccess(ctx, fgax.CanEdit)
	})
}

// DenyQueryIfNotAuthenticated denies a query if the user is not authenticated
func DenyQueryIfNotAuthenticated() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		if _, err := auth.GetAuthenticatedUserContext(ctx); err != nil {
			log.Err(err).Msg("unable to get authenticated user context")

			return err
		}

		return nil
	})
}

// DenyMutationIfNotAuthenticated denies a mutation if the user is not authenticated
func DenyMutationIfNotAuthenticated() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		if _, err := auth.GetAuthenticatedUserContext(ctx); err != nil {
			log.Err(err).Msg("unable to get authenticated user context")

			return err
		}

		return nil
	})
}
