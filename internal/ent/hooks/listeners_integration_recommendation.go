package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/samber/do/v2"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/asset"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/integrationrecommendation"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/integrationrecommendations"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/gala"
)

type integrationRecommendationRecomputePayload struct {
	OwnerID string `json:"owner_id"`
}

func RegisterGalaIntegrationRecommendationListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	triggerName := "integration_recommendation.trigger"

	ids, err := gala.RegisterListeners(registry, gala.Definition[integrationRecommendationRecomputePayload]{
		Topic:  integrationRecommendationRecomputeTopic(),
		Name:   "integration_recommendation.recompute",
		Handle: handleIntegrationRecommendationRebuild,
	})
	if err != nil {
		return nil, err
	}

	recomputationIDs, err := gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeAsset),
			Name:       triggerName + ".asset",
			Operations: []string{ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String(), eventqueue.SoftDeleteOne},
			Handle:     handleIntegrationRecommendationTrigger,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeEntity),
			Name:       triggerName + ".entity",
			Operations: []string{ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String(), eventqueue.SoftDeleteOne},
			Handle:     handleIntegrationRecommendationTrigger,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeOrganizationSetting),
			Name:       triggerName + ".organization_setting",
			Operations: []string{ent.OpUpdate.String(), ent.OpUpdateOne.String()},
			Handle:     handleIntegrationRecommendationTrigger,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeOrgMembership),
			Name:       triggerName + ".org_membership",
			Operations: []string{ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String(), eventqueue.SoftDeleteOne},
			Handle:     handleIntegrationRecommendationTrigger,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeUser),
			Name:       triggerName + ".user",
			Operations: []string{ent.OpUpdate.String(), ent.OpUpdateOne.String()},
			Handle:     handleIntegrationRecommendationTrigger,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeIntegration),
			Name:       triggerName + ".integration",
			Operations: []string{ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String(), eventqueue.SoftDeleteOne},
			Handle:     handleIntegrationRecommendationTrigger,
		},
	)
	if err != nil {
		return nil, err
	}

	return append(ids, recomputationIDs...), nil
}

func integrationRecommendationRecomputeTopic() gala.Topic[integrationRecommendationRecomputePayload] {
	return gala.Topic[integrationRecommendationRecomputePayload]{Name: gala.TopicName("integration_recommendation.recompute")}
}

func handleIntegrationRecommendationTrigger(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	switch payload.MutationType {
	case generated.TypeOrganizationSetting:
		if !eventqueue.MutationFieldChanged(payload, organizationsetting.FieldIdentityProvider) {
			return nil
		}

	case generated.TypeUser:
		if !eventqueue.MutationFieldChanged(payload, user.FieldAuthProvider) &&
			!eventqueue.MutationFieldChanged(payload, user.FieldLastLoginProvider) {
			return nil
		}
	}

	ownerIDs, err := integrationRecommendationOwnerIDs(ctx, client, payload)
	if err != nil {
		return err
	}

	for _, ownerID := range ownerIDs {
		if err := enqueueIntegrationRecommendationRecompute(ctx, ownerID); err != nil {
			return err
		}
	}

	return nil
}

func handleIntegrationRecommendationRebuild(ctx gala.HandlerContext, payload integrationRecommendationRecomputePayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	ownerID := strings.TrimSpace(payload.OwnerID)
	if ownerID == "" {
		return nil
	}

	rt := intruntime.FromClient(ctx.Context, client)
	if rt == nil {
		return nil
	}

	rebuildCtx := rule.WithInternalContext(ctx.Context)

	recommendations, err := integrationrecommendations.Compute(rebuildCtx, client, rt.Catalog(), ownerID)
	if err != nil {
		return err
	}

	return storeIntegrationRecommendations(rebuildCtx, client, ownerID, recommendations)
}

