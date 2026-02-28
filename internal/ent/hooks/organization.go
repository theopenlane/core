package hooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v84"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/gravatar"

	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
)

// HookOrganization runs on org mutations to set default values that are not provided
func HookOrganization() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrganizationFunc(func(ctx context.Context, m *generated.OrganizationMutation) (generated.Value, error) {
			// if this is a soft delete, skip this hook
			if isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			// existingCaller is captured here so it can be mutated after org creation
			// to propagate the new org ID back to the original caller pointer
			var existingCaller *auth.Caller

			if m.Op().Is(ent.OpCreate) {
				// add bypass capabilities to the caller for the duration of org creation
				// so that downstream hooks skip owner-field and managed-group guards
				const orgCreationCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup
				if caller, hasCaller := auth.CallerFromContext(ctx); hasCaller {
					existingCaller = caller
					ctx = auth.WithCaller(ctx, caller.WithCapabilities(orgCreationCaps))
				} else {
					ctx = auth.WithCaller(ctx, &auth.Caller{Capabilities: orgCreationCaps})
				}

				// generate a default org setting schema if not provided
				if err := createOrgSettings(ctx, m); err != nil {
					return nil, err
				}

				// check if this is a child org, error if parent org is a personal org
				if err := personalOrgNoChildren(ctx, m); err != nil {
					return nil, err
				}

				// trim trailing whitespace from the name
				if name, ok := m.Name(); ok {
					m.SetName(strings.TrimSpace(name))
				}
			}

			// set default display name and avatar if not provided
			setDefaultsOnMutations(m)

			// check for uploaded files (e.g. avatar image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkAvatarFile(ctx, m)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			if m.Op().Is(ent.OpCreate) {
				orgCreated, ok := v.(*generated.Organization)
				if !ok {
					return nil, err
				}

				// create the admin organization member if not using an API token (which is not associated with a user)
				// otherwise add the API token for admin access to the newly created organization
				if err := createOrgMemberOwner(ctx, orgCreated.ID, m); err != nil {
					return v, err
				}

				// create the database, if the org has a dedicated db and dbx is available
				if orgCreated.DedicatedDb {
					// on create the org will not yet have access to the settings
					// allow the request to proceed to get the org settings
					allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

					settings, err := orgCreated.Setting(allowCtx)
					if err != nil {
						logx.FromContext(ctx).Error().Err(err).Msg("unable to get organization settings")

						return nil, err
					}

					if _, err := m.Job.Insert(ctx, jobs.DatabaseArgs{
						OrganizationID: orgCreated.ID,
						Location:       settings.GeoLocation.String(),
					}, nil); err != nil {
						return nil, err
					}
				}

				// update the session to drop the user into the new organization
				// if the org is not a personal org, as personal orgs are created during registration
				// and sessions are already set
				if !orgCreated.PersonalOrg {
					am := authmanager.New(m.Client())
					if err := updateUserAuthSession(ctx, am, orgCreated.ID); err != nil {
						return v, err
					}

					// propagate the new org ID back through the original caller pointer
					// so that callers holding the same *Caller see the updated org
					if existingCaller != nil {
						existingCaller.OrganizationID = orgCreated.ID
					}

					if err := postOrganizationCreation(ctx, orgCreated, m); err != nil {
						return v, err
					}
				}
			}

			return v, err
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookOrganizationDelete runs on org delete mutations to ensure the org can be deleted
func HookOrganizationDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrganizationFunc(func(ctx context.Context, m *generated.OrganizationMutation) (generated.Value, error) {
			// by pass checks on invite or pre-allowed request
			// this includes things like the edge-cleanup on user deletion
			if _, allow := privacy.DecisionFromContext(ctx); allow {
				return next.Mutate(ctx, m)
			}

			// ensure it's a soft delete or a hard delete, otherwise skip this hook
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			// validate the organization can be deleted
			if err := validateOrgDeletion(ctx, m); err != nil {
				return nil, err
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			if err := updateOrgSubscriptionOnDelete(ctx, m); err != nil {
				// do not block the delete if we can't update the subscription
				// the subscription will be cancelled with the event hook after this
				// mutation completes
				logx.FromContext(ctx).Error().Err(err).Msg("failed to update org subscription on organization delete")
			}

			newOrgID, err := updateUserDefaultOrgOnDelete(ctx, m)
			// if we got an error, log it and return
			// if we didn't get a new org id, keep going and don't
			// update the session cookie
			// returning an error here would rollback the delete but permissions would already be removed
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to update user's default organization on organization delete")
				return v, nil
			}

			if newOrgID == "" {
				return v, nil
			}

			// if the deleted org was the current org, update the session cookie
			am := authmanager.New(m.Client())

			if err := updateUserAuthSession(ctx, am, newOrgID); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to update user auth session on organization delete")

				return v, nil
			}

			return v, nil
		})
	}, ent.OpDeleteOne|ent.OpDelete|ent.OpUpdate|ent.OpUpdateOne)
}

