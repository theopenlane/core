package hooks

import (
	"context"
	"net/mail"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

type directoryAccountIdentityResolutionSkipKey struct{}

type directoryAccountIdentityMatchBasis string

const (
	directoryAccountIdentityMatchBasisLinkedDirectoryEmail directoryAccountIdentityMatchBasis = "linked_directory_account_email"
	directoryAccountIdentityMatchBasisHolderEmail          directoryAccountIdentityMatchBasis = "identity_holder_email"
	directoryAccountIdentityMatchBasisHolderAlternateEmail directoryAccountIdentityMatchBasis = "identity_holder_alternate_email"
	directoryAccountIdentityMatchBasisCreated              directoryAccountIdentityMatchBasis = "created_from_directory_account"
)

// HookDirectoryAccount links human directory accounts to identity holders
// The hook is best-effort so directory evidence persists even when identity
// resolution is ambiguous or temporarily unavailable
func HookDirectoryAccount() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.DirectoryAccountFunc(func(ctx context.Context, m *generated.DirectoryAccountMutation) (generated.Value, error) {
			if shouldSkipDirectoryAccountIdentityResolution(ctx) {
				return next.Mutate(ctx, m)
			}

			identityLinkManaged := m.IdentityHolderIDCleared()
			if _, ok := m.IdentityHolderID(); ok {
				identityLinkManaged = true
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			if identityLinkManaged {
				return retVal, nil
			}

			account := retVal.(*generated.DirectoryAccount)

			if err := reconcileDirectoryAccountIdentityHolder(privacy.DecisionContext(ctx, privacy.Allow), m.Client(), account); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("directory_account_id", account.ID).Msg("failed to reconcile identity holder for directory account")
			}

			return retVal, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// reconcileDirectoryAccountIdentityHolder links or creates an identity holder for a directory account
func reconcileDirectoryAccountIdentityHolder(ctx context.Context, client *generated.Client, account *generated.DirectoryAccount) error {
	if account.IdentityHolderID != nil {
		if !isHumanDirectoryAccount(account.AccountType) {
			return nil
		}

		if err := syncIdentityHolderLifecycleFromDirectoryAccounts(ctx, client, *account.IdentityHolderID); err != nil {
			return err
		}

		return enrichIdentityHolderFromDirectoryAccount(ctx, client, *account.IdentityHolderID, account)
	}

	if !isHumanDirectoryAccount(account.AccountType) {
		return nil
	}

	email, ok := directoryAccountCanonicalEmail(account)
	if !ok {
		return nil
	}

	holder, basis, ambiguous, err := resolveExistingIdentityHolder(ctx, client, account.OwnerID, email)
	if err != nil {
		return err
	}

	if ambiguous {
		logx.FromContext(ctx).Warn().Str("directory_account_id", account.ID).Str("owner_id", account.OwnerID).Str("canonical_email", email).Msg("directory account matched multiple identity holders, leaving account unlinked")

		return nil
	}

	if holder == nil {
		holder, err = createIdentityHolderFromDirectoryAccount(ctx, client, account, email)
		if err != nil {
			return err
		}

		basis = directoryAccountIdentityMatchBasisCreated
	}

	if err := linkDirectoryAccountToIdentityHolder(ctx, client, account, holder.ID); err != nil {
		return err
	}

	logx.FromContext(ctx).Debug().Str("directory_account_id", account.ID).Str("identity_holder_id", holder.ID).Str("basis", string(basis)).Msg("linked directory account to identity holder")

	if err := syncIdentityHolderLifecycleFromDirectoryAccounts(ctx, client, holder.ID); err != nil {
		return err
	}

	return enrichIdentityHolderFromDirectoryAccount(ctx, client, holder.ID, account)
}

// resolveExistingIdentityHolder finds a unique existing holder for a directory account email
func resolveExistingIdentityHolder(ctx context.Context, client *generated.Client, ownerID, email string) (*generated.IdentityHolder, directoryAccountIdentityMatchBasis, bool, error) {
	holder, ambiguous, err := lookupHolderFromLinkedDirectoryAccounts(ctx, client, ownerID, email)
	if err != nil || holder != nil || ambiguous {
		return holder, directoryAccountIdentityMatchBasisLinkedDirectoryEmail, ambiguous, err
	}

	holder, ambiguous, err = lookupUniqueIdentityHolderByEmail(ctx, client, ownerID, email)
	if err != nil || holder != nil || ambiguous {
		return holder, directoryAccountIdentityMatchBasisHolderEmail, ambiguous, err
	}

	holder, ambiguous, err = lookupUniqueIdentityHolderByAlternateEmail(ctx, client, ownerID, email)
	if err != nil || holder != nil || ambiguous {
		return holder, directoryAccountIdentityMatchBasisHolderAlternateEmail, ambiguous, err
	}

	return nil, "", false, nil
}

// lookupHolderFromLinkedDirectoryAccounts finds a holder through already-linked directory accounts
func lookupHolderFromLinkedDirectoryAccounts(ctx context.Context, client *generated.Client, ownerID, email string) (*generated.IdentityHolder, bool, error) {
	accounts, err := client.DirectoryAccount.Query().
		Where(
			directoryaccount.OwnerID(ownerID),
			directoryaccount.CanonicalEmailEqualFold(email),
			directoryaccount.IdentityHolderIDNotNil(),
		).
		All(ctx)
	if err != nil {
		return nil, false, err
	}

	holderIDs := mapx.MapSetFromSlice(lo.Map(accounts, func(account *generated.DirectoryAccount, _ int) string {
		return lo.FromPtr(account.IdentityHolderID)
	}))

	switch len(holderIDs) {
	case 0:
		return nil, false, nil
	case 1:
		var holderID string
		for id := range holderIDs {
			holderID = id
		}

		holder, err := client.IdentityHolder.Query().
			Where(identityholder.ID(holderID), identityholder.OwnerID(ownerID)).
			Only(ctx)
		switch {
		case err == nil:
			return holder, false, nil
		case generated.IsNotFound(err):
			return nil, false, nil
		default:
			return nil, false, err
		}
	default:
		return nil, true, nil
	}
}

// lookupUniqueIdentityHolderByEmail finds a unique holder by primary email
func lookupUniqueIdentityHolderByEmail(ctx context.Context, client *generated.Client, ownerID, email string) (*generated.IdentityHolder, bool, error) {
	holders, err := client.IdentityHolder.Query().
		Where(identityholder.OwnerID(ownerID), identityholder.EmailEqualFold(email)).
		Limit(2).
		All(ctx)
	if err != nil {
		return nil, false, err
	}

	switch len(holders) {
	case 0:
		return nil, false, nil
	case 1:
		return holders[0], false, nil
	default:
		return nil, true, nil
	}
}

// lookupUniqueIdentityHolderByAlternateEmail finds a unique holder by alternate email
func lookupUniqueIdentityHolderByAlternateEmail(ctx context.Context, client *generated.Client, ownerID, email string) (*generated.IdentityHolder, bool, error) {
	holders, err := client.IdentityHolder.Query().
		Where(identityholder.OwnerID(ownerID), identityholder.AlternateEmailEqualFold(email)).
		Limit(2).
		All(ctx)
	if err != nil {
		return nil, false, err
	}

	switch len(holders) {
	case 0:
		return nil, false, nil
	case 1:
		return holders[0], false, nil
	default:
		return nil, true, nil
	}
}

// createIdentityHolderFromDirectoryAccount creates a holder from directory account evidence
func createIdentityHolderFromDirectoryAccount(ctx context.Context, client *generated.Client, account *generated.DirectoryAccount, email string) (*generated.IdentityHolder, error) {
	create := client.IdentityHolder.Create().
		SetOwnerID(account.OwnerID).
		SetEmail(email).
		SetFullName(identityHolderPreferredNameFromDirectoryAccount(account, email)).
		SetIdentityHolderType(enums.IdentityHolderTypeUnspecified)

	if title := lo.FromPtr(account.JobTitle); title != "" {
		create.SetTitle(title)
	}

	if department := lo.FromPtr(account.Department); department != "" {
		create.SetDepartment(department)
	}

	holder, err := create.Save(ctx)
	if err == nil {
		return holder, nil
	}

	if !generated.IsConstraintError(err) {
		return nil, err
	}

	holder, ambiguous, lookupErr := lookupUniqueIdentityHolderByEmail(ctx, client, account.OwnerID, email)
	if lookupErr != nil {
		return nil, lookupErr
	}

	if ambiguous || holder == nil {
		return nil, err
	}

	return holder, nil
}

// linkDirectoryAccountToIdentityHolder stores the resolved holder link on the directory account
func linkDirectoryAccountToIdentityHolder(ctx context.Context, client *generated.Client, account *generated.DirectoryAccount, holderID string) error {
	updateCtx := withSkipDirectoryAccountIdentityResolution(ctx)

	if err := client.DirectoryAccount.UpdateOneID(account.ID).SetIdentityHolderID(holderID).Exec(updateCtx); err != nil {
		return err
	}

	account.IdentityHolderID = &holderID

	return nil
}

// enrichIdentityHolderFromDirectoryAccount fills empty holder fields from directory evidence
func enrichIdentityHolderFromDirectoryAccount(ctx context.Context, client *generated.Client, holderID string, account *generated.DirectoryAccount) error {
	holder, err := client.IdentityHolder.Get(ctx, holderID)
	switch {
	case err == nil:
	case generated.IsNotFound(err):
		return nil
	default:
		return err
	}

	update := client.IdentityHolder.UpdateOneID(holderID)
	changed := false

	if name := identityHolderDescriptiveNameFromDirectoryAccount(account); name != "" && holder.FullName == holder.Email {
		update.SetFullName(name)
		changed = true
	}

	if title := lo.FromPtr(account.JobTitle); title != "" && holder.Title == "" {
		update.SetTitle(title)
		changed = true
	}

	if department := lo.FromPtr(account.Department); department != "" && holder.Department == "" {
		update.SetDepartment(department)
		changed = true
	}

	if !changed {
		return nil
	}

	return update.Exec(ctx)
}

// syncIdentityHolderLifecycleFromDirectoryAccounts updates holder lifecycle fields from linked directory accounts
func syncIdentityHolderLifecycleFromDirectoryAccounts(ctx context.Context, client *generated.Client, holderID string) error {
	holder, err := client.IdentityHolder.Get(ctx, holderID)
	switch {
	case err == nil:
	case generated.IsNotFound(err):
		return nil
	default:
		return err
	}

	status, isActive, err := aggregateIdentityHolderLifecycle(ctx, client, holder.OwnerID, holderID)
	if err != nil {
		return err
	}

	if holder.Status == status && holder.IsActive == isActive {
		return nil
	}

	return client.IdentityHolder.UpdateOneID(holderID).
		SetStatus(status).
		SetIsActive(isActive).
		Exec(ctx)
}

// aggregateIdentityHolderLifecycle collapses linked directory account states into a holder status
func aggregateIdentityHolderLifecycle(ctx context.Context, client *generated.Client, ownerID, holderID string) (enums.UserStatus, bool, error) {
	accounts, err := client.DirectoryAccount.Query().
		Where(
			directoryaccount.OwnerID(ownerID),
			directoryaccount.IdentityHolderID(holderID),
		).
		All(ctx)
	if err != nil {
		return "", false, err
	}

	effectiveStatuses := mapx.MapSetFromSlice(lo.FilterMap(accounts, func(account *generated.DirectoryAccount, _ int) (enums.DirectoryAccountStatus, bool) {
		if !isHumanDirectoryAccount(account.AccountType) {
			return "", false
		}

		return effectiveDirectoryAccountStatus(account), true
	}))

	if len(effectiveStatuses) != 1 {
		return enums.UserStatusUnknown, false, nil
	}

	for status := range effectiveStatuses {
		return identityHolderLifecycleFromEffectiveDirectoryStatus(status), status == enums.DirectoryAccountStatusActive, nil
	}

	return enums.UserStatusUnknown, false, nil
}

// effectiveDirectoryAccountStatus derives the status used for holder lifecycle aggregation
func effectiveDirectoryAccountStatus(account *generated.DirectoryAccount) enums.DirectoryAccountStatus {
	if account.RemovedAt != nil {
		return enums.DirectoryAccountStatusDeleted
	}

	return account.Status
}

// identityHolderLifecycleFromEffectiveDirectoryStatus maps a directory lifecycle state to a holder lifecycle state
func identityHolderLifecycleFromEffectiveDirectoryStatus(status enums.DirectoryAccountStatus) enums.UserStatus {
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

// shouldSkipDirectoryAccountIdentityResolution reports whether hook recursion should be bypassed
func shouldSkipDirectoryAccountIdentityResolution(ctx context.Context) bool {
	v, _ := ctx.Value(directoryAccountIdentityResolutionSkipKey{}).(bool)
	return v
}

// withSkipDirectoryAccountIdentityResolution marks a context to bypass recursive hook re-entry
func withSkipDirectoryAccountIdentityResolution(ctx context.Context) context.Context {
	return context.WithValue(ctx, directoryAccountIdentityResolutionSkipKey{}, true)
}

// isHumanDirectoryAccount reports whether the directory account should participate in holder resolution
func isHumanDirectoryAccount(accountType enums.DirectoryAccountType) bool {
	switch accountType {
	case enums.DirectoryAccountTypeUser, enums.DirectoryAccountTypeGuest:
		return true
	default:
		return false
	}
}

// directoryAccountCanonicalEmail returns the account email when it is usable for holder resolution
func directoryAccountCanonicalEmail(account *generated.DirectoryAccount) (string, bool) {
	email := lo.FromPtr(account.CanonicalEmail)
	if email == "" {
		return "", false
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", false
	}

	return email, true
}

// identityHolderDescriptiveNameFromDirectoryAccount returns the best descriptive name available from directory data
func identityHolderDescriptiveNameFromDirectoryAccount(account *generated.DirectoryAccount) string {
	if name := account.DisplayName; name != "" {
		if email, ok := directoryAccountCanonicalEmail(account); !ok || name != email {
			return name
		}
	}

	givenName := lo.FromPtr(account.GivenName)
	familyName := lo.FromPtr(account.FamilyName)

	switch {
	case givenName != "" && familyName != "":
		return givenName + " " + familyName
	case givenName != "":
		return givenName
	default:
		return familyName
	}
}

// identityHolderPreferredNameFromDirectoryAccount returns the required holder name value for auto-created holders
func identityHolderPreferredNameFromDirectoryAccount(account *generated.DirectoryAccount, email string) string {
	if name := identityHolderDescriptiveNameFromDirectoryAccount(account); name != "" {
		return name
	}

	return email
}
