package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaIdentityResolutionListeners registers listeners that resolve directory
// accounts to identity holders asynchronously after mutations commit
func RegisterGalaIdentityResolutionListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeDirectoryAccount),
			Name:       "identityresolution.directory_account_created",
			Operations: []string{ent.OpCreate.String()},
			Handle:     handleDirectoryAccountCreated,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeDirectoryAccount),
			Name:       "identityresolution.directory_account_updated",
			Operations: []string{ent.OpUpdateOne.String()},
			Handle:     handleDirectoryAccountUpdated,
		},
	)
}

// handleDirectoryAccountCreated runs full identity resolution for a newly created directory account
func handleDirectoryAccountCreated(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	accountID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || accountID == "" {
		return nil
	}

	account, err := client.DirectoryAccount.Get(ctx.Context, accountID)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil
		}

		return err
	}

	holder, err := resolveIdentityHolder(ctx.Context, client, account)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Str("directory_account_id", accountID).Msg("identity resolution failed")

		return err
	}

	if holder == nil {
		return nil
	}

	if err := client.DirectoryAccount.UpdateOneID(account.ID).SetIdentityHolderID(holder.ID).Exec(ctx.Context); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Str("directory_account_id", accountID).Str("identity_holder_id", holder.ID).Msg("failed to link directory account to identity holder")

		return err
	}

	if account.PrimarySource {
		if err := enrichFromPrimarySource(ctx.Context, client, holder, account); err != nil {
			logx.FromContext(ctx.Context).Error().Err(err).Str("identity_holder_id", holder.ID).Msg("primary source enrichment failed")

			return err
		}
	}

	if err := syncEmailAliases(ctx.Context, client, holder); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Str("identity_holder_id", holder.ID).Msg("email alias sync failed")

		return err
	}

	return nil
}

// handleDirectoryAccountUpdated re-enriches and syncs aliases when an existing directory account is updated
func handleDirectoryAccountUpdated(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	accountID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || accountID == "" {
		return nil
	}

	account, err := client.DirectoryAccount.Get(ctx.Context, accountID)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil
		}

		return err
	}

	// If no identity holder linked yet, attempt full resolution
	if account.IdentityHolderID == nil || *account.IdentityHolderID == "" {
		holder, err := resolveIdentityHolder(ctx.Context, client, account)
		if err != nil {
			logx.FromContext(ctx.Context).Error().Err(err).Str("directory_account_id", accountID).Msg("identity resolution on update failed")

			return err
		}

		if holder == nil {
			return nil
		}

		if err := client.DirectoryAccount.UpdateOneID(account.ID).SetIdentityHolderID(holder.ID).Exec(ctx.Context); err != nil {
			logx.FromContext(ctx.Context).Error().Err(err).Str("directory_account_id", accountID).Str("identity_holder_id", holder.ID).Msg("failed to link directory account to identity holder on update")

			return err
		}

		if account.PrimarySource {
			if err := enrichFromPrimarySource(ctx.Context, client, holder, account); err != nil {
				logx.FromContext(ctx.Context).Error().Err(err).Str("identity_holder_id", holder.ID).Msg("primary source enrichment failed")

				return err
			}
		}

		return syncEmailAliases(ctx.Context, client, holder)
	}

	holderID := *account.IdentityHolderID

	holder, err := client.IdentityHolder.Get(ctx.Context, holderID)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil
		}

		return err
	}

	if account.PrimarySource {
		if err := enrichFromPrimarySource(ctx.Context, client, holder, account); err != nil {
			logx.FromContext(ctx.Context).Error().Err(err).Str("identity_holder_id", holderID).Msg("primary source enrichment failed on update")

			return err
		}
	}

	return syncEmailAliases(ctx.Context, client, holder)
}