// setDefaultsOnMutations sets default values on mutations that are not provided
func setDefaultsOnMutations(m *generated.OrganizationMutation) {
	if name, ok := m.Name(); ok {
		if displayName, ok := m.DisplayName(); ok {
			if displayName == "" {
				m.SetDisplayName(name)
			}
		}

		url := gravatar.New(name, nil)
		m.SetAvatarRemoteURL(url)
	}
}

// createOrgSettings creates the default organization settings for a new org
func createOrgSettings(ctx context.Context, m *generated.OrganizationMutation) error {
	// if this is empty generate a default org setting schema
	if _, exists := m.SettingID(); !exists {
		// sets up default org settings using schema defaults
		orgSettingID, err := defaultOrganizationSettings(ctx, m)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error creating default organization settings")

			return err
		}

		// add the org setting ID to the input
		m.SetSettingID(orgSettingID)
	}

	return nil
}

// createOrgSubscription creates the default organization subscription for a new org
func createOrgSubscription(ctx context.Context, orgCreated *generated.Organization, m utils.GenericMutation) (*generated.OrgSubscription, error) {
	// ensure we can always pull the org subscription for the organization
	allowCtx := auth.WithCaller(ctx, auth.NewWebhookCaller(orgCreated.ID))

	orgSubscriptions, err := orgCreated.OrgSubscriptions(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error getting org subscriptions")
		return nil, err
	}

	// if this is empty generate a default org setting schema
	if len(orgSubscriptions) == 0 {
		sub, err := defaultOrgSubscription(ctx, orgCreated, m)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error creating default org subscription")

			return nil, err
		}

		orgSubscriptions = []*generated.OrgSubscription{sub}
	}

	logx.FromContext(ctx).Debug().Msg("created default org subscription")

	return orgSubscriptions[0], nil
}

const (
	// subscriptionPendingUpdate is the status for a pending subscription update
	// when the object is initially created in our database
	subscriptionPendingUpdate = "PENDING_UPDATE"
)

// defaultOrgSubscription is the default way to create an org subscription when an organization is first created
func defaultOrgSubscription(ctx context.Context, orgCreated *generated.Organization, m utils.GenericMutation) (*generated.OrgSubscription, error) {
	subs, err := m.Client().OrgSubscription.Create().
		SetStripeSubscriptionID(subscriptionPendingUpdate).
		SetOwnerID(orgCreated.ID).
		SetActive(true).
		SetStripeSubscriptionStatus(string(stripe.SubscriptionStatusTrialing)).Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating default orgsubscription")

		return nil, err
	}

	return subs, nil
}

// createEntityTypes creates the default entity types for a new org
func createEntityTypes(ctx context.Context, orgID string, m *generated.OrganizationMutation) error {
	if m.EntConfig == nil || len(m.EntConfig.EntityTypes) == 0 {
		return nil
	}

	builders := make([]*generated.EntityTypeCreate, 0, len(m.EntConfig.EntityTypes))
	for _, entityType := range m.EntConfig.EntityTypes {
		builders = append(builders, m.Client().EntityType.Create().
			SetName(entityType).
			SetOwnerID(orgID),
		)
	}

	if err := m.Client().EntityType.CreateBulk(builders...).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating entity types")

		return err
	}

	return nil
}

// postOrganizationCreation runs after an organization is created to perform additional setup
func postOrganizationCreation(ctx context.Context, orgCreated *generated.Organization, m *generated.OrganizationMutation) error {
	// capture the original org id, ignore error as this will not be set in all cases
	originalOrg, _ := auth.GetOrganizationIDFromContext(ctx) //nolint:errcheck

	// set the new org id in the auth context to process the rest of the post creation steps
	ctx, err := auth.SetOrganizationIDInAuthContext(ctx, orgCreated.ID)
	if err != nil {
		return err
	}

	// create default entity types, if configured
	if err := createEntityTypes(ctx, orgCreated.ID, m); err != nil {
		return err
	}

	// create generated groups
	if err := generateOrganizationGroups(ctx, m, orgCreated.ID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating generated groups")

		return err
	}

	// create subscriptions if the entitlement manager is enabled
	if m.EntitlementManager.Config.IsEnabled() {
		orgSubs, err := createOrgSubscription(ctx, orgCreated, m)
		if err != nil {
			return err
		}

		opts := []reconciler.OrgModuleOption{reconciler.WithTrial()}
		if m.Client().EntConfig != nil && m.Client().EntConfig.Modules.DevMode {
			opts = append(opts, reconciler.WithAllModules())
		}

		_, err = reconciler.CreateDefaultOrgModulesProductsPrices(ctx, m.Client(), orgSubs, orgCreated.ID, opts...)
		if err != nil {
			return err
		}
	}

	// reset the original org id in the auth context if it was previously set
	if originalOrg != "" {
		if ctx, err = auth.SetOrganizationIDInAuthContext(ctx, originalOrg); err != nil {
			return err
		}
	}

	return nil
}

