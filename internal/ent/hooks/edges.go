package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/iam/auth"
)

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
