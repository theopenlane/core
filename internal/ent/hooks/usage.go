package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/rs/zerolog"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/usage"
	usagepkg "github.com/theopenlane/core/internal/usage"
	"github.com/theopenlane/core/pkg/enums"
)

// HookUsage caps and increments usage for the provided type - this function is called from the mixin
func HookUsage(t enums.UsageType) ent.Hook {
	return hookUsageBase(t)
}

// HookUsageUsers tracks member count on OrgMembership records
func HookUsageUsers() ent.Hook {
	return hookUsageBase(enums.UsageUsers)
}

// HookUsagePrograms tracks the number of programs owned by an organization
func HookUsagePrograms() ent.Hook {
	return hookUsageBase(enums.UsagePrograms)
}

// hookUsageCount increments usage on create and decrements on delete
func hookUsageBase(t enums.UsageType) ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			var orgID string
			var ok bool

			// not sure if there is a better way to do this, but we want only mutations with a valid organization context and client proceed
			// since there could be record types that don't have the organization as an "owner" (just an edge for example) we need to check for both
			switch om := m.(type) {
			case interface{ OwnerID() (string, bool) }:
				orgID, ok = om.OwnerID()
			case interface{ OrganizationID() (string, bool) }:
				orgID, ok = om.OrganizationID()
			}

			// we also have to ensure that the mutation has a client to use because we can't query the usage without it
			// if the type assertion fails we just allow the mutation to proceed without blocking
			cm, _ := m.(interface{ Client() *generated.Client })
			if !ok || orgID == "" || cm == nil {
				return next.Mutate(ctx, m)
			}

			// on create, we check if the usage limit is reached before allowing the mutation to proceed
			// after successful mutation creation we update the usage
			// on delete, we let the mutation proceed and then adjust usage after it's successful
			switch {
			case m.Op().Is(ent.OpCreate):
				if err := usagepkg.CheckUsageDelta(ctx, cm.Client(), orgID, t, 1); err != nil {
					return nil, err
				}

				v, err := next.Mutate(ctx, m)
				if err != nil {
					return v, err
				}

				// there are going to be edge cases with bulk creation and other things we can't really handle here, so keeping it simple and just incrementing the usage of the associated record by 1
				if err = adjustUsage(ctx, cm.Client(), orgID, t, 1); err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("failed to update usage")
				}

				return v, err
			case m.Op().Is(ent.OpDelete | ent.OpDeleteOne):
				v, err := next.Mutate(ctx, m)
				if err != nil {
					return v, err
				}

				// there are going to be edge cases with bulk deletes and other things that we can't really handle here, so keeping it simple and just decrementing the usage of the associated record by 1
				if err = adjustUsage(ctx, cm.Client(), orgID, t, -1); err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("failed to update usage")
				}

				return v, err
			default:
				return next.Mutate(ctx, m)
			}
		})
	}, ent.OpCreate|ent.OpDelete|ent.OpDeleteOne)
}

// updateStorageUsage updates the storage usage for an organization
func updateStorageUsage(ctx context.Context, client *generated.Client, orgID string, delta int64) error {
	return adjustUsage(ctx, client, orgID, enums.UsageStorage, delta)
}

// adjustUsage updates the usage record for the given type by delta
// The error is returned so callers can decide whether a failed adjustment should abort the surrounding mutation
func adjustUsage(ctx context.Context, client *generated.Client, orgID string, t enums.UsageType, delta int64) error {
	u, errQuery := client.Usage.Query().Where(usage.OrganizationID(orgID), usage.ResourceTypeEQ(t)).Only(ctx)

	var updated *generated.Usage
	var err error

	zerolog.Ctx(ctx).Debug().Str("org_id", orgID).Str("type", t.String()).Int64("delta", delta).Msg("adjusting usage")
	switch {
	case errQuery == nil && u != nil:
		updated, err = client.Usage.UpdateOne(u).AddUsed(delta).Save(ctx)
	case generated.IsNotFound(errQuery):
		updated, err = client.Usage.Create().SetOrganizationID(orgID).SetResourceType(t).SetUsed(delta).Save(ctx)
	default:
		return errQuery
	}

	if err == nil {
		usagepkg.RecordUsageUpdate(t, "update")
		usagepkg.EmitThresholdEvents(ctx, client, orgID, t)

		zerolog.Ctx(ctx).Debug().Str("org_id", orgID).Str("type", t.String()).Int64("limit", updated.Limit).Int64("used", updated.Used).Msg("usage updated")
	}

	return err
}

// HookUsageStorage updates storage usage related to File objects using the persisted file size of File records
// storage is a bit of a special use case and can't simply rely on number of file records
// because the size of the files can vary greatly, so we need to track the size of the files in addition to the number of records
func HookUsageStorage() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.FileFunc(func(ctx context.Context, m *generated.FileMutation) (generated.Value, error) {
			// handle deletes to subtract the size of the deleted file from the org's storage usage
			if isDeleteOp(ctx, m) {
				id, ok := m.ID()
				if !ok {
					return nil, fmt.Errorf("%w for delete", ErrMissingFileID)
				}

				// we need to get the file from the database to get the persisted file size before the actual deletion occurs
				f, err := m.Client().File.Get(ctx, id)
				if err != nil {
					return nil, err
				}

				// once we have the file information, we can let the mut proceed with the delete
				v, err := next.Mutate(ctx, m)
				if err != nil {
					return v, err
				}

				// after successful deletion, update the storage usage
				err = updateStorageUsage(ctx, m.Client(), f.OrganizationID, -f.PersistedFileSize)
				if err != nil {
					zerolog.Ctx(ctx).Error().Err(err).Msg("failed to update storage usage")
				}

				return v, err
			}

			// usage is tracked by organization, so we need to get the orgID from the mutation
			orgID, ok := m.OrganizationID()
			if !ok || orgID == "" {
				return next.Mutate(ctx, m)
			}

			var delta int64

			// for file creation, we add the new file's size to the storage usage
			switch {
			case m.Op().Is(ent.OpCreate):
				delta, _ = m.PersistedFileSize()
			// for updates, we check if the persisted file size has changed and update the usage accordingly
			case m.Op().Is(ent.OpUpdate) || m.Op().Is(ent.OpUpdateOne):
				if _, ok := m.PersistedFileSize(); ok {
					id, okID := m.ID()
					if !okID {
						return nil, fmt.Errorf("%w for update", ErrMissingFileID)
					}

					f, err := m.Client().File.Get(ctx, id)
					if err != nil {
						return nil, err
					}

					newSize, _ := m.PersistedFileSize()
					delta = newSize - f.PersistedFileSize
				}
			}

			// if the delta is 0, we can skip the usage check and just call the next mutator
			if delta == 0 {
				return next.Mutate(ctx, m)
			}

			// we want to check if the new size is within the usage limit before proceeding so that if the updated file (or updating files in-place) exceeds the limit, we can prevent the mutation
			if err := usagepkg.CheckUsageDelta(ctx, m.Client(), orgID, enums.UsageStorage, delta); err != nil {
				return nil, err
			}

			// if the usage check passes, we can proceed with the mutation
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			// after the mutation, we update the storage usage with the delta ensuring we only update usage after successful database operation and storage occurs
			err = updateStorageUsage(ctx, m.Client(), orgID, delta)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("failed to update storage usage")
			}

			return v, err
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne)
}
