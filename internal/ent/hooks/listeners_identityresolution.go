package hooks

import (
	"context"
	"net/mail"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// IdentityResolutionListeners resolves directory accounts to identity holders after mutations commit
func IdentityResolutionListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaDirectoryAccount,
			Operations: []string{entityops.OpCreate, entityops.OpUpdateOne},
			Handle:     handleDirectoryAccountMutation,
		},
	}
}

// handleDirectoryAccountMutation runs the matching cascade for unlinked accounts and links
// the resolved holder; already-linked accounts only re-enrich and re-sync aliases
func handleDirectoryAccountMutation(inv entityops.Invocation, _ entityops.MutationPayload) error {
	account, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.DirectoryAccount.Get)
	if err != nil || !ok {
		return err
	}

	if lo.FromPtr(account.IdentityHolderID) != "" {
		holder, err := inv.Client.IdentityHolder.Get(inv.Context, *account.IdentityHolderID)
		if err != nil {
			if entgen.IsNotFound(err) {
				return nil
			}

			return err
		}

		return enrichAndSyncHolder(logx.WithFields(inv.Context, map[string]any{"identity_holder_id": holder.ID}), inv.Client, holder, account)
	}

	holder, err := resolveIdentityHolder(inv.Context, inv.Client, account)
	if err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("identity resolution failed")

		return err
	}

	if holder == nil {
		return nil
	}

	ctx := logx.WithFields(inv.Context, map[string]any{"identity_holder_id": holder.ID})

	if err := inv.Client.DirectoryAccount.UpdateOneID(account.ID).SetIdentityHolderID(holder.ID).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to link directory account to identity holder")

		return err
	}

	return enrichAndSyncHolder(ctx, inv.Client, holder, account)
}

// enrichAndSyncHolder enriches primary-source account and rebuilds its email aliases
func enrichAndSyncHolder(ctx context.Context, client *entgen.Client, holder *entgen.IdentityHolder, account *entgen.DirectoryAccount) error {
	if account.PrimarySource {
		if err := enrichFromPrimarySource(ctx, client, holder, account); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("primary source enrichment failed")

			return err
		}
	}

	if err := syncEmailAliases(ctx, client, holder); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("email alias sync failed")

		return err
	}

	return nil
}

// resolveIdentityHolder runs a priority-ordered matching cascade to find or create an
// identity holder for the directory account
func resolveIdentityHolder(ctx context.Context, client *entgen.Client, account *entgen.DirectoryAccount) (*entgen.IdentityHolder, error) {
	ownerID := account.OwnerID
	emails := confirmedAccountEmails(account)
	hasEmail := len(emails) > 0
	hasName := lo.FromPtr(account.GivenName) != "" && lo.FromPtr(account.FamilyName) != ""

	// exact canonical (SSO) email match on IdentityHolder
	if lo.FromPtr(account.CanonicalEmail) != "" {
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

	// any confirmed email match against identity holder primary email or aliases
	if hasEmail {
		holder, err := client.IdentityHolder.Query().
			Where(identityholder.OwnerID(ownerID),
				emailOrAliasMatch(identityholder.FieldEmail, identityholder.FieldEmailAliases, emails)).
			First(ctx)
		if err == nil {
			return holder, nil
		}

		if !entgen.IsNotFound(err) {
			return nil, err
		}
	}

	// check directory accounts for any that might be linked to an identity holder
	// through their canonical email or confirmed aliases
	if hasEmail {
		sibling, err := client.DirectoryAccount.Query().
			Where(directoryaccount.OwnerID(ownerID),
				directoryaccount.IdentityHolderIDNotNil(),
				directoryaccount.IDNEQ(account.ID),
				emailOrAliasMatch(directoryaccount.FieldCanonicalEmail, directoryaccount.FieldEmailAliases, emails)).
			First(ctx)
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
	if lo.FromPtr(account.CanonicalEmail) == "" {
		return nil, nil
	}

	return createIdentityHolder(ctx, client, account)
}

// emailOrAliasMatch matches rows whose email column equals any of the given emails or
// whose alias JSON column contains one
func emailOrAliasMatch(emailColumn, aliasColumn string, emails []string) func(*sql.Selector) {
	return func(s *sql.Selector) {
		predicates := append(
			[]*sql.Predicate{sql.In(s.C(emailColumn), lo.ToAnySlice(emails)...)},
			lo.Map(emails, func(email string, _ int) *sql.Predicate {
				return sqljson.ValueContains(aliasColumn, email)
			})...,
		)

		s.Where(sql.Or(predicates...))
	}
}

// confirmedAccountEmails returns the canonical email followed by every confirmed alias
// for a directory account, deduplicated with empties removed
func confirmedAccountEmails(account *entgen.DirectoryAccount) []string {
	return lo.Uniq(lo.Compact(append([]string{lo.FromPtr(account.CanonicalEmail)}, account.EmailAliases...)))
}

// createIdentityHolder creates a new identity holder from a directory account with
// conservative defaults, using primary source fields when available
func createIdentityHolder(ctx context.Context, client *entgen.Client, account *entgen.DirectoryAccount) (*entgen.IdentityHolder, error) {
	canonicalEmail := lo.FromPtr(account.CanonicalEmail)

	// check if it is a valid email, otherwise skip creation; identity holders are required to have a valid email address
	if _, err := mail.ParseAddress(canonicalEmail); err != nil {
		logx.FromContext(ctx).Info().Str("email", canonicalEmail).Msg("identityholder: email is not a valid address, skipping identity holder creation for directory account")

		return nil, nil
	}

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
// directory accounts' canonical emails and confirmed aliases, excluding the primary email
func syncEmailAliases(ctx context.Context, client *entgen.Client, holder *entgen.IdentityHolder) error {
	accounts, err := client.DirectoryAccount.Query().
		Where(directoryaccount.IdentityHolderID(holder.ID)).
		Select(directoryaccount.FieldCanonicalEmail, directoryaccount.FieldEmailAliases).All(ctx)
	if err != nil {
		return err
	}

	aliases := lo.Uniq(lo.FlatMap(accounts, func(a *entgen.DirectoryAccount, _ int) []string {
		return lo.Filter(confirmedAccountEmails(a), func(email string, _ int) bool {
			return !strings.EqualFold(email, holder.Email)
		})
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
