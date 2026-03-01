package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	sync "github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
)

// RegisterGalaEntitlementListeners registers entitlement mutation listeners on Gala.
func RegisterGalaEntitlementListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeOrganization),
			Name:  "entitlements.organization",
			Operations: []string{
				ent.OpCreate.String(),
				ent.OpDelete.String(),
				ent.OpDeleteOne.String(),
				eventqueue.SoftDeleteOne,
			},
			Handle: handleOrganizationMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeOrganizationSetting),
			Name:  "entitlements.organization_setting",
			Operations: []string{
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Handle: handleOrganizationSettingMutationGala,
		},
	)
}

// handleOrganizationMutationGala routes organization mutations to entitlement handlers.
func handleOrganizationMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	switch strings.TrimSpace(payload.Operation) {
	case ent.OpCreate.String():
		return handleOrganizationCreatedGala(ctx, payload)
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), eventqueue.SoftDeleteOne:
		return handleOrganizationDeleteGala(ctx, payload)
	default:
		return nil
	}
}

// handleOrganizationSettingMutationGala handles billing updates on organization settings.
func handleOrganizationSettingMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	switch strings.TrimSpace(payload.Operation) {
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return handleOrganizationSettingsUpdateOneGala(ctx, payload)
	default:
		return nil
	}
}

// handleOrganizationDeleteGala deactivates an organization's customer subscription when deleted.
func handleOrganizationDeleteGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
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

// handleOrganizationCreatedGala reconciles entitlements after organization creation.
func handleOrganizationCreatedGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	inv, ok := newEntitlementInvocation(ctx, payload, orgAllowContext)
	if !ok {
		return nil
	}

	return inv.reconcile()
}

// handleOrganizationSettingsUpdateOneGala updates Stripe customer details for billing changes.
func handleOrganizationSettingsUpdateOneGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if !lo.SomeBy([]string{"billing_email", "billing_phone", "billing_address"}, func(field string) bool {
		return eventqueue.MutationFieldChanged(payload, field)
	}) {
		return nil
	}

	inv, ok := newEntitlementInvocation(ctx, payload, orgAllowContext)
	if !ok {
		return nil
	}

	orgSettingID := inv.entityID
	if orgSettingID == "" {
		if id, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties); ok {
			orgSettingID = id
		}
	}

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

	params := entitlements.GetUpdatedFields(payload.ProposedChanges, orgCustomer)

	if params != nil {
		if _, err := inv.client.EntitlementManager.UpdateCustomer(inv.Context(), orgCustomer.StripeCustomerID, params); err != nil {
			inv.Logger().Err(err).Str("stripe_customer_id", orgCustomer.StripeCustomerID).Msg("failed to update stripe customer metadata")
			return err
		}
	}

	return inv.reconcile()
}

var errMissingOrgCustomerPrereqs = errors.New("entitlement invocation missing prerequisites")

// entitlementInvocation bundles all data needed by entitlement listeners.
type entitlementInvocation struct {
	ctx      context.Context
	client   *entgen.Client
	orgID    string
	entityID string
	allow    context.Context
}

// Context returns the listener context associated with the invocation.
func (inv *entitlementInvocation) Context() context.Context {
	return inv.ctx
}

// Logger returns a contextual logger for the invocation.
func (inv *entitlementInvocation) Logger() *zerolog.Logger {
	return logx.FromContext(inv.Context())
}

// Allow returns the elevated context used for entitlement operations.
func (inv *entitlementInvocation) Allow() context.Context {
	return inv.allow
}

// orgAllowContext bypasses privacy rules for entitlement logic against an organization.
func orgAllowContext(ctx context.Context) context.Context {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	return auth.WithCaller(allowCtx, auth.NewWebhookCaller(""))
}

// softDeleteAllowContext extends orgAllowContext to include soft-deleted records.
func softDeleteAllowContext(ctx context.Context) context.Context {
	ctx = orgAllowContext(ctx)

	return entx.SkipSoftDelete(ctx)
}

// newEntitlementInvocation gathers prerequisites for entitlement mutation handling.
func newEntitlementInvocation(handlerCtx gala.HandlerContext, payload eventqueue.MutationGalaPayload, allow func(context.Context) context.Context) (*entitlementInvocation, bool) {
	handlerCtx, client, ok := eventqueue.ClientFromHandler(handlerCtx)
	if !ok || client.EntitlementManager == nil {
		return nil, false
	}

	if allow == nil {
		allow = orgAllowContext
	}

	allowCtx := allow(handlerCtx.Context)

	entityID, ok := eventqueue.MutationEntityID(payload, handlerCtx.Envelope.Headers.Properties)
	if !ok {
		return nil, false
	}

	orgID := entityID

	if strings.TrimSpace(payload.MutationType) == entgen.TypeOrganizationSetting {
		setting, err := client.OrganizationSetting.Get(allowCtx, entityID)
		if err != nil {
			logx.FromContext(handlerCtx.Context).Error().Err(err).Str("organization_setting_id", entityID).Msg("failed to resolve organization from organization setting")

			return nil, false
		}

		orgID = setting.OrganizationID
	}

	return &entitlementInvocation{
		ctx:      handlerCtx.Context,
		client:   client,
		orgID:    orgID,
		entityID: entityID,
		allow:    allowCtx,
	}, true
}

// fetchOrganizationCustomerByOrgSettingID loads organization and customer data for a setting.
func fetchOrganizationCustomerByOrgSettingID(inv *entitlementInvocation, orgSettingID string) (*entitlements.OrganizationCustomer, error) {
	if inv == nil || inv.client == nil || orgSettingID == "" {
		return nil, fmt.Errorf("%w: organization_setting_id=%s", errMissingOrgCustomerPrereqs, orgSettingID)
	}

	orgSetting, err := inv.client.OrganizationSetting.Get(inv.Allow(), orgSettingID)
	if err != nil {
		return nil, err
	}

	org, err := inv.client.Organization.Query().
		Where(organization.ID(orgSetting.OrganizationID)).
		Only(inv.Allow())
	if err != nil {
		return nil, err
	}

	stripeCustomerID := ""
	if org.StripeCustomerID != nil {
		stripeCustomerID = *org.StripeCustomerID
	}

	contact := entitlements.ContactInfo{
		Email: orgSetting.BillingEmail,
		Phone: orgSetting.BillingPhone,
	}

	if address := orgSetting.BillingAddress; address != (models.Address{}) {
		contact.Line1 = lo.ToPtr(address.Line1)
		contact.Line2 = lo.ToPtr(address.Line2)
		contact.City = lo.ToPtr(address.City)
		contact.State = lo.ToPtr(address.State)
		contact.Country = lo.ToPtr(address.Country)
		contact.PostalCode = lo.ToPtr(address.PostalCode)
	}

	return &entitlements.OrganizationCustomer{
		OrganizationID:         org.ID,
		OrganizationName:       org.Name,
		OrganizationSettingsID: orgSetting.ID,
		StripeCustomerID:       stripeCustomerID,
		ContactInfo:            contact,
	}, nil
}

// reconcile runs entitlement reconciliation for the invocation's organization.
func (inv *entitlementInvocation) reconcile() error {
	if inv == nil || inv.client == nil || inv.client.EntitlementManager == nil {
		return nil
	}

	reconciler, err := sync.New(
		sync.WithDB(inv.client),
		sync.WithStripeClient(inv.client.EntitlementManager),
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
