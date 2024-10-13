package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/entx"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/utils/passwd"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
)

const (
	personalOrgPrefix = "Personal Organization"
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
						displayName := strings.Split(email, "@")[0]

						m.SetDisplayName(displayName)
					}
				}
			}

			// check for uploaded files (e.g. avatar image)
			fileIDs := objects.GetFileIDsFromContext(ctx)

			if len(fileIDs) > 0 {
				m.AddFileIDs(fileIDs...)

				if err := checkAvatarFile(ctx, m); err != nil {
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

			if m.Op().Is(ent.OpCreate) {
				userCreated.Sub = userCreated.ID

				// set the subject to the user id
				if _, err := m.Client().User.
					UpdateOneID(userCreated.ID).
					SetSub(userCreated.Sub).
					Save(ctx); err != nil {
					return nil, err
				}

				// when a user is created, we create a personal user org
				setting, err := createPersonalOrg(ctx, m.Client(), userCreated)
				if err != nil {
					return nil, err
				}

				userCreated.Edges.Setting = setting
			}

			return userCreated, err
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookDeleteUser runs on user deletions to clean up personal organizations
func HookDeleteUser() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.UserFunc(func(ctx context.Context, m *generated.UserMutation) (generated.Value, error) {
			if m.Op().Is(ent.OpDelete|ent.OpDeleteOne) || entx.CheckIsSoftDelete(ctx) {
				userID, _ := m.ID()
				// get the personal org id
				user, err := m.Client().User.Get(ctx, userID)
				if err != nil {
					return nil, err
				}

				personalOrgIDs, err := m.Client().User.QueryOrganizations(user).Where(organization.PersonalOrg(true)).IDs(ctx)
				if err != nil {
					return nil, err
				}

				// run the mutation first
				v, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				// cleanup personal org(s)
				allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
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
	// caser is used to capitalize the first letter of words
	caser := cases.Title(language.AmericanEnglish)

	// generate random name for personal orgs
	name := caser.String(petname.Generate(2, " ")) //nolint:mnd
	displayName := name
	personalOrg := true
	desc := fmt.Sprintf("%s - %s %s", personalOrgPrefix, caser.String(user.FirstName), caser.String(user.LastName))

	return generated.CreateOrganizationInput{
		Name:        name,
		DisplayName: &displayName,
		Description: &desc,
		PersonalOrg: &personalOrg,
	}
}

// createPersonalOrg creates an org for a user with a unique random name
func createPersonalOrg(ctx context.Context, dbClient *generated.Client, user *generated.User) (*generated.UserSetting, error) {
	// this prevents a privacy check that would be required for regular orgs, but not a personal org
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	orgInput := getPersonalOrgInput(user)

	org, err := dbClient.Organization.Create().SetInput(orgInput).Save(ctx)
	if err != nil {
		// retry on unique constraint
		if generated.IsConstraintError(err) {
			return createPersonalOrg(ctx, dbClient, user)
		}

		log.Error().Err(err).Msg("unable to create personal org")

		return nil, err
	}

	// Create Role as owner for user in the personal org
	input := generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           &enums.RoleOwner,
	}

	if _, err := dbClient.OrgMembership.Create().
		SetInput(input).
		Save(ctx); err != nil {
		log.Error().Err(err).Msg("unable to add user as owner to organization")

		return nil, err
	}

	// set default org
	return setDefaultOrg(ctx, dbClient, user, org)
}

func setDefaultOrg(ctx context.Context, dbClient *generated.Client, user *generated.User, org *generated.Organization) (*generated.UserSetting, error) {
	setting, err := user.Setting(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to get user settings")

		return nil, err
	}

	setting, err = dbClient.UserSetting.UpdateOneID(setting.ID).SetDefaultOrg(org).Save(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to set default org")

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

// checkAvatarFile checks if an avatar file is provided and sets the local file ID
func checkAvatarFile(ctx context.Context, m *generated.UserMutation) error {
	file, _ := objects.FilesFromContextWithKey(ctx, "avatarFile")

	if file == nil {
		return nil
	}

	if len(file) > 1 {
		return ErrTooManyAvatarFiles
	}

	m.SetAvatarLocalFileID(file[0].ID)

	return nil
}