// resolveIdentityHolder runs a priority-ordered matching cascade to find or create
// an identity holder for the given directory account
func resolveIdentityHolder(ctx context.Context, client *entgen.Client, account *entgen.DirectoryAccount) (*entgen.IdentityHolder, error) {
	ownerID := account.OwnerID
	hasEmail := account.CanonicalEmail != nil && *account.CanonicalEmail != ""
	hasName := account.GivenName != nil && *account.GivenName != "" &&
		account.FamilyName != nil && *account.FamilyName != ""

	// exact email match on IdentityHolder
	if hasEmail {
		holder, err := client.IdentityHolder.Query().
			Where(identityholder.OwnerID(ownerID),
				identityholder.Email(*account.CanonicalEmail)).Only(ctx)
		if err == nil {
			return holder, nil
		}

		if !entgen.IsNotFound(err) {
			return nil, err
		}
	}

	// check directory accounts for any that might be linked to identity holder but didn't match on email
	if hasEmail {
		sibling, err := client.DirectoryAccount.Query().
			Where(directoryaccount.OwnerID(ownerID),
				directoryaccount.CanonicalEmail(*account.CanonicalEmail),
				directoryaccount.IdentityHolderIDNotNil(),
				directoryaccount.IDNEQ(account.ID)).First(ctx)
		if err == nil && sibling.IdentityHolderID != nil {
			return client.IdentityHolder.Get(ctx, *sibling.IdentityHolderID)
		}

		if err != nil && !entgen.IsNotFound(err) {
			return nil, err
		}
	}

	// name-based match via sibling directory accounts with no identity holder link
	if hasName {
		siblings, err := client.DirectoryAccount.Query().
			Where(directoryaccount.OwnerID(ownerID),
				directoryaccount.GivenName(*account.GivenName),
				directoryaccount.FamilyName(*account.FamilyName),
				directoryaccount.IdentityHolderIDNotNil(),
				directoryaccount.IDNEQ(account.ID)).All(ctx)
		if err != nil {
			return nil, err
		}

		holderIDs := lo.Uniq(lo.FilterMap(siblings, func(s *entgen.DirectoryAccount, _ int) (string, bool) {
			if s.IdentityHolderID == nil {
				return "", false
			}
			return *s.IdentityHolderID, true
		}))

		if len(holderIDs) == 1 {
			return client.IdentityHolder.Get(ctx, holderIDs[0])
		}
	}

	// Step 4: create new identity holder (requires canonical email)
	if !hasEmail {
		return nil, nil
	}

	return createIdentityHolder(ctx, client, account)
}

// createIdentityHolder creates a new identity holder from a directory account with
// conservative defaults, using primary source fields when available
func createIdentityHolder(ctx context.Context, client *entgen.Client, account *entgen.DirectoryAccount) (*entgen.IdentityHolder, error) {
	canonicalEmail := *account.CanonicalEmail
	exists := resolveIsOpenlaneUser(ctx, client, canonicalEmail)

	create := client.IdentityHolder.Create().
		SetOwnerID(account.OwnerID).
		SetEmail(canonicalEmail).
		SetIsOpenlaneUser(exists).
		SetFullName(buildFullName(account.DisplayName, account.GivenName, account.FamilyName, canonicalEmail))

	if account.PrimarySource {
		applyPrimarySourceDefaults(create, account)
	} else {
		create.SetStatus(enums.UserStatusUnknown)
	}

	holder, err := create.Save(ctx)
	if err == nil {
		return holder, nil
	}

	if !entgen.IsConstraintError(err) {
		return nil, err
	}

	// Race condition: another listener won the create; re-read the winner
	return client.IdentityHolder.Query().
		Where(identityholder.OwnerID(account.OwnerID),
			identityholder.Email(canonicalEmail)).Only(ctx)
}

// resolveIsOpenlaneUser checks if the user with the email is a member of the organization and sets to true if found
func resolveIsOpenlaneUser(ctx context.Context, client *entgen.Client, email string) bool {
	if email == "" {
		return false
	}

	exists, err := client.OrgMembership.Query().Where(
		orgmembership.HasUserWith(user.Email(email)),
	).Exist(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("email", email).Msg("error determining if users exists in organization")
	}

	return exists
}

