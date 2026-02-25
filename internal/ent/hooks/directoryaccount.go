package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
)

const directoryAccountIdentityHolderField = "identity_holder_id"

// HookDirectoryAccount links newly created directory accounts to an identity holder.
// The dedupe key is (owner_id, canonical_email).
func HookDirectoryAccount() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.DirectoryAccountFunc(func(ctx context.Context, m *generated.DirectoryAccountMutation) (generated.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			account, ok := retVal.(*generated.DirectoryAccount)
			if !ok {
				return retVal, nil
			}

			ownerID := account.OwnerID
			if ownerID == "" {
				return retVal, nil
			}

			// covers optional + nillable
			if account.CanonicalEmail == nil || *account.CanonicalEmail == "" {
				return retVal, nil
			}
			canonicalEmail := *account.CanonicalEmail

			// Run linking logic with privacy bypass so backend syncs can always reconcile identities
			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

			holder, err := getOrCreateIdentityHolder(allowCtx, m.Client(), ownerID, canonicalEmail, account.DisplayName, account.JobTitle, account.Department)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get or create identity holder for directory account")

				return nil, err
			}

			// Keep a direct link from directory account -> identity holder
			update := m.Client().DirectoryAccount.UpdateOneID(account.ID)
			if err := update.Mutation().SetField(directoryAccountIdentityHolderField, holder.ID); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to set identity holder edge field on directory account mutation")

				return nil, err
			}

			if err := update.Exec(allowCtx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to link directory account to identity holder")

				return nil, err
			}

			return retVal, nil
		})
	}, ent.OpCreate)
}

// getOrCreateIdentityHolder finds an identity holder by (owner_id, email) or creates one
func getOrCreateIdentityHolder(ctx context.Context, client *generated.Client, ownerID, canonicalEmail, displayName string, jobTitle, department *string) (*generated.IdentityHolder, error) {
	// First try to find an existing holder for this owner/email pair.
	holder, err := client.IdentityHolder.Query().
		Where(identityholder.OwnerID(ownerID), identityholder.Email(canonicalEmail)).
		Only(ctx)
	if err == nil {
		return holder, nil
	}

	if !generated.IsNotFound(err) {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to query for existing identity holder")

		return nil, err
	}

	create := client.IdentityHolder.Create().
		SetOwnerID(ownerID).
		SetEmail(canonicalEmail).
		// Keep full_name stable even when display_name is missing.
		SetFullName(lo.Ternary(displayName != "", displayName, canonicalEmail))

	if jobTitle != nil && *jobTitle != "" {
		create.SetTitle(*jobTitle)
	}

	if department != nil && *department != "" {
		create.SetDepartment(*department)
	}

	// Create when absent; if another writer wins, re-read the winner.
	holder, err = create.Save(ctx)
	if err == nil {
		return holder, nil
	}

	if !generated.IsConstraintError(err) {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create identity holder")

		return nil, err
	}

	return client.IdentityHolder.Query().
		Where(identityholder.OwnerID(ownerID), identityholder.Email(canonicalEmail)).
		Only(ctx)
}
