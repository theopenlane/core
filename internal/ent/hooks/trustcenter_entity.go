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

// HookTrustcenterEntityCreate scopes the entity to the customer type by default.
// If the customer entity does not exist ( maybe old orgs ), it creates it before proceeding to the
// trustcenter entity creation
func HookTrustcenterEntityCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustcenterEntityFunc(func(ctx context.Context, m *generated.TrustcenterEntityMutation) (generated.Value, error) {

			trustCenterID, ok := m.TrustCenterID()
			if !ok || trustCenterID == "" {
				return nil, ErrTrustCenterIDRequired
			}

			_, err := m.Client().TrustCenter.Get(ctx, trustCenterID)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("unable to get trust center")
				return nil, err
			}

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
