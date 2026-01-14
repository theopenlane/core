package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/gravatar"
	"github.com/theopenlane/utils/passwd"
	"github.com/theopenlane/utils/ulids"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
)

const (
	personalOrgPrefix = "Personal Organization"
)

var (
	// caser is used to capitalize the first letter of words
	caser = cases.Title(language.AmericanEnglish)
)

// HookUser runs on user mutations validate and hash the password and set default values that are not provided
func HookUser() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.UserFunc(func(ctx context.Context, m *generated.UserMutation) (generated.Value, error) {
			if password, ok := m.Password(); ok {
				// validate password before its encrypted
				if passwd.Strength(password) < passwd.Moderate {
					return nil, auth.ErrPasswordTooWeak
				}

				hash, err := passwd.CreateDerivedKey(password)
				if err != nil {
					return nil, err
				}

				m.SetPassword(hash)
			}

			if email, ok := m.Email(); ok {
				// use the email without the domain as the display name, if not provided on creation
				if m.Op().Is(ent.OpCreate) {
					displayName, _ := m.DisplayName()
					if displayName == "" {
						// first try first and last name
						firstName, _ := m.FirstName()
						lastName, _ := m.LastName()

						if firstName != "" && lastName != "" {
							displayName = strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))
						} else {
							// if first and last name are not provided, use the email without the domain
							displayName = strings.Split(email, "@")[0]
						}

						m.SetDisplayName(displayName)
					}

					// set a default avatar if one is not provided
					if _, ok := m.AvatarRemoteURL(); !ok {
						url := gravatar.New(displayName, nil)
						m.SetAvatarRemoteURL(url)
					}

					// lowercase the email for uniqueness
					m.SetEmail(strings.ToLower(email))
				}
			}

			// check for uploaded files (e.g. avatar image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkAvatarFile(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			// user settings are required, if this is empty generate a default setting schema
			if m.Op().Is(ent.OpCreate) {
				settingID, _ := m.SettingID()
				if settingID == "" {
					// sets up default user settings using schema defaults
					userSettingID, err := defaultUserSettings(ctx, m)
					if err != nil {
						return nil, err
					}

					// add the user setting ID to the input
					m.SetSettingID(userSettingID)
				}
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			userCreated, ok := v.(*generated.User)
			if !ok {
				return nil, err
			}

			// handle display name updates for managed groups
			if m.Op().Is(ent.OpUpdateOne) {
				if err := updateSystemManagedGroupForUser(ctx, m, userCreated); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("error updating system managed group name for the user")
					return nil, err
				}
			}

			if m.Op().Is(ent.OpCreate) {
				userCreated.Sub = userCreated.ID

				// set the subject to the user id
				if err := m.Client().User.
					UpdateOneID(userCreated.ID).
					SetSub(userCreated.Sub).
					Exec(ctx); err != nil {
					return nil, err
				}

				// when a user is created, we create a personal user org
				setting, org, err := createPersonalOrg(ctx, m.Client(), userCreated)
				if err != nil {
					return nil, err
				}

				userCreated.Edges.Setting = setting

				// update the personal org setting with the user's email
				if err := updatePersonalOrgSetting(ctx, m.Client(), userCreated, org); err != nil {
					return nil, err
				}
			}

			return userCreated, err
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookUserPermissions runs on user creations to add user _self permissions
// these are used for parent inherited relations on other objects in the system
func HookUserPermissions() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.UserFunc(func(ctx context.Context, m *generated.UserMutation) (generated.Value, error) {
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// add user _self permissions after creation
			userID, _ := m.ID()

			req := fgax.TupleRequest{
				SubjectID:   userID,
				SubjectType: auth.UserSubjectType,
				ObjectID:    userID,
				ObjectType:  auth.UserSubjectType,
				Relation:    fgax.SelfRelation,
			}

			if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{fgax.GetTupleKey(req)}, nil); err != nil {
				return nil, err
			}

			return v, err
		})
	}, ent.OpCreate)
}

// HookDeleteUser runs on user deletions to clean up personal organizations
func HookDeleteUser() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.UserFunc(func(ctx context.Context, m *generated.UserMutation) (generated.Value, error) {
			if isDeleteOp(ctx, m) {
				userID, _ := m.ID()
				// get the personal org id
				user, err := m.Client().User.Get(ctx, userID)
				if err != nil {
					return nil, err
				}

				// the user might not be currently in the authorized context to see the personal org
				// so we need to allow the context to see the personal org
				allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

				personalOrgIDs, err := m.Client().User.QueryOrganizations(user).Where(organization.PersonalOrg(true)).IDs(allowCtx)
				if err != nil {
					return nil, err
				}

				// run the mutation first
				v, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				// cleanup personal org(s) using the allow context set above
				if _, err := m.Client().
					Organization.
					Delete().
					Where(organization.IDIn(personalOrgIDs...)).
					Exec(allowCtx); err != nil {
					return nil, err
				}

				return v, err
			}

			return next.Mutate(ctx, m)
		})
	}
}

