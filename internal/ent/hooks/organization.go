package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	dbx "github.com/theopenlane/dbx/pkg/dbxclient"
	dbxenums "github.com/theopenlane/dbx/pkg/enums"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/gravatar"
	"github.com/theopenlane/utils/marionette"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/pkg/auth"
	"github.com/theopenlane/core/pkg/enums"
)

// HookOrganization runs on org mutations to set default values that are not provided
func HookOrganization() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrganizationFunc(func(ctx context.Context, mutation *generated.OrganizationMutation) (generated.Value, error) {
			// if this is a soft delete, skip this hook
			if entx.CheckIsSoftDelete(ctx) {
				return next.Mutate(ctx, mutation)
			}

			if mutation.Op().Is(ent.OpCreate) {
				// generate a default org setting schema if not provided
				if err := createOrgSettings(ctx, mutation); err != nil {
					return nil, err
				}

				// check if this is a child org, error if parent org is a personal org
				if err := personalOrgNoChildren(ctx, mutation); err != nil {
					return nil, err
				}
			}

			// set default display name and avatar if not provided
			setDefaultsOnMutations(mutation)

			v, err := next.Mutate(ctx, mutation)
			if err != nil {
				return v, err
			}

			if mutation.Op().Is(ent.OpCreate) {
				orgCreated, ok := v.(*generated.Organization)
				if !ok {
					return nil, err
				}

				// create the admin organization member if not using an API token (which is not associated with a user)
				// otherwise add the API token for admin access to the newly created organization
				if err := createOrgMemberOwner(ctx, orgCreated.ID, mutation); err != nil {
					return v, err
				}

				// create the database, if the org has a dedicated db and dbx is available
				if orgCreated.DedicatedDb && mutation.DBx != nil {
					settings, err := orgCreated.Setting(ctx)
					if err != nil {
						mutation.Logger.Errorw("unable to get organization settings")

						return nil, err
					}

					if err := mutation.Marionette.Queue(marionette.TaskFunc(func(ctx context.Context) error {
						return createDatabase(ctx, orgCreated.ID, settings.GeoLocation.String(), mutation)
					}), marionette.WithErrorf("could not send create the database for %s", orgCreated.Name),
					); err != nil {
						mutation.Logger.Errorw("unable to queue database creation")

						return v, err
					}
				}

				// update the session to drop the user into the new organization
				// if the org is not a personal org, as personal orgs are created during registration
				// and sessions are already set
				if !orgCreated.PersonalOrg {
					as := newAuthSession(mutation.SessionConfig, mutation.TokenManager)

					if err := updateUserAuthSession(ctx, as, orgCreated.ID); err != nil {
						return v, err
					}

					if err := postCreation(ctx, orgCreated, mutation); err != nil {
						return v, err
					}
				}
			}

			return v, err
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookOrganizationDelete runs on org delete mutations to ensure the org can be deleted
func HookOrganizationDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrganizationFunc(func(ctx context.Context, mutation *generated.OrganizationMutation) (generated.Value, error) {
			// by pass checks on invite or pre-allowed request
			// this includes things like the edge-cleanup on user deletion
			if _, allow := privacy.DecisionFromContext(ctx); allow {
				return next.Mutate(ctx, mutation)
			}

			// ensure it's a soft delete or a hard delete, otherwise skip this hook
			if !isDeleteOp(ctx, mutation) {
				return next.Mutate(ctx, mutation)
			}

			// validate the organization can be deleted
			if err := validateDeletion(ctx, mutation); err != nil {
				return nil, err
			}

			v, err := next.Mutate(ctx, mutation)
			if err != nil {
				return v, err
			}

			newOrgID, err := updateUserDefaultOrgOnDelete(ctx, mutation)
			// if we got an error, return it
			// if we didn't get a new org id, keep going and don't
			// update the session cookie
			if err != nil || newOrgID == "" {
				return v, err
			}

			// if the deleted org was the current org, update the session cookie
			as := newAuthSession(mutation.SessionConfig, mutation.TokenManager)

			if err := updateUserAuthSession(ctx, as, newOrgID); err != nil {
				return v, err
			}

			return v, nil
		})
	}, ent.OpDeleteOne|ent.OpDelete|ent.OpUpdate|ent.OpUpdateOne)
}

// setDefaultsOnMutations sets default values on mutations that are not provided
func setDefaultsOnMutations(mutation *generated.OrganizationMutation) {
	if name, ok := mutation.Name(); ok {
		if displayName, ok := mutation.DisplayName(); ok {
			if displayName == "" {
				mutation.SetDisplayName(name)
			}
		}

		url := gravatar.New(name, nil)
		mutation.SetAvatarRemoteURL(url)
	}
}

// createOrgSettings creates the default organization settings for a new org
func createOrgSettings(ctx context.Context, mutation *generated.OrganizationMutation) error {
	// if this is empty generate a default org setting schema
	if _, exists := mutation.SettingID(); !exists {
		// sets up default org settings using schema defaults
		orgSettingID, err := defaultOrganizationSettings(ctx, mutation)
		if err != nil {
			mutation.Logger.Errorw("error creating default organization settings", "error", err)

			return err
		}

		// add the org setting ID to the input
		mutation.SetSettingID(orgSettingID)
	}

	return nil
}

