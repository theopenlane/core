package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

const (
	customerEntityTypeName = "customer"
)

// HookTrustcenterEntityCreate scopes the entity to the customer type by default.
// If the customer entity does not exist ( maybe old orgs ), it creates it before proceeding to the
// trustcenter entity creation
func HookTrustcenterEntityCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustcenterEntityFunc(func(ctx context.Context, m *generated.TrustcenterEntityMutation) (generated.Value, error) {

			trustcenterID, err := m.Client().TrustCenter.Query().OnlyID(ctx)
			if generated.IsNotFound(err) {
				return nil, err
			}

			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("unable to fetch trustcenter")
				return nil, ErrTrustCenterIDRequired
			}

			m.SetTrustCenterID(trustcenterID)

			existingEntityType, err := m.Client().EntityType.Query().
				Where(
					entitytype.NameEqualFold(customerEntityTypeName),
				).
				Only(ctx)

			if err == nil {
				m.SetEntityTypeID(existingEntityType.ID)
				return next.Mutate(ctx, m)
			}

			if !generated.IsNotFound(err) {
				logx.FromContext(ctx).Error().Err(err).Msg("unable to query for customer entity type")
				return nil, err
			}

			entity, err := m.Client().EntityType.Create().
				SetName(customerEntityTypeName).
				Save(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("unable to create customer entity type")
				return nil, err
			}

			m.SetEntityTypeID(entity.ID)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// HookTrustcenterEntityFiles runs on trustcenter entity mutations
// and checks for an uploaded logo file
func HookTrustcenterEntityFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustcenterEntityFunc(func(ctx context.Context, m *generated.TrustcenterEntityMutation) (generated.Value, error) {
			files, _ := pkgobjects.FilesFromContextWithKey(ctx, "logoFile")
			if len(files) > 0 {
				var err error

				ctx, err = checkTrustcenterEntityFiles(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

func checkTrustcenterEntityFiles(ctx context.Context, m *generated.TrustcenterEntityMutation) (context.Context, error) {
	key := "logoFile"

	files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	if len(files) > 1 {
		return ctx, ErrTooManyAvatarFiles
	}

	m.SetLogoFileID(files[0].ID)

	adapter := pkgobjects.NewGenericMutationAdapter(m,
		func(mut *generated.TrustcenterEntityMutation) (string, bool) { return mut.ID() },
		func(mut *generated.TrustcenterEntityMutation) string { return mut.Type() },
	)

	return pkgobjects.ProcessFilesForMutation(ctx, adapter, key, "trustcenter_entity")
}
