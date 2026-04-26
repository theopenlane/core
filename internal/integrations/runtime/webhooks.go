package runtime

import (
	"context"
	"encoding/json"
	"maps"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/utils/keygen"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// reconcileInstallationWebhooks ensures the persisted webhook rows match the definition contract for one integration
func (r *Runtime) reconcileInstallationWebhooks(ctx context.Context, integration *ent.Integration, previousIntegrationID string) error {
	def, err := r.resolveDefinitionForInstallation(integration)
	if err != nil {
		return err
	}

	db := r.DB()
	existing, err := db.IntegrationWebhook.Query().Where(integrationwebhook.IntegrationIDEQ(integration.ID), integrationwebhook.ExternalEventIDIsNil()).All(ctx)
	if err != nil {
		return err
	}

	currentWebhooks := lo.Associate(def.Webhooks, func(w types.WebhookRegistration) (string, struct{}) {
		return w.Name, struct{}{}
	})

	staleIDs := lo.FilterMap(existing, func(row *ent.IntegrationWebhook, _ int) (string, bool) {
		_, current := currentWebhooks[row.Name]
		return row.ID, !current
	})

	if len(staleIDs) > 0 {
		if _, err := db.IntegrationWebhook.Delete().Where(integrationwebhook.IDIn(staleIDs...)).Exec(ctx); err != nil {
			return err
		}
	}

	for _, webhook := range def.Webhooks {
		if _, err := r.ensureWebhook(ctx, integration, webhook, previousIntegrationID); err != nil {
			return err
		}
	}

	return nil
}

// PrepareWebhookDelivery records one delivery idempotency key when present
func (r *Runtime) PrepareWebhookDelivery(ctx context.Context, webhook *ent.IntegrationWebhook, deliveryID string) (bool, error) {
	if deliveryID == "" {
		return false, nil
	}

	createErr := r.DB().IntegrationWebhook.Create().
		SetOwnerID(webhook.OwnerID).
		SetIntegrationID(webhook.IntegrationID).
		SetProvider(webhook.Provider).
		SetName(webhook.Name).
		SetExternalEventID(deliveryID).
		Exec(ctx)
	if createErr != nil {
		if ent.IsConstraintError(createErr) {
			return true, nil
		}

		return false, createErr
	}

	return false, nil
}

// FinalizeWebhookDelivery updates persisted delivery metadata for one webhook endpoint
func (r *Runtime) FinalizeWebhookDelivery(ctx context.Context, webhook *ent.IntegrationWebhook, deliveryID string, status string, deliveryErr error) error {
	if webhook == nil {
		return nil
	}

	update := r.DB().IntegrationWebhook.UpdateOneID(webhook.ID).
		SetLastDeliveryAt(time.Now().UTC()).
		SetLastDeliveryStatus(status)

	if deliveryID != "" {
		update.SetLastDeliveryID(deliveryID)
	}

	if deliveryErr != nil {
		update.SetLastDeliveryError(deliveryErr.Error())
	} else {
		update.ClearLastDeliveryError()
	}

	return update.Exec(ctx)
}

// ResolveWebhookByEndpoint returns the persisted inbound webhook row for the given endpoint ID
func (r *Runtime) ResolveWebhookByEndpoint(ctx context.Context, endpointID string) (*ent.IntegrationWebhook, error) {
	return r.DB().IntegrationWebhook.Query().
		Where(
			integrationwebhook.EndpointIDEQ(endpointID),
			integrationwebhook.ExternalEventIDIsNil(),
		).
		Only(ctx)
}

// EnsureWebhook returns the persisted webhook row for one integration and definition webhook
func (r *Runtime) EnsureWebhook(ctx context.Context, integration *ent.Integration, webhookName string, previousIntegrationID string) (*ent.IntegrationWebhook, error) {
	def, err := r.resolveDefinitionForInstallation(integration)
	if err != nil {
		return nil, err
	}

	webhook, found := lo.Find(def.Webhooks, func(w types.WebhookRegistration) bool {
		return w.Name == webhookName
	})
	if !found {
		return nil, registry.ErrWebhookNotFound
	}

	return r.ensureWebhook(ctx, integration, webhook, previousIntegrationID)
}

// DispatchWebhookEvent emits one normalized integration webhook event through Gala
func (r *Runtime) DispatchWebhookEvent(ctx context.Context, integration *ent.Integration, webhookName string, event types.WebhookReceivedEvent) error {
	if integration == nil {
		return ErrInstallationRequired
	}

	registration, err := r.Registry().WebhookEvent(integration.DefinitionID, webhookName, event.Name)
	if err != nil {
		return err
	}

	metadata := types.ExecutionMetadata{
		OwnerID:       integration.OwnerID,
		IntegrationID: integration.ID,
		DefinitionID:  integration.DefinitionID,
		Webhook:       webhookName,
		Event:         event.Name,
		DeliveryID:    event.DeliveryID,
	}

	tags := types.GetTagsForExecutionMetadata(metadata)
	if integration.Name != "" {
		tags = append(tags, strcase.UpperSnakeCase(integration.Name))
	}

	if integration.Family != "" {
		tags = append(tags, strcase.UpperSnakeCase(integration.Family))
	}

	receipt := r.Gala().EmitWithHeaders(types.WithExecutionMetadata(ctx, metadata), registration.Topic, operations.WebhookEnvelope{
		ExecutionMetadata: metadata,
		Payload:           jsonx.CloneRawMessage(event.Payload),
		Headers:           maps.Clone(event.Headers),
	}, gala.Headers{
		IdempotencyKey: event.DeliveryID,
		Properties:     metadata.Properties(),
		Tags:           tags,
	})

	return receipt.Err
}

// HandleWebhookEvent processes one emitted integration webhook envelope
func (r *Runtime) HandleWebhookEvent(ctx context.Context, envelope operations.WebhookEnvelope) error {
	ctx, integration, metadata, err := r.bootstrapHandlerContext(ctx, envelope.ExecutionMetadata)
	if err != nil {
		return err
	}

	webhook, err := r.EnsureWebhook(ctx, integration, envelope.Webhook, "")
	if err != nil {
		return err
	}

	registration, err := r.Registry().WebhookEvent(integration.DefinitionID, envelope.Webhook, envelope.Event)
	if err != nil {
		return err
	}

	return registration.Handle(ctx, types.WebhookHandleRequest{
		Integration: integration,
		Webhook:     webhook,
		DB:          r.DB(),
		Event: types.WebhookReceivedEvent{
			Name:       envelope.Event,
			DeliveryID: envelope.DeliveryID,
			Payload:    jsonx.CloneRawMessage(envelope.Payload),
			Headers:    maps.Clone(envelope.Headers),
		},
		Ingest: func(ingestCtx context.Context, payloadSets []types.IngestPayloadSet) error {
			return operations.EmitPayloadSets(ingestCtx, operations.IngestContext{
				Registry:    r.Registry(),
				DB:          r.DB(),
				Runtime:     r.Gala(),
				Integration: integration,
			}, envelope.Webhook, registration.Ingest, payloadSets, operations.IngestOptionsFromMetadata(integrationgenerated.IntegrationIngestSourceWebhook, metadata))
		},
		DispatchOperation: func(dispatchCtx context.Context, operation string, config json.RawMessage) error {
			_, dispatchErr := r.Dispatch(dispatchCtx, operations.DispatchRequest{
				IntegrationID: integration.ID,
				Operation:     operation,
				Config:        jsonx.CloneRawMessage(config),
				RunType:       enums.IntegrationRunTypeWebhook,
			})

			return dispatchErr
		},
		CleanupInstallation: func(cleanupCtx context.Context) error {
			return r.cleanupInstallation(privacy.DecisionContext(cleanupCtx, privacy.Allow), integration.ID)
		},
	})
}

// ensureWebhook creates or updates the persisted webhook row for one integration and webhook registration
func (r *Runtime) ensureWebhook(ctx context.Context, intg *ent.Integration, registration types.WebhookRegistration, previousIntegrationID string) (*ent.IntegrationWebhook, error) {
	allowedEvents := lo.Map(registration.Events, func(event types.WebhookEventRegistration, _ int) string {
		return event.Name
	})

	db := r.DB()

	integrationIDs := []string{intg.ID}
	if previousIntegrationID != "" {
		integrationIDs = append(integrationIDs, previousIntegrationID)
	}

	rows, err := db.IntegrationWebhook.Query().
		Where(
			integrationwebhook.IntegrationIDIn(integrationIDs...),
			integrationwebhook.NameEQ(registration.Name),
			integrationwebhook.ExternalEventIDIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		endpointID := keygen.PrefixedSecret("tolwh")
		endpointURL := webhookEndpointURL(registration, endpointID)

		return db.IntegrationWebhook.Create().
			SetOwnerID(intg.OwnerID).
			SetIntegrationID(intg.ID).
			SetProvider(intg.DefinitionID).
			SetName(registration.Name).
			SetAllowedEvents(allowedEvents).
			SetEndpointID(endpointID).
			SetEndpointURL(endpointURL).
			Save(ctx)
	}

	row := rows[0]
	endpointURL := webhookEndpointURL(registration, lo.FromPtr(row.EndpointID))

	// remove duplicates that should not exist
	duplicateIDs := lo.FilterMap(rows, func(candidate *ent.IntegrationWebhook, _ int) (string, bool) {
		return candidate.ID, candidate.ID != row.ID
	})

	if len(duplicateIDs) > 0 {
		if _, err := db.IntegrationWebhook.Delete().
			Where(integrationwebhook.IDIn(duplicateIDs...)).
			Exec(ctx); err != nil {
			return nil, err
		}
	}

	update := db.IntegrationWebhook.UpdateOneID(row.ID).
		SetAllowedEvents(allowedEvents).
		SetEndpointURL(endpointURL)

	if row.IntegrationID != intg.ID {
		update.SetIntegrationID(intg.ID)
	}

	return update.Save(ctx)
}

// webhookEndpointURL is a small constructor that allows us to override if required but otherwise return a consistent pattern for the route
func webhookEndpointURL(registration types.WebhookRegistration, endpointID string) string {
	if registration.EndpointURLTemplate == "" {
		return "/v1/integrations/webhook/" + endpointID
	}

	return strings.ReplaceAll(registration.EndpointURLTemplate, "{endpointID}", endpointID)
}