// validateOrgDeletion ensures the organization can be deleted
func validateOrgDeletion(ctx context.Context, m *generated.OrganizationMutation) error {
	deletedID, ok := m.ID()
	if !ok || deletedID == "" {
		return nil
	}

	// do not allow deletion of personal orgs, these are deleted when the user is deleted
	exists, _ := m.Client().Organization.Query().
		Where(
			organization.ID(deletedID),
			organization.PersonalOrgEQ(true),
		).
		Exist(ctx)

	if exists {
		logx.FromContext(ctx).Debug().Msg("attempt to delete personal org detected")

		return fmt.Errorf("%w: %s", ErrInvalidInput, "cannot delete personal organizations")
	}

	return nil
}

// updateUserDefaultOrgOnDelete updates the user's default org if the org being deleted is the user's default org
func updateUserDefaultOrgOnDelete(ctx context.Context, m *generated.OrganizationMutation) (string, error) {
	currentUserID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		return "", err
	}

	// check if this organization is the user's default org
	deletedOrgID, ok := m.ID()
	if !ok {
		return "", nil
	}

	return checkAndUpdateDefaultOrg(ctx, currentUserID, deletedOrgID, m.Client())
}

// updateOrgSubscriptionOnDelete updates the org subscription to inactive and sets the status to canceled
func updateOrgSubscriptionOnDelete(ctx context.Context, m *generated.OrganizationMutation) error {
	deletedOrgID, ok := m.ID()
	if !ok || deletedOrgID == "" {
		return nil
	}

	if err := m.Client().OrgSubscription.Update().Where(orgsubscription.OwnerID(deletedOrgID)).SetActive(false).SetExpiresAt(time.Now()).Exec(ctx); err != nil {
		return err
	}

	return nil
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
		// get the first org that was not the org being deleted and where the user is a member
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

// defaultOrganizationSettings creates the default organizations settings for a new org
func defaultOrganizationSettings(ctx context.Context, m *generated.OrganizationMutation) (string, error) {
	input := generated.CreateOrganizationSettingInput{}

	personalOrg, _ := m.PersonalOrg()

	if !personalOrg {
		userID, err := auth.GetSubjectIDFromContext(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("unable to get user id from context")
			return "", err
		}

		user, err := m.Client().User.Get(ctx, userID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("unable to fetch user from database")
			return "", err
		}

		billingContact := user.FirstName + " " + user.LastName
		input.BillingEmail = &user.Email
		input.BillingContact = &billingContact

		// automatically add the user's email domain to the allowed email domains
		// and enable auto-join for matching domains
		emailDomain := strings.SplitAfter(user.Email, "@")[1]

		input.AllowedEmailDomains = []string{emailDomain}
		input.AllowMatchingDomainsAutojoin = lo.ToPtr(true)
	}

	organizationSetting, err := m.Client().OrganizationSetting.Create().SetInput(input).Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating organization settings")
		return "", err
	}

	return organizationSetting.ID, nil
}

// personalOrgNoChildren checks if the mutation is for a child org, and if so returns an error
// if the parent org is a personal org
func personalOrgNoChildren(ctx context.Context, m *generated.OrganizationMutation) error {
	// check if this is a child org, error if parent org is a personal org
	parentOrgID, ok := m.ParentID()
	if ok {
		// check if parent org is a personal org
		parentOrg, err := m.Client().Organization.Query().
			Select("personal_org").
			Where(organization.ID(parentOrgID)).
			Only(ctx)
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
		SubjectType: generated.TypeOrganization,
		ObjectID:    childOrgID,
		ObjectType:  generated.TypeOrganization,
		Relation:    fgax.ParentRelation,
	}

	tuple := fgax.GetTupleKey(req)

	if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{
		tuple,
	}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create relationship tuple")

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
		return addTokenEditPermissions(ctx, m, oID, GetObjectTypeFromEntMutation(m))
	}

	// get userID from context
	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to get user id from context, unable to add user to organization")

		return err
	}

	// Add User as owner of organization
	owner := enums.RoleOwner
	input := generated.CreateOrgMembershipInput{
		UserID:         userID,
		OrganizationID: oID,
		Role:           &owner,
	}

	if err := m.Client().OrgMembership.Create().SetInput(input).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating org membership for owner")

		return err
	}

	// set the user's default org to the new org
	return updateDefaultOrgIfPersonal(ctx, userID, oID, m.Client())
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
		if err = client.UserSetting.
			UpdateOneID(userSetting.ID).
			SetDefaultOrgID(orgID).
			Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}