// createEntityTypes creates the default entity types for a new org
func createEntityTypes(ctx context.Context, orgID string, mutation *generated.OrganizationMutation) error {
	if len(mutation.EntConfig.EntityTypes) == 0 {
		return nil
	}

	builders := make([]*generated.EntityTypeCreate, 0, len(mutation.EntConfig.EntityTypes))
	for _, entityType := range mutation.EntConfig.EntityTypes {
		builders = append(builders, mutation.Client().EntityType.Create().
			SetName(entityType).
			SetOwnerID(orgID),
		)
	}

	if err := mutation.Client().EntityType.CreateBulk(builders...).Exec(ctx); err != nil {
		mutation.Logger.Errorw("error creating entity types", "error", err)

		return err
	}

	return nil
}

// postCreation runs after an organization is created to perform additional setup
func postCreation(ctx context.Context, orgCreated *generated.Organization, mutation *generated.OrganizationMutation) error {
	// capture the original org id, ignore error as this will not be set in all cases
	originalOrg, _ := auth.GetOrganizationIDFromContext(ctx) // nolint: errcheck

	// set the new org id in the auth context to process the rest of the post creation steps
	if err := auth.SetOrganizationIDInAuthContext(ctx, orgCreated.ID); err != nil {
		return err
	}

	// create default entity types, if configured
	if err := createEntityTypes(ctx, orgCreated.ID, mutation); err != nil {
		return err
	}

	// reset the original org id in the auth context if it was previously set
	if originalOrg != "" {
		if err := auth.SetOrganizationIDInAuthContext(ctx, originalOrg); err != nil {
			return err
		}
	}

	return nil
}

// validateDeletion ensures the organization can be deleted
func validateDeletion(ctx context.Context, mutation *generated.OrganizationMutation) error {
	deletedID, ok := mutation.ID()
	if !ok {
		return nil
	}

	// do not allow deletion of personal orgs, these are deleted when the user is deleted
	deletedOrg, err := mutation.Client().Organization.Get(ctx, deletedID)
	if err != nil {
		return err
	}

	if deletedOrg.PersonalOrg {
		mutation.Logger.Debugw("attempt to delete personal org detected")

		return fmt.Errorf("%w: %s", ErrInvalidInput, "cannot delete personal organizations")
	}

	return nil
}

// updateUserDefaultOrgOnDelete updates the user's default org if the org being deleted is the user's default org
func updateUserDefaultOrgOnDelete(ctx context.Context, mutation *generated.OrganizationMutation) (string, error) {
	currentUserID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return "", err
	}

	// check if this organization is the user's default org
	deletedOrgID, ok := mutation.ID()
	if !ok {
		return "", nil
	}

	return checkAndUpdateDefaultOrg(ctx, currentUserID, deletedOrgID, mutation.Client())
}

// checkAndUpdateDefaultOrg checks if the old organization is the user's default org and updates it if needed
// this is used when an organization is deleted, as well as when a user is removed from an organization
func checkAndUpdateDefaultOrg(ctx context.Context, userID string, oldOrgID string, client *generated.Client) (string, error) {
	// check if this is the user's default org
	userSetting, err := client.
		UserSetting.
		Query().
		Where(
			usersetting.UserIDEQ(userID),
		).
		WithDefaultOrg().
		Only(ctx)
	if err != nil {
		return "", err
	}

	// if the user's default org was deleted this will now be nil
	if userSetting.Edges.DefaultOrg == nil || userSetting.Edges.DefaultOrg.ID == oldOrgID {
		// set the user's default org another org
		// get the first org that was not the org being deleted
		newDefaultOrgID, err := client.
			Organization.
			Query().
			FirstID(ctx)
		if err != nil {
			return "", err
		}

		if _, err = client.UserSetting.
			UpdateOneID(userSetting.ID).
			SetDefaultOrgID(newDefaultOrgID).
			Save(ctx); err != nil {
			return "", err
		}

		return newDefaultOrgID, nil
	}

	return userSetting.Edges.DefaultOrg.ID, nil
}

// createDatabase creates a new database for the organization
func createDatabase(ctx context.Context, orgID, geo string, mutation *generated.OrganizationMutation) error {
	// set default geo if not provided
	if geo == "" {
		geo = enums.Amer.String()
	}

	input := dbx.CreateDatabaseInput{
		OrganizationID: orgID,
		Geo:            &geo,
		Provider:       &dbxenums.Turso,
	}

	mutation.Logger.Infow("creating database", "org", input.OrganizationID, "geo", input.Geo, "provider", input.Provider)

	if _, err := mutation.DBx.CreateDatabase(ctx, input); err != nil {
		mutation.Logger.Errorw("error creating database", "error", err)

		return err
	}

	// create the database
	return nil
}

