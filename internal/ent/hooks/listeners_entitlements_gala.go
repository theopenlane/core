package hooks

import (
	"context"
	"errors"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	galaMutationOrganizationTopic = gala.Topic[eventqueue.MutationGalaPayload]{
		Name: gala.TopicName(entgen.TypeOrganization),
	}
	galaMutationOrganizationSettingTopic = gala.Topic[eventqueue.MutationGalaPayload]{
		Name: gala.TopicName(entgen.TypeOrganizationSetting),
	}
)

// RegisterGalaEntitlementListeners registers initial gala-native entitlement listeners.
func RegisterGalaEntitlementListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	topicRegistrations := []gala.Registration[eventqueue.MutationGalaPayload]{
		{
			Topic: galaMutationOrganizationTopic,
			Codec: gala.JSONCodec[eventqueue.MutationGalaPayload]{},
			Policy: gala.TopicPolicy{
				EmitMode:   gala.EmitModeDurable,
				QueueClass: gala.QueueClassWorkflow,
			},
		},
		{
			Topic: galaMutationOrganizationSettingTopic,
			Codec: gala.JSONCodec[eventqueue.MutationGalaPayload]{},
			Policy: gala.TopicPolicy{
				EmitMode:   gala.EmitModeDurable,
				QueueClass: gala.QueueClassWorkflow,
			},
		},
	}

	for _, topicRegistration := range topicRegistrations {
		err := topicRegistration.Register(registry)
		if err == nil {
			continue
		}

		if errors.Is(err, gala.ErrTopicAlreadyRegistered) {
			continue
		}

		return nil, err
	}

	definitions := []gala.Definition[eventqueue.MutationGalaPayload]{
		{
			Topic:  galaMutationOrganizationTopic,
			Name:   "entitlements.organization",
			Handle: handleOrganizationMutationGala,
		},
		{
			Topic:  galaMutationOrganizationSettingTopic,
			Name:   "entitlements.organization_setting",
			Handle: handleOrganizationSettingMutationGala,
		},
	}

	ids := make([]gala.ListenerID, 0, len(definitions))
	for _, definition := range definitions {
		id, err := definition.Register(registry)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// handleOrganizationMutationGala routes organization mutations to the correct entitlement handler.
func handleOrganizationMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	switch payload.Operation {
	case ent.OpCreate.String():
		return handleOrganizationCreatedGala(ctx, payload)
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), SoftDeleteOne:
		return handleOrganizationDeleteGala(ctx, payload)
	default:
		return nil
	}
}

// handleOrganizationSettingMutationGala handles billing-related organization setting updates.
func handleOrganizationSettingMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	switch payload.Operation {
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return handleOrganizationSettingsUpdateOneGala(ctx, payload)
	default:
		return nil
	}
}

// handleOrganizationCreatedGala reconciles entitlements after organization creation.
func handleOrganizationCreatedGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	inv, ok := newEntitlementInvocationFromGala(ctx, payload, orgAllowContext)
	if !ok {
		return nil
	}

	return inv.reconcile()
}

// handleOrganizationDeleteGala deactivates customer subscription when an organization is deleted.
func handleOrganizationDeleteGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	inv, ok := newEntitlementInvocationFromGala(ctx, payload, softDeleteAllowContext)
	if !ok {
		return nil
	}

	org, err := inv.client.Organization.Query().Where(
		organization.And(
			organization.ID(inv.orgID),
			organization.DeletedAtNotNil(),
		),
	).Only(inv.Allow())
	if err != nil {
		inv.Logger().Err(err).Str("organization_id", inv.orgID).Msg("organization delete event unable to load organization")
		return nil
	}

	if org.StripeCustomerID == nil {
		return nil
	}

	if err := inv.client.EntitlementManager.FindAndDeactivateCustomerSubscription(inv.Context(), *org.StripeCustomerID); err != nil {
		inv.Logger().Error().Err(err).Msg("failed to deactivate customer subscription")
		return err
	}

	return nil
}