// getPersonalOrgInput generates the input for a new personal organization
// personal orgs are assigned to all new users when registering
func getPersonalOrgInput(user *generated.User) generated.CreateOrganizationInput {
	// generate random name for personal orgs
	randomName := caser.String(gofakeit.PetName())
	name := fmt.Sprintf("%s-%s", randomName, ulids.New().String())
	displayName := randomName
	personalOrg := true

	return generated.CreateOrganizationInput{
		Name:        name,
		DisplayName: &displayName,
		Description: personalOrgDescription(user),
		PersonalOrg: &personalOrg,
	}
}

// personalOrgDescription generates a description for a personal org based on the
// user's first and last name
func personalOrgDescription(user *generated.User) *string {
	desc := personalOrgPrefix

	if user.FirstName == "" && user.LastName == "" {
		return &desc
	}

	desc = fmt.Sprintf("%s - %s %s", desc, caser.String(user.FirstName), caser.String(user.LastName))

	return &desc
}

// createPersonalOrg creates an org for a user with a unique random name
func createPersonalOrg(ctx context.Context, dbClient *generated.Client, user *generated.User) (*generated.UserSetting, *generated.Organization, error) {
	// this prevents a privacy check that would be required for regular orgs, but not a personal org
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	orgInput := getPersonalOrgInput(user)

	org, err := dbClient.Organization.Create().SetInput(orgInput).Save(ctx)
	if err != nil {
		// retry on unique constraint
		if generated.IsConstraintError(err) {
			return createPersonalOrg(ctx, dbClient, user)
		}

		logx.FromContext(ctx).Error().Err(err).Msg("unable to create personal org")

		return nil, nil, err
	}

	// Create Role as owner for user in the personal org
	input := generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           &enums.RoleOwner,
	}

	if err := dbClient.OrgMembership.Create().
		SetInput(input).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to add user as owner to organization")

		return nil, nil, err
	}

	setting, err := setDefaultOrg(ctx, dbClient, user, org)
	if err != nil {
		return nil, nil, err
	}

	return setting, org, nil
}

func updatePersonalOrgSetting(ctx context.Context, dbClient *generated.Client, user *generated.User, org *generated.Organization) error {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	setting, err := org.Setting(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to get org settings")

		return err
	}

	if err := dbClient.OrganizationSetting.UpdateOneID(setting.ID).SetBillingEmail(user.Email).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to update org settings")

		return err
	}

	return nil
}

// setDefaultOrg sets the default org for a user in their settings
func setDefaultOrg(ctx context.Context, dbClient *generated.Client, user *generated.User, org *generated.Organization) (*generated.UserSetting, error) {
	setting, err := user.Setting(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to get user settings")

		return nil, err
	}

	setting, err = dbClient.UserSetting.UpdateOneID(setting.ID).SetDefaultOrg(org).Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to set default org")

		return nil, err
	}

	// set the default org on the settings to eager load for the response
	setting.Edges.DefaultOrg = org

	return setting, nil
}

// defaultUserSettings creates the default user settings for a new user
func defaultUserSettings(ctx context.Context, user *generated.UserMutation) (string, error) {
	input := generated.CreateUserSettingInput{}

	userSetting, err := user.Client().UserSetting.Create().SetInput(input).Save(ctx)
	if err != nil {
		return "", err
	}

	return userSetting.ID, nil
}

func updateSystemManagedGroupForUser(ctx context.Context, m *generated.UserMutation, user *generated.User) error {
	displayName, ok := m.DisplayName()
	if !ok {
		return nil
	}

	oldDisplayName, err := m.OldDisplayName(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error getting old display name")
		return err
	}

	// if the display name is still the same, nothing to do really
	if oldDisplayName == displayName {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	memberships, err := m.Client().OrgMembership.Query().
		Where(orgmembership.UserID(user.ID)).
		All(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error querying user's org memberships")
		return err
	}

	for _, membership := range memberships {
		newCtx := auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{
			SubjectID:       user.ID,
			OrganizationID:  membership.OrganizationID,
			OrganizationIDs: []string{membership.OrganizationID},
		})

		groups, err := m.Client().Group.Query().
			Where(
				group.OwnerID(membership.OrganizationID),
				group.CreatedBy(user.ID),
				group.IsManaged(true),
				group.Name(getUserGroupName(oldDisplayName, user.ID)),
			).
			All(privacy.DecisionContext(newCtx, privacy.Allow))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error querying user's system managed groups for org")
			return err
		}

		if len(groups) == 0 {
			continue
		}

		groupIDs := make([]string, 0, len(groups))
		for _, g := range groups {
			groupIDs = append(groupIDs, g.ID)
		}

		err = m.Client().Group.Update().
			Where(group.IDIn(groupIDs...)).
			SetName(getUserGroupName(displayName, user.ID)).
			SetDisplayName(displayName).
			SetDescription(getUserGroupName(displayName, user.ID)).
			Exec(privacy.DecisionContext(newCtx, privacy.Allow))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).
				Str("old_display_name", oldDisplayName).Str("new_display_name", displayName).
				Msg("error updating system managed group names in bulk")

			return err
		}
	}

	return nil
}
