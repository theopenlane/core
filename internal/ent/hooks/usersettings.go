package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"

	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/invite"
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
	"github.com/theopenlane/core/pkg/logx"
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
		logx.FromContext(ctx).Error().Err(err).Msg("unable to get authenticated user context")

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
// and sends welcome email after verification
func HookUserSettingEmailConfirmation() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.UserSettingFunc(func(ctx context.Context, m *generated.UserSettingMutation) (generated.Value, error) {
			// check if EmailConfirmed is being set to true
			emailConfirmed, ok := m.EmailConfirmed()
			if !ok || !emailConfirmed {
				// not setting email confirmed to true, continue with normal flow
				return next.Mutate(ctx, m)
			}

			oldEmailConfirmed, _ := m.OldEmailConfirmed(ctx)
			if oldEmailConfirmed {
				// email was already confirmed, continue with normal flow
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
				logx.FromContext(ctx).Error().Err(err).Msg("unable to get user for auto-join")

				return nil, err
			}

			// perform auto-join logic
			if err := autoJoinOrganizationsForUser(allowCtx, m.Client(), user); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("auto-join failed")

				return nil, err
			}

			// send a welcome email to the user
			if err := sendRegisterWelcomeEmail(ctx, user, m); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("could not send welcome email")
			}

			return v, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// sendRegisterWelcomeEmail sends a welcome email to the user after registration welcoming to the platform
func sendRegisterWelcomeEmail(ctx context.Context, user *generated.User, m *generated.UserSettingMutation) error {
	// if there is not job client, we can't send the email
	if m.Job == nil {
		logx.FromContext(ctx).Info().Msg("no job client, skipping welcome email")

		return nil
	}

	email, err := m.Emailer.NewWelcomeEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating welcome email")

		return err
	}

	if _, err = m.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error queueing email verification")

		return err
	}

	return nil
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
		logx.FromContext(ctx).Error().Err(err).Msg("unable to query organizations for auto-join")

		return err
	}

	lastOrgID := ""
	for _, org := range orgs {
		// check if user is already a member
		exists, err := dbClient.OrgMembership.Query().
			Where(
				orgmembership.UserID(user.ID),
				orgmembership.OrganizationID(org.ID),
			).
			Exist(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error checking organization membership")

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

		// add organization to the context for auto-join
		ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{
			SubjectID:       user.ID,
			SubjectEmail:    user.Email,
			OrganizationID:  org.ID,
			OrganizationIDs: []string{org.ID},
		})

		if err := dbClient.OrgMembership.Create().
			SetInput(input).
			Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("unable to auto-join user to organization")

			// swallowing this error means some transactions are not rolled back and you can end up in a partial membership state
			return err
		}

		if err := markPendingInvitesAsAccepted(ctx, dbClient, user.Email, org.ID); err != nil {
			return err
		}

		logx.FromContext(ctx).Debug().Str("user_id", user.ID).Str("org_id", org.ID).Str("domain", userDomain).Msg("user auto-joined organization based on email domain match")

		lastOrgID = org.ID
	}

	if lastOrgID != "" {
		// set the user's default organization to the last one they were added to
		if _, err := dbClient.UserSetting.Update().
			Where(usersetting.UserID(user.ID), usersetting.DeletedAtIsNil()).
			SetDefaultOrgID(lastOrgID).
			Save(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("unable to update user's default organization after auto-join")

			return err
		}
	}

	return nil
}

func markPendingInvitesAsAccepted(ctx context.Context, dbClient *generated.Client, email, orgID string) error {
	_, err := dbClient.Invite.Update().
		Where(
			invite.Recipient(email),
			invite.OwnerID(orgID),
			invite.StatusIn(enums.InvitationSent, enums.ApprovalRequired),
			invite.DeletedAtIsNil(),
		).
		SetStatus(enums.InvitationAccepted).
		Save(ctx)

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to mark pending invites as accepted")
		return err
	}

	return nil
}
