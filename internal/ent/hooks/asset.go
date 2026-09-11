package hooks

import (
	"context"
	"net/mail"

	"entgo.io/ent"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
)

// HookAssetCreate sets the display name for assets everytime one is created
func HookAssetCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssetFunc(func(ctx context.Context, m *generated.AssetMutation) (generated.Value, error) {
			name, _ := m.DisplayName()
			if name == "" {
				name, _ := m.Name()
				m.SetDisplayName(name)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// HookAssetInternalOwner resolves the internal_owner field to an existing user when the value is
// a parseable email address belonging to a member of the asset's owner organization, setting
// internal_owner_user_id and clearing internal_owner; when no matching member is found the
// mutation is left untouched
func HookAssetInternalOwner() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssetFunc(func(ctx context.Context, m *generated.AssetMutation) (generated.Value, error) {
			rawValue, ok := m.InternalOwner()
			if !ok {
				return next.Mutate(ctx, m)
			}

			address, err := mail.ParseAddress(rawValue)
			if err != nil {
				return next.Mutate(ctx, m)
			}

			ownerID, err := assetOwnerOrgID(ctx, m)
			if err != nil {
				return nil, err
			}

			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

			userID, err := m.Client().User.Query().
				Where(
					user.EmailEqualFold(address.Address),
					user.HasOrgMembershipsWith(orgmembership.OrganizationID(ownerID)),
				).
				OnlyID(allowCtx)
			if err != nil {
				if generated.IsNotFound(err) {
					return next.Mutate(ctx, m)
				}

				return nil, err
			}

			m.SetInternalOwnerUserID(userID)
			m.ClearInternalOwner()

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// assetOwnerOrgID resolves the owner organization id for the asset being mutated, falling back
// to the persisted value on update since owner_id is not resent on update mutations; an asset
// created without an owner (system admin bypass) resolves to an empty id
func assetOwnerOrgID(ctx context.Context, m *generated.AssetMutation) (string, error) {
	if ownerID, ok := m.OwnerID(); ok {
		return ownerID, nil
	}

	if m.Op().Is(ent.OpUpdateOne) {
		return m.OldOwnerID(ctx)
	}

	return "", nil
}
