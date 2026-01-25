package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/logx"
)

// HookCampaignTargetLinkUser links campaign targets to existing users by email.
func HookCampaignTargetLinkUser() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.CampaignTargetFunc(func(ctx context.Context, m *generated.CampaignTargetMutation) (generated.Value, error) {
			email, ok := m.Email()
			if !ok || email == "" {
				return next.Mutate(ctx, m)
			}

			if m.UserCleared() {
				return next.Mutate(ctx, m)
			}

			if _, exists := m.UserID(); exists {
				return next.Mutate(ctx, m)
			}

			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
			existingUser, err := m.Client().User.Query().
				Where(user.EmailEqualFold(email)).
				Only(allowCtx)
			if err != nil {
				if generated.IsNotFound(err) {
					return next.Mutate(ctx, m)
				}

				logx.FromContext(ctx).Error().Err(err).Msg("failed to lookup user for campaign target")
				return next.Mutate(ctx, m)
			}

			m.SetUserID(existingUser.ID)

			return next.Mutate(ctx, m)
		})
	}, hook.And(
		hook.HasFields("email"),
		hook.HasOp(ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate),
	))
}
