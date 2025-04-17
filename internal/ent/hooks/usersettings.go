package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/generated/webauthn"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
)

// HookUserSetting runs on user settings mutations and validates input on update
func HookUserSetting() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.UserSettingFunc(func(ctx context.Context, m *generated.UserSettingMutation) (generated.Value, error) {
			org, ok := m.DefaultOrgID()
			if ok && !allowDefaultOrgUpdate(ctx, m, org) {
				return nil, rout.InvalidField(rout.ErrOrganizationNotFound)
			}

			if m.Op().Is(ent.OpUpdateOne) {
				userID, err := auth.GetSubjectIDFromContext(ctx)
				if err == nil && userID != "" {
					// if webauthn is disabled, clean up the passkey we stored previously
					if allowed, _ := m.IsWebauthnAllowed(); !allowed {
						_, err := m.Client().Webauthn.Delete().
							Where(webauthn.OwnerID(userID)).
							Exec(ctx)
						if err != nil {
							return nil, err
						}
					}
				}
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

	req := fgax.AccessCheck{
		SubjectID:   owner.ID,
		SubjectType: auth.UserSubjectType,
		ObjectID:    orgID,
	}

	allow, err := m.Authz.CheckOrgReadAccess(ctx, req)
	if err != nil {
		return false
	}

	return allow
}
