package hooks

import (
	"context"
	"slices"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// HookOrgSubscriptionCreatePolicy is used on orgsubscription creation mutations
// if the features are set, it will create a conditional tuple that restricts access
// to the organization based on the listed features
func HookOrgSubscriptionCreatePolicy() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// setup vars before switch
			orgSubsID := ""

			allowedFeatures := []string{}

			switch m := m.(type) {
			case *generated.OrgSubscriptionMutation:
				allowedFeatures, _ = m.FeatureLookupKeys()

				orgSubscriptionID, ok := m.ID()
				if !ok || orgSubscriptionID == "" {
					return retVal, nil
				}

				orgsubs, err := m.Client().OrgSubscription.Query().
					Where(orgsubscription.ID(orgSubscriptionID)).
					Select("feature_lookup_keys").Only(ctx)
				if err != nil {
					return nil, err
				}

				allowedFeatures = orgsubs.FeatureLookupKeys
			}

			if allowedFeatures == nil {
				allowedFeatures = []string{}
			}

			if err := updateOrgSubscriptionConditionalTuples(ctx, m, orgSubsID, allowedFeatures); err != nil {
				return nil, err
			}

			return retVal, nil
		})
	},
		hook.HasOp(ent.OpCreate),
	)
}

// HookOrgSubscriptionUpdatePolicy is used on organization subscription tting mutations where the features are set in the request
// it will update the conditional tuple that restricts access to the organization based on the feature list
func HookOrgSubscriptionUpdatePolicy() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.OrgSubscriptionFunc(func(ctx context.Context, m *generated.OrgSubscriptionMutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			orgSubsID, _ := m.ID()

			features, okSet := m.FeatureLookupKeys()
			okClear := m.FeatureLookupKeysCleared()
			appendedFeatures, okAppend := m.AppendedFeatureLookupKeys()

			var featureUpdates []string

			switch {
			case okSet:
				featureUpdates = features
			case okClear:
				featureUpdates = []string{}
			case okAppend:
				originalFeatures, err := m.OldFeatureLookupKeys(ctx)
				if err != nil {
					return nil, err
				}

				featureUpdates = slices.Concat(originalFeatures, appendedFeatures)
			default:
				// we shouldn't get here because the hook is only called when the features are set
				return retVal, nil
			}

			// update the conditional tuples with the new set of features
			if err := updateOrgSubscriptionConditionalTuples(ctx, m, orgSubsID, featureUpdates); err != nil {
				return nil, err
			}

			return retVal, nil
		})
	},
		hook.And(
			hook.Or(
				hook.HasFields("feature_lookup_keys"),
				hook.HasAddedFields("feature_lookup_keys"),
				hook.HasClearedFields("feature_lookup_keys"),
			),
			hook.HasOp(ent.OpUpdateOne|ent.OpUpdate),
		),
	)
}

// updateOrgSubscriptionConditionalTuples will update (or create) a conditional tuple for the organization
// that restricts access based on the feature list in the subscription
// the tuple will look like the following, where the enabled_features are the features that are allowed
// if the list is empty, then all features are allowed
//
// user: organization:openlane
// relation: subscriber
// object: plan:starter

// TODO(MKA): move to iam/fgax?
const SubscriberMemberRelation = "subscriber"

func updateOrgSubscriptionConditionalTuples(ctx context.Context, m ent.Mutation, orgSubsID string, features []string) error {
	tk := fgax.TupleRequest{
		ObjectID:         orgSubsID,
		ObjectType:       generated.TypeOrgSubscription,
		SubjectID:        orgSubsID,
		SubjectType:      generated.TypeOrgSubscription,
		SubjectRelation:  SubscriberMemberRelation,
		Relation:         utils.OrgAccessCheckRelation,
		ConditionName:    utils.OrgSubsFeaturesConditionName,
		ConditionContext: utils.NewOrgSubscriptionConditionContext(features),
	}

	if _, err := utils.AuthzClient(ctx, m).UpdateConditionalTupleKey(ctx, fgax.GetTupleKey(tk)); err != nil {
		log.Error().Err(err).Msg("failed to create org access restriction tuple")

		return err
	}

	return nil
}
