package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"

	"github.com/rs/zerolog"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/enums"
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

// HookUserSettingEmailConfirmation runs on user settings mutations and handles auto-join when email is confirmed
func HookUserSettingEmailConfirmation() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.UserSettingFunc(func(ctx context.Context, m *generated.UserSettingMutation) (generated.Value, error) {
			// check if EmailConfirmed is being set to true
			emailConfirmed, ok := m.EmailConfirmed()
			if !ok || !emailConfirmed {
				// not setting email confirmed to true, continue with normal flow
				return next.Mutate(ctx, m)
			}

			// execute the mutation first, so the user setting is updated
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// get the user associated with this user setting
			userSettingID, _ := m.ID()
			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

			user, err := m.Client().User.Query().
				Where(user.HasSettingWith(usersetting.ID(userSettingID))).
				Only(allowCtx)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("unable to get user for auto-join")

				return nil, err
			}

			// perform auto-join logic
			if err := autoJoinOrganizationsForUser(allowCtx, m.Client(), user); err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("auto-join failed")

				return nil, err
			}

			return v, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// autoJoinOrganizationsForUser automatically adds a user to organizations with matching email domains
func autoJoinOrganizationsForUser(ctx context.Context, dbClient *generated.Client, user *generated.User) error {
	userDomain := strings.Split(user.Email, "@")[1]

	// find all organizations with allow_matching_domains_autojoin enabled
	orgs, err := dbClient.Organization.Query().
		Where(organization.And(
			organization.PersonalOrg(false),
			organization.HasSettingWith(
				organizationsetting.And(
					organizationsetting.AllowMatchingDomainsAutojoin(true),
					func(s *sql.Selector) {
						s.Where(sqljson.ValueContains(organizationsetting.FieldAllowedEmailDomains, userDomain))
					},
				),
			),
		)).
		WithSetting().All(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("unable to query organizations for auto-join")

		return err
	}

	for _, org := range orgs {
		// check if user is already a member
		exists, err := dbClient.OrgMembership.Query().
			Where(
				orgmembership.UserID(user.ID),
				orgmembership.OrganizationID(org.ID),
			).
			Exist(ctx)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error checking organization membership")

			continue
		}

		if exists {

			continue
		}

		// add user as a member to the organization
		defaultRole := enums.RoleMember
		input := generated.CreateOrgMembershipInput{
			OrganizationID: org.ID,
			UserID:         user.ID,
			Role:           &defaultRole,
		}

		if err := dbClient.OrgMembership.Create().
			SetInput(input).
			Exec(ctx); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("unable to auto-join user to organization")

			continue
		}

		zerolog.Ctx(ctx).Debug().Str("user_id", user.ID).Str("org_id", org.ID).Str("domain", userDomain).Msg("user auto-joined organization based on email domain match")
	}

	return nil
}