func enqueueIntegrationRecommendationRecompute(ctx gala.HandlerContext, ownerID string) error {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil
	}

	galaApp, err := do.Invoke[*gala.Gala](ctx.Injector)
	if err != nil || galaApp == nil {
		return nil
	}

	topic := integrationRecommendationRecomputeTopic()

	exists, err := galaApp.HasActiveJobForTopic(ctx.Context, topic.Name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	receipt := galaApp.EmitWithHeaders(ctx.Context, topic.Name, integrationRecommendationRecomputePayload{
		OwnerID: ownerID,
	}, gala.Headers{
		IdempotencyKey: "integration_recommendation.recompute." + ownerID,
		Properties: map[string]string{
			"owner_id": ownerID,
		},
		Tags: []string{
			"integration_recommendation",
		},
	})

	return receipt.Err
}

func integrationRecommendationOwnerIDs(ctx gala.HandlerContext, client *generated.Client, payload eventqueue.MutationGalaPayload) ([]string, error) {
	entityID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || entityID == "" {
		return nil, nil
	}

	queryCtx := rule.WithInternalContext(ctx.Context)

	switch payload.MutationType {
	case generated.TypeAsset:
		record, err := client.Asset.Query().Where(asset.IDEQ(entityID)).Only(queryCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return nil, nil
			}

			return nil, err
		}

		return []string{record.OwnerID}, nil

	case generated.TypeEntity:
		record, err := client.Entity.Query().Where(entity.IDEQ(entityID)).Only(queryCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return nil, nil
			}

			return nil, err
		}

		return []string{record.OwnerID}, nil

	case generated.TypeIntegration:
		record, err := client.Integration.Query().Where(integration.IDEQ(entityID)).Only(queryCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return nil, nil
			}

			return nil, err
		}

		return []string{record.OwnerID}, nil

	case generated.TypeOrganizationSetting:
		record, err := client.OrganizationSetting.Query().Where(organizationsetting.IDEQ(entityID)).Only(queryCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return nil, nil
			}

			return nil, err
		}

		return []string{record.OrganizationID}, nil

	case generated.TypeOrgMembership:
		record, err := client.OrgMembership.Query().Where(orgmembership.IDEQ(entityID)).Only(queryCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return nil, nil
			}

			return nil, err
		}

		return []string{record.OrganizationID}, nil

	case generated.TypeUser:
		memberships, err := client.OrgMembership.Query().
			Where(orgmembership.UserIDEQ(entityID)).
			All(queryCtx)
		if err != nil {
			return nil, err
		}

		ownerIDs := make([]string, 0, len(memberships))
		seen := map[string]struct{}{}
		for _, membership := range memberships {
			if membership.OrganizationID == "" {
				continue
			}

			if _, ok := seen[membership.OrganizationID]; ok {
				continue
			}

			seen[membership.OrganizationID] = struct{}{}
			ownerIDs = append(ownerIDs, membership.OrganizationID)
		}

		return ownerIDs, nil

	default:
		return nil, nil
	}
}

func storeIntegrationRecommendations(ctx context.Context, client *generated.Client, ownerID string, recommendations []integrationrecommendations.Recommendation) error {
	existing, err := client.IntegrationRecommendation.Query().
		Where(integrationrecommendation.OwnerIDEQ(ownerID)).
		All(ctx)
	if err != nil {
		return err
	}

	existingByDefinitionID := make(map[string]*generated.IntegrationRecommendation, len(existing))
	for _, item := range existing {
		existingByDefinitionID[item.DefinitionID] = item
	}

	seen := map[string]struct{}{}
	for _, recommendation := range recommendations {
		seen[recommendation.DefinitionID] = struct{}{}

		item, ok := existingByDefinitionID[recommendation.DefinitionID]
		if !ok {
			if err := client.IntegrationRecommendation.Create().
				SetOwnerID(ownerID).
				SetDefinitionID(recommendation.DefinitionID).
				SetWeight(recommendation.Weight).
				SetLabel(recommendation.Label).
				Exec(ctx); err != nil {
				return err
			}

			continue
		}

		if item.Weight == recommendation.Weight && item.Label == recommendation.Label {
			continue
		}

		if err := client.IntegrationRecommendation.UpdateOneID(item.ID).
			SetWeight(recommendation.Weight).
			SetLabel(recommendation.Label).
			Exec(ctx); err != nil {
			return err
		}
	}

	for _, item := range existing {
		if _, ok := seen[item.DefinitionID]; ok {
			continue
		}

		if err := client.IntegrationRecommendation.DeleteOneID(item.ID).Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}
