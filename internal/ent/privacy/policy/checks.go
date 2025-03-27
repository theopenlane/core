package policy

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
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
		// check if the user has access to view the organization
		// check the query first for the IDS
		query, ok := q.(*generated.OrganizationQuery)
		if ok {
			// if we get an error (privacy.Deny or privacy.Allow are both "errors")
			// return as the result
			// if its nil, we didn't check anything and should continue to the next check
			if err := rule.CheckOrgAccessBasedOnRequest(ctx, fgax.CanView, query); err != nil {
				return err
			}
		}

		// otherwise check against the current context
		return rule.CheckCurrentOrgAccess(ctx, nil, fgax.CanView)
	})
}

// CheckOrgWriteAccess checks if the requestor has access to edit the organization
func CheckOrgWriteAccess() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		return rule.CheckCurrentOrgAccess(ctx, m, fgax.CanEdit)
	})
}

// DenyQueryIfNotAuthenticated denies a query if the user is not authenticated
func DenyQueryIfNotAuthenticated() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, _ ent.Query) error {
		if res, err := auth.GetAuthenticatedUserFromContext(ctx); err != nil || res == nil {
			log.Err(err).Msg("unable to get authenticated user context")

			return err
		}

		return nil
	})
}

// DenyMutationIfNotAuthenticated denies a mutation if the user is not authenticated
func DenyMutationIfNotAuthenticated() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, _ ent.Mutation) error {
		if res, err := auth.GetAuthenticatedUserFromContext(ctx); err != nil || res == nil {
			log.Err(err).Msg("unable to get authenticated user context")

			return err
		}

		return nil
	})
}