// defaultOrganizationSettings creates the default organizations settings for a new org
func defaultOrganizationSettings(ctx context.Context, mutation *generated.OrganizationMutation) (string, error) {
	input := generated.CreateOrganizationSettingInput{}

	organizationSetting, err := mutation.Client().OrganizationSetting.Create().SetInput(input).Save(ctx)
	if err != nil {
		return "", err
	}

	return organizationSetting.ID, nil
}

// personalOrgNoChildren checks if the mutation is for a child org, and if so returns an error
// if the parent org is a personal org
func personalOrgNoChildren(ctx context.Context, mutation *generated.OrganizationMutation) error {
	// check if this is a child org, error if parent org is a personal org
	parentOrgID, ok := mutation.ParentID()
	if ok {
		// check if parent org is a personal org
		parentOrg, err := mutation.Client().Organization.Get(ctx, parentOrgID)
		if err != nil {
			return err
		}

		if parentOrg.PersonalOrg {
			return ErrPersonalOrgsNoChildren
		}
	}

	return nil
}

// createParentOrgTuple creates a parent org tuple if the newly created org has a parent
func createParentOrgTuple(ctx context.Context, m *generated.OrganizationMutation, parentOrgID, childOrgID string) error {
	req := fgax.TupleRequest{
		SubjectID:   parentOrgID,
		SubjectType: "organization",
		ObjectID:    childOrgID,
		ObjectType:  "organization",
		Relation:    fgax.ParentRelation,
	}

	tuple := fgax.GetTupleKey(req)

	if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{
		tuple,
	}, nil); err != nil {
		m.Logger.Errorw("failed to create relationship tuple", "error", err)

		return err
	}

	return nil
}

// createOrgMemberOwner creates the owner of the organization
func createOrgMemberOwner(ctx context.Context, oID string, m *generated.OrganizationMutation) error {
	// This is handled by the user create hook for personal orgs
	personalOrg, _ := m.PersonalOrg()
	if personalOrg {
		return nil
	}

	// If this is a child org, create a parent org tuple instead of owner
	parentOrgID, ok := m.ParentID()
	if ok && parentOrgID != "" {
		return createParentOrgTuple(ctx, m, parentOrgID, oID)
	}

	// if this was created with an API token, do not create an owner but add the service tuple to fga
	if auth.IsAPITokenAuthentication(ctx) {
		return createServiceTuple(ctx, oID, m)
	}

	// get userID from context
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		m.Logger.Errorw("unable to get user id from echo context, unable to add user to organization")

		return err
	}

	// Add User as owner of organization
	owner := enums.RoleOwner
	input := generated.CreateOrgMembershipInput{
		UserID:         userID,
		OrganizationID: oID,
		Role:           &owner,
	}

	if _, err := m.Client().OrgMembership.Create().SetInput(input).Save(ctx); err != nil {
		m.Logger.Errorw("error creating org membership for owner", "error", err)

		return err
	}

	// set the user's default org to the new org
	return updateDefaultOrgIfPersonal(ctx, userID, oID, m.Client())
}

// createServiceTuple creates a service tuple for the organization and api key so the organization can be accessed
func createServiceTuple(ctx context.Context, oID string, m *generated.OrganizationMutation) error {
	// get userID from context
	subjectID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		m.Logger.Errorw("unable to get user id from echo context, unable to add user to organization")

		return err
	}

	// allow the api token to edit the newly created organization, no other users will have access
	// so this is the minimum required access
	role := fgax.CanEdit

	// get tuple key
	req := fgax.TupleRequest{
		SubjectID:   subjectID,
		SubjectType: "service",
		ObjectID:    oID,
		ObjectType:  "organization",
		Relation:    role,
	}

	tuple := fgax.GetTupleKey(req)

	if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{tuple}, nil); err != nil {
		m.Logger.Errorw("failed to create relationship tuple", "error", err)

		return err
	}

	m.Logger.Debugw("created relationship tuples", "relation", role, "object", tuple.Object)

	return nil
}

// updateDefaultOrgIfPersonal updates the user's default org if the user has no default org or
// the default org is their personal org
// the client must be passed in, rather than using the client in the context  because
// this function is sometimes called from a REST handler where the client is not available in the context
func updateDefaultOrgIfPersonal(ctx context.Context, userID, orgID string, client *generated.Client) error {
	// check if the user has a default org
	userSetting, err := client.
		UserSetting.
		Query().
		Where(
			usersetting.UserIDEQ(userID),
		).
		WithDefaultOrg().
		Only(ctx)
	if err != nil {
		return err
	}

	// if the user has no default org, or the default org is the personal org, set the new org as the default org
	if userSetting.Edges.DefaultOrg == nil || userSetting.Edges.DefaultOrg.ID == "" ||
		userSetting.Edges.DefaultOrg.PersonalOrg {
		if _, err = client.UserSetting.
			UpdateOneID(userSetting.ID).
			SetDefaultOrgID(orgID).
			Save(ctx); err != nil {
			return err
		}
	}

	return nil
}