// handleOrganizationSettingsUpdateOneGala updates Stripe customer fields and reconciles organization entitlements.
func handleOrganizationSettingsUpdateOneGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if !payloadTouchesFieldsGala(payload, "billing_email", "billing_phone", "billing_address") {
		return nil
	}

	inv, ok := newEntitlementInvocationFromGala(ctx, payload, orgAllowContext)
	if !ok {
		return nil
	}

	orgSettingID := inv.entityID
	if orgSettingID == "" {
		inv.Logger().Warn().Msg("organization settings update missing entity id; skipping stripe update")
		return nil
	}

	orgCustomer, err := fetchOrganizationCustomerByOrgSettingID(inv, orgSettingID)
	if err != nil {
		inv.Logger().Err(err).Str("organization_setting_id", orgSettingID).Msg("failed to fetch organization customer")
		return err
	}

	if orgCustomer == nil || orgCustomer.StripeCustomerID == "" {
		return inv.reconcile()
	}

	params := entitlements.GetUpdatedFields(stringMapToAnyMap(ctx.Envelope.Headers.Properties), orgCustomer)
	if params != nil {
		if _, err := inv.client.EntitlementManager.UpdateCustomer(inv.Context(), orgCustomer.StripeCustomerID, params); err != nil {
			inv.Logger().Err(err).Str("stripe_customer_id", orgCustomer.StripeCustomerID).Msg("failed to update stripe customer metadata")
			return err
		}
	}

	return inv.reconcile()
}

// payloadTouchesFieldsGala reports whether mutation metadata contains any of the requested field names.
func payloadTouchesFieldsGala(payload eventqueue.MutationGalaPayload, fields ...string) bool {
	if len(fields) == 0 {
		return false
	}

	changed := append(append([]string(nil), payload.ChangedFields...), payload.ClearedFields...)
	changedSet := map[string]struct{}{}
	for _, field := range changed {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		changedSet[field] = struct{}{}
	}

	for _, field := range fields {
		if _, ok := changedSet[field]; ok {
			return true
		}
	}

	return false
}

// newEntitlementInvocationFromGala constructs entitlement invocation context from gala handler context.
func newEntitlementInvocationFromGala(
	handlerContext gala.HandlerContext,
	payload eventqueue.MutationGalaPayload,
	allow func(context.Context) context.Context,
) (*entitlementInvocation, bool) {
	client, err := gala.ResolveFromContext[*entgen.Client](handlerContext)
	if err != nil || client == nil || client.EntitlementManager == nil {
		return nil, false
	}

	if allow == nil {
		allow = orgAllowContext
	}

	allowCtx := allow(handlerContext.Context)

	entityID := strings.TrimSpace(payload.EntityID)
	if entityID == "" {
		return nil, false
	}

	orgID := entityID
	if strings.TrimSpace(payload.MutationType) == entgen.TypeOrganizationSetting {
		setting, getErr := client.OrganizationSetting.Get(allowCtx, entityID)
		if getErr != nil {
			logx.FromContext(handlerContext.Context).
				Err(getErr).
				Str("organization_setting_id", entityID).
				Msg("failed to resolve organization from organization setting")
			return nil, false
		}

		orgID = setting.OrganizationID
	}

	invocationPayload := &events.MutationPayload{
		MutationType:    payload.MutationType,
		Operation:       payload.Operation,
		EntityID:        payload.EntityID,
		ChangedFields:   append([]string(nil), payload.ChangedFields...),
		ClearedFields:   append([]string(nil), payload.ClearedFields...),
		ChangedEdges:    append([]string(nil), payload.ChangedEdges...),
		AddedIDs:        events.CloneStringSliceMap(payload.AddedIDs),
		RemovedIDs:      events.CloneStringSliceMap(payload.RemovedIDs),
		ProposedChanges: events.CloneAnyMap(payload.ProposedChanges),
		Client:          client,
	}

	return &entitlementInvocation{
		ctx:      handlerContext.Context,
		payload:  invocationPayload,
		client:   client,
		orgID:    orgID,
		entityID: entityID,
		allow:    allowCtx,
	}, true
}

// stringMapToAnyMap converts string headers map into a map accepted by entitlement helper APIs.
func stringMapToAnyMap(values map[string]string) map[string]any {
	if len(values) == 0 {
		return nil
	}

	converted := make(map[string]any, len(values))
	for key, value := range values {
		if strings.TrimSpace(key) == "" {
			continue
		}

		converted[key] = value
	}

	if len(converted) == 0 {
		return nil
	}

	return converted
}