// applyPrimarySourceDefaults sets authoritative fields on a new identity holder create builder from a primary source directory account
func applyPrimarySourceDefaults(create *entgen.IdentityHolderCreate, account *entgen.DirectoryAccount) {
	create.SetStatus(mapDirectoryAccountStatus(account.Status))
	create.SetIsActive(account.Status == enums.DirectoryAccountStatusActive)
	create.SetNillableTitle(account.JobTitle)
	create.SetNillableDepartment(account.Department)
	create.SetNillablePhoneNumber(account.PhoneNumber)
	create.SetExternalUserID(account.ExternalID)

	if account.AddedAt != nil {
		create.SetStartDate(models.DateTime(*account.AddedAt))
	}

	if account.RemovedAt != nil {
		create.SetEndDate(models.DateTime(*account.RemovedAt))
	}
}

// enrichFromPrimarySource updates an existing identity holder with authoritative fields from a primary source directory account
func enrichFromPrimarySource(ctx context.Context, client *entgen.Client, holder *entgen.IdentityHolder, account *entgen.DirectoryAccount) error {
	update := client.IdentityHolder.UpdateOneID(holder.ID)

	exists := resolveIsOpenlaneUser(ctx, client, holder.Email)

	update.SetIsOpenlaneUser(exists)
	update.SetStatus(mapDirectoryAccountStatus(account.Status))
	update.SetIsActive(account.Status == enums.DirectoryAccountStatusActive)
	update.SetExternalUserID(account.ExternalID)

	if name := buildFullName(account.DisplayName, account.GivenName, account.FamilyName, ""); name != "" {
		update.SetFullName(name)
	}

	update.SetNillablePhoneNumber(account.PhoneNumber)
	update.SetNillableTitle(account.JobTitle)
	update.SetNillableDepartment(account.Department)
	update.SetNillableAvatarRemoteURL(account.AvatarRemoteURL)

	if account.AddedAt != nil {
		update.SetStartDate(models.DateTime(*account.AddedAt))
	}

	if account.RemovedAt != nil {
		update.SetEndDate(models.DateTime(*account.RemovedAt))
	}

	return update.Exec(ctx)
}

// syncEmailAliases rebuilds the identity holder's email_aliases from all linked
// directory accounts' canonical emails, excluding the primary email
func syncEmailAliases(ctx context.Context, client *entgen.Client, holder *entgen.IdentityHolder) error {
	accounts, err := client.DirectoryAccount.Query().
		Where(directoryaccount.IdentityHolderID(holder.ID),
			directoryaccount.CanonicalEmailNotNil()).
		Select(directoryaccount.FieldCanonicalEmail).All(ctx)
	if err != nil {
		return err
	}

	aliases := lo.Uniq(lo.FilterMap(accounts, func(a *entgen.DirectoryAccount, _ int) (string, bool) {
		if a.CanonicalEmail == nil || *a.CanonicalEmail == "" || strings.EqualFold(*a.CanonicalEmail, holder.Email) {
			return "", false
		}

		return *a.CanonicalEmail, true
	}))

	return client.IdentityHolder.UpdateOneID(holder.ID).
		SetEmailAliases(aliases).
		Exec(ctx)
}

// mapDirectoryAccountStatus interprets a DirectoryAccountStatus into the corresponding UserStatus
func mapDirectoryAccountStatus(status enums.DirectoryAccountStatus) enums.UserStatus {
	switch status {
	case enums.DirectoryAccountStatusActive:
		return enums.UserStatusActive
	case enums.DirectoryAccountStatusInactive:
		return enums.UserStatusInactive
	case enums.DirectoryAccountStatusSuspended:
		return enums.UserStatusSuspended
	case enums.DirectoryAccountStatusDeleted:
		return enums.UserStatusDeactivated
	default:
		return enums.UserStatusUnknown
	}
}

// buildFullName returns the first non-empty name from: displayName, givenName+" "+familyName,
// givenName alone, familyName alone, or fallback
func buildFullName(displayName string, givenName, familyName *string, fallback string) string {
	if displayName != "" {
		return displayName
	}

	hasGiven := givenName != nil && *givenName != ""
	hasFamily := familyName != nil && *familyName != ""

	switch {
	case hasGiven && hasFamily:
		return *givenName + " " + *familyName
	case hasGiven:
		return *givenName
	case hasFamily:
		return *familyName
	default:
		return fallback
	}
}
