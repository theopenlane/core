package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// HookUserSetting runs on user settings mutations and validates input on update
func HookUserSetting() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.UserSettingFunc(func(ctx context.Context, m *generated.UserSettingMutation) (generated.Value, error) {
			org, ok := m.DefaultOrgID()
			if ok && !allowDefaultOrgUpdate(ctx, m, org) {
				return nil, rout.InvalidField(rout.ErrOrganizationNotFound)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// allowDefaultOrgUpdate checks if the user has access to the organization being updated as their default org
func allowDefaultOrgUpdate(ctx context.Context, m *generated.UserSettingMutation, orgID string) bool {
	// allow if explicitly allowed or if it's an internal request
	if _, allow := privacy.DecisionFromContext(ctx); allow || rule.IsInternalRequest(ctx) {
		return true
	}

	// allow for org invite tokens
	if rule.ContextHasPrivacyTokenOfType[*token.OrgInviteToken](ctx) {
		return true
	}

	// ensure user has access to the organization
	// the ID is always set on update
	userSettingID, _ := m.ID()

	owner, err := m.Client().
		User.
		Query().
		Where(
			user.HasSettingWith(usersetting.ID(userSettingID)),
		).
		Only(ctx)
	if err != nil {
		return false
	}

	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("unable to get authenticated user context")

		return false
	}

	req := fgax.AccessCheck{
		SubjectID:   owner.ID,
		SubjectType: auth.UserSubjectType,
		ObjectID:    orgID,
		Context:     utils.NewOrganizationContextKey(au.SubjectEmail),
	}

	allow, err := m.Authz.CheckOrgReadAccess(ctx, req)
	if err != nil {
		return false
	}

	return allow
}
