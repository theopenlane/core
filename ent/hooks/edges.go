package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/ent/privacy/token"
	"github.com/theopenlane/iam/auth"
)

// HookEdgePermissions runs on edge mutations to ensure the user has access to the object they are trying to add for edges.
// It uses the accessmap generated to get the object type and checks if the user has access to it.
func HookEdgePermissions() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if skipperEdgePermissionChecks(ctx) {
				return next.Mutate(ctx, m)
			}
			// ensure the user has access to the object they are trying to add for edges
			if err := checkAccessForEdges(ctx, m); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// skipperEdgePermissionChecks checks if the current context should skip edge permission checks
// this is used for internal requests made by the system or when the user is a system admin
func skipperEdgePermissionChecks(ctx context.Context) bool {
	if _, allow := privacy.DecisionFromContext(ctx); allow || rule.IsInternalRequest(ctx) {
		return true
	}

	if auth.IsSystemAdminFromContext(ctx) {
		return true
	}

	skipTokenType := []token.PrivacyToken{
		&token.OauthTooToken{},
		&token.VerifyToken{},
		&token.SignUpToken{},
		&token.OrgInviteToken{},
		&token.ResetToken{},
		&token.JobRunnerRegistrationToken{},
	}

	if skip := rule.SkipTokenInContext(ctx, skipTokenType); skip {
		return true
	}

	return false
}
