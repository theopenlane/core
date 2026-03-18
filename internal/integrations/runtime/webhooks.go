package runtime

import (
	"context"
	"encoding/json"
	"maps"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/do/v2"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// SyncWebhooks ensures the persisted webhook rows match the definition contract for one installation
func (r *Runtime) SyncWebhooks(ctx context.Context, installation *ent.Integration) error {
	if installation == nil {
		return ErrInstallationRequired
	}

	def, ok := r.Registry().Definition(installation.DefinitionID)
	if !ok {
		return registry.ErrDefinitionNotFound
	}

	db := do.MustInvoke[*ent.Client](r.injector)
	existing, err := db.IntegrationWebhook.Query().
		Where(
			integrationwebhook.IntegrationIDEQ(installation.ID),
			integrationwebhook.ExternalEventIDIsNil(),
		).
		Order(
			integrationwebhook.ByUpdatedAt(sql.OrderDesc()),
			integrationwebhook.ByCreatedAt(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return err
	}

	currentWebhooks := make(map[string]struct{}, len(def.Webhooks))
	for _, webhook := range def.Webhooks {
		currentWebhooks[webhook.Name] = struct{}{}
	}

	staleIDs := make([]string, 0, len(existing))
	seenBaseRows := make(map[string]struct{}, len(existing))
	for _, webhook := range existing {
		baseKey := webhook.Provider + "\x00" + webhook.Name
		if _, ok := seenBaseRows[baseKey]; ok {
			staleIDs = append(staleIDs, webhook.ID)
			continue
		}

		if webhook.Provider != installation.DefinitionSlug {
			staleIDs = append(staleIDs, webhook.ID)
			continue
		}

		if _, ok := currentWebhooks[webhook.Name]; !ok {
			staleIDs = append(staleIDs, webhook.ID)
			continue
		}

		seenBaseRows[baseKey] = struct{}{}
	}

	if len(staleIDs) > 0 {
		if _, err := db.IntegrationWebhook.Delete().
			Where(integrationwebhook.IDIn(staleIDs...)).
			Exec(ctx); err != nil {
			return err
		}
	}

	for _, webhook := range def.Webhooks {
		if _, err := r.ensureWebhook(ctx, installation, webhook); err != nil {
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

	createErr := do.MustInvoke[*ent.Client](r.injector).IntegrationWebhook.Create().
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

	update := do.MustInvoke[*ent.Client](r.injector).IntegrationWebhook.UpdateOneID(webhook.ID).
		SetLastDeliveryAt(time.Now().UTC()).
		SetLastDeliveryStatus(status).
		SetStatus(enums.IntegrationWebhookStatusActive)

	if deliveryID != "" {
		update.SetLastDeliveryID(deliveryID)
	}

	if deliveryErr != nil {
		update.SetStatus(enums.IntegrationWebhookStatusFailed)
		update.SetLastDeliveryError(deliveryErr.Error())
	} else {
		update.ClearLastDeliveryError()
	}

	return update.Exec(ctx)
}

// EnsureWebhook returns the persisted webhook row for one installation and definition webhook
func (r *Runtime) EnsureWebhook(ctx context.Context, installation *ent.Integration, webhookName string) (*ent.IntegrationWebhook, error) {
	if installation == nil {
		return nil, ErrInstallationRequired
	}

	def, ok := r.Registry().Definition(installation.DefinitionID)
	if !ok {
		return nil, registry.ErrDefinitionNotFound
	}

	for _, webhook := range def.Webhooks {
		if webhook.Name == webhookName {
			return r.ensureWebhook(ctx, installation, webhook)
		}
	}

	return nil, registry.ErrWebhookNotFound
}

// DispatchWebhookEvent emits one normalized integration webhook event through Gala
func (r *Runtime) DispatchWebhookEvent(ctx context.Context, installation *ent.Integration, webhookName string, event types.WebhookReceivedEvent) error {
	if installation == nil {
		return ErrInstallationRequired
	}

	registration, err := r.Registry().WebhookEvent(installation.DefinitionID, webhookName, event.Name)
	if err != nil {
		return err
	}

	receipt := do.MustInvoke[*gala.Gala](r.injector).EmitWithHeaders(ctx, registration.Topic, operations.WebhookEnvelope{
		IntegrationID: installation.ID,
		DefinitionID:  installation.DefinitionID,
		Webhook:       webhookName,
		Event:         event.Name,
		DeliveryID:    event.DeliveryID,
		Payload:       jsonx.CloneRawMessage(event.Payload),
		Headers:       maps.Clone(event.Headers),
	}, gala.Headers{
		IdempotencyKey: event.DeliveryID,
		Properties: map[string]string{
			"integration_id": installation.ID,
			"definition_id":  installation.DefinitionID,
			"webhook":        webhookName,
			"event":          event.Name,
		},
	})

	return receipt.Err
}

// HandleWebhookEvent processes one emitted integration webhook envelope
func (r *Runtime) HandleWebhookEvent(ctx context.Context, envelope operations.WebhookEnvelope) error {
	installation, err := r.ResolveInstallation(ctx, "", envelope.IntegrationID, envelope.DefinitionID)
	if err != nil {
		return err
	}

	webhook, err := r.EnsureWebhook(ctx, installation, envelope.Webhook)
	if err != nil {
		return err
	}

	registration, err := r.Registry().WebhookEvent(installation.DefinitionID, envelope.Webhook, envelope.Event)
	if err != nil {
		return err
	}

	return registration.Handle(ctx, types.WebhookHandleRequest{
		Integration: installation,
		Webhook:     webhook,
		DB:          do.MustInvoke[*ent.Client](r.injector),
		Event: types.WebhookReceivedEvent{
			Name:       envelope.Event,
			DeliveryID: envelope.DeliveryID,
			Payload:    jsonx.CloneRawMessage(envelope.Payload),
			Headers:    maps.Clone(envelope.Headers),
		},
		Ingest: func(ingestCtx context.Context, payloadSets []types.IngestPayloadSet) error {
			return operations.EmitPayloadSets(
				ingestCtx,
				operations.IngestContext{
					Registry:     r.Registry(),
					DB:           do.MustInvoke[*ent.Client](r.injector),
					Runtime:      do.MustInvoke[*gala.Gala](r.injector),
					Installation: installation,
				},
				envelope.Webhook,
				registration.Ingest,
				payloadSets,
				operations.IngestOptions{
					Source:       integrationgenerated.IntegrationIngestSourceWebhook,
					Webhook:      envelope.Webhook,
					WebhookEvent: envelope.Event,
					DeliveryID:   envelope.DeliveryID,
				},
			)
		},
		DispatchOperation: func(dispatchCtx context.Context, operation string, config json.RawMessage) error {
			_, dispatchErr := r.Dispatch(dispatchCtx, operations.DispatchRequest{
				InstallationID: installation.ID,
				Operation:      operation,
				Config:         jsonx.CloneRawMessage(config),
				RunType:        enums.IntegrationRunTypeEvent,
				Force:          true,
			})

			return dispatchErr
		},
	})
}

// ensureWebhook creates or updates the persisted webhook row for one installation and webhook registration
func (r *Runtime) ensureWebhook(ctx context.Context, installation *ent.Integration, registration types.WebhookRegistration) (*ent.IntegrationWebhook, error) {
	allowedEvents := lo.Map(registration.Events, func(event types.WebhookEventRegistration, _ int) string {
		return event.Name
	})

	status := enums.IntegrationWebhookStatusPending
	if installation.Status == enums.IntegrationStatusConnected {
		status = enums.IntegrationWebhookStatusActive
	}

	db := do.MustInvoke[*ent.Client](r.injector)
	existing, err := db.IntegrationWebhook.Query().
		Where(
			integrationwebhook.IntegrationIDEQ(installation.ID),
			integrationwebhook.ProviderEQ(installation.DefinitionSlug),
			integrationwebhook.NameEQ(registration.Name),
			integrationwebhook.ExternalEventIDIsNil(),
		).
		Order(
			integrationwebhook.ByUpdatedAt(sql.OrderDesc()),
			integrationwebhook.ByCreatedAt(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(existing) == 0 {
		return db.IntegrationWebhook.Create().
			SetOwnerID(installation.OwnerID).
			SetIntegrationID(installation.ID).
			SetProvider(installation.DefinitionSlug).
			SetName(registration.Name).
			SetAllowedEvents(allowedEvents).
			SetStatus(status).
			Save(ctx)
	}

	if len(existing) > 1 {
		duplicateIDs := lo.Map(existing[1:], func(record *ent.IntegrationWebhook, _ int) string {
			return record.ID
		})
		if _, err := db.IntegrationWebhook.Delete().
			Where(integrationwebhook.IDIn(duplicateIDs...)).
			Exec(ctx); err != nil {
			return nil, err
		}
	}

	return db.IntegrationWebhook.UpdateOneID(existing[0].ID).
		SetAllowedEvents(allowedEvents).
		SetStatus(status).
		Save(ctx)
}
