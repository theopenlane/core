package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	customerEntityTypeName = "customer"
)

// HookTrustCenterEntityCreate scopes the entity to the customer type by default.
// If the customer entity does not exist ( maybe old orgs ), it creates it before proceeding to the
// trustcenter entity creation
func HookTrustCenterEntityCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterEntityFunc(func(ctx context.Context, m *generated.TrustCenterEntityMutation) (generated.Value, error) {
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

// HookTrustCenterEntityFiles runs on trustcenter entity mutations
// and checks for an uploaded logo file
func HookTrustCenterEntityFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterEntityFunc(func(ctx context.Context, m *generated.TrustCenterEntityMutation) (generated.Value, error) {
			var err error

			ctx, err = checkTrustCenterEntityFiles(ctx, m)
			if err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// checkTrustCenterEntityFiles checks for logo files in the context
func checkTrustCenterEntityFiles(ctx context.Context, m *generated.TrustCenterEntityMutation) (context.Context, error) {
	key := "logoFile"
	ctx, err := processSingleMutationFile(ctx, m, key, "trust_center_entity", ErrTooManyAvatarFiles,
		func(mut *generated.TrustCenterEntityMutation, id string) { mut.SetLogoFileID(id) },
		func(mut *generated.TrustCenterEntityMutation) (string, bool) { return mut.ID() },
		func(mut *generated.TrustCenterEntityMutation) string { return mut.Type() },
	)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
