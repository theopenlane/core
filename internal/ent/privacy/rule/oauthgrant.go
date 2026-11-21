package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
)

// AllowIfOwnOauthGrant lets a user revoke or delete their own personal connections, everything else falls through
func AllowIfOwnOauthGrant() privacy.MutationRule {
	return privacy.OauthGrantMutationRuleFunc(func(ctx context.Context, m *generated.OauthGrantMutation) error {
		if !m.Op().Is(ent.OpUpdateOne|ent.OpDeleteOne) && !entx.CheckIsSoftDeleteType(ctx, generated.TypeOauthGrant) {
			return privacy.Skip
		}

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller.SubjectID == "" {
			return privacy.Skip
		}

		id, ok := m.ID()
		if !ok {
			return privacy.Skip
		}

		grant, err := m.Client().OauthGrant.Get(auth.WithInternalReadContext(ctx), id)
		if err != nil {
			return privacy.Skip
		}

		if grant.Kind == enums.OAuthGrantKindUser && grant.UserID == caller.SubjectID {
			return privacy.Allow
		}

		return privacy.Skip
	})
}
