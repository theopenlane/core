package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/rs/zerolog"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	entreconciler "github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
)

type entitlementInvocation struct {
	event   *soiree.EventContext
	payload *MutationPayload
	client  *entgen.Client
	orgID   string
	allow   context.Context
}

func (inv *entitlementInvocation) Context() context.Context {
	return inv.event.Context()
}

func (inv *entitlementInvocation) Logger() *zerolog.Logger {
	return zerolog.Ctx(inv.Context())
}

func (inv *entitlementInvocation) Allow() context.Context {
	return inv.allow
}

func orgAllowContext(ctx context.Context) context.Context {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	return contextx.With(allowCtx, auth.OrgSubscriptionContextKey{})
}

func softDeleteAllowContext(ctx context.Context) context.Context {
	ctx = orgAllowContext(ctx)
	return context.WithValue(ctx, entx.SoftDeleteSkipKey{}, true)
}

func newEntitlementInvocation(event *soiree.EventContext, payload *MutationPayload, allow func(context.Context) context.Context) (*entitlementInvocation, bool) {
	client := mutationClient(event, payload)
	if client == nil || client.EntitlementManager == nil {
		return nil, false
	}

	orgID, ok := mutationOrganizationID(event, payload)
	if !ok {
		return nil, false
	}

	if allow == nil {
		allow = orgAllowContext
	}

	return &entitlementInvocation{
		event:   event,
		payload: payload,
		client:  client,
		orgID:   orgID,
		allow:   allow(event.Context()),
	}, true
}

func mutationOrganizationID(ctx *soiree.EventContext, payload *MutationPayload) (string, bool) {
	if payload != nil && payload.EntityID != "" {
		return payload.EntityID, true
	}

	if ctx == nil {
		return "", false
	}

	if id, ok := ctx.Properties().String("ID"); ok && id != "" {
		return id, true
	}

	if raw, ok := ctx.Properties().Get("ID"); ok && raw != nil {
		if str, ok := raw.(fmt.Stringer); ok {
			value := str.String()
			if value == "" {
				return "", false
			}

			return value, true
		}

		value := fmt.Sprint(raw)
		if value == "" || value == "<nil>" {
			return "", false
		}

		return value, true
	}

	return "", false
}

func mutationClient(ctx *soiree.EventContext, payload *MutationPayload) *entgen.Client {
	if payload != nil && payload.Client != nil {
		return payload.Client
	}

	client, ok := soiree.ClientAs[*entgen.Client](ctx)
	if !ok {
		return nil
	}

	return client
}

func mutationTouches(m ent.Mutation, fields ...string) bool {
	if m == nil {
		return false
	}

	for _, field := range fields {
		if _, ok := m.Field(field); ok {
			return true
		}
	}

	return false
}

func handleOrganizationMutation(ctx *soiree.EventContext, payload *MutationPayload) error {
	if payload == nil {
		return nil
	}

	switch payload.Operation {
	case ent.OpCreate.String():
		return handleOrganizationCreated(ctx, payload)
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), SoftDeleteOne:
		return handleOrganizationDelete(ctx, payload)
	default:
		return nil
	}
}

func (inv *entitlementInvocation) reconcile() error {
	if inv == nil || inv.client == nil || inv.client.EntitlementManager == nil {
		return nil
	}

	reconciler, err := entreconciler.New(
		entreconciler.WithDB(inv.client),
		entreconciler.WithStripeClient(inv.client.EntitlementManager),
	)
	if err != nil {
		inv.Logger().Err(err).Msg("unable to construct entitlement reconciler")
		return err
	}

	if _, err := reconciler.Reconcile(inv.Context(), []string{inv.orgID}); err != nil {
		inv.Logger().Err(err).Msg("entitlement reconciliation failed")
		return err
	}

	inv.Logger().Debug().Msg("entitlement reconciliation completed")

	return nil
}

func handleOrganizationSettingMutation(ctx *soiree.EventContext, payload *MutationPayload) error {
	if payload == nil {
		return nil
	}

	switch payload.Operation {
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return handleOrganizationSettingsUpdateOne(ctx, payload)
	default:
		return nil
	}
}

func handleOrganizationDelete(ctx *soiree.EventContext, payload *MutationPayload) error {
	inv, ok := newEntitlementInvocation(ctx, payload, softDeleteAllowContext)
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

func handleOrganizationCreated(ctx *soiree.EventContext, payload *MutationPayload) error {
	inv, ok := newEntitlementInvocation(ctx, payload, orgAllowContext)
	if !ok {
		return nil
	}

	return inv.reconcile()
}

func handleOrganizationSettingsUpdateOne(ctx *soiree.EventContext, payload *MutationPayload) error {
	if !mutationTouches(payload.Mutation, "billing_email", "billing_phone", "billing_address") {
		return nil
	}

	inv, ok := newEntitlementInvocation(ctx, payload, orgAllowContext)
	if !ok {
		return nil
	}

	return inv.reconcile()
}
