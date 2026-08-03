package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	sync "github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaEntitlementListeners registers entitlement mutation listeners on Gala.
func RegisterGalaEntitlementListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return eventqueue.RegisterMutationListeners(g,
		eventqueue.MutationListener{
			Schema: entgen.TypeOrganization,
			Name:   "entitlements.organization",
			Operations: []string{
				ent.OpCreate.String(),
				ent.OpDelete.String(),
				ent.OpDeleteOne.String(),
				eventqueue.SoftDeleteOne,
			},
			Handle: handleOrganizationMutationGala,
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeOrganizationSetting,
			Name:   "entitlements.organization_setting",
			Operations: []string{
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Fields: []string{"billing_email", "billing_phone", "billing_address"},
			Handle: handleOrganizationSettingsUpdateOneGala,
		},
	)
}

// handleOrganizationMutationGala routes organization mutations to entitlement handlers.
func handleOrganizationMutationGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	switch strings.TrimSpace(payload.Operation) {
	case ent.OpCreate.String():
		return handleOrganizationCreatedGala(inv, payload)
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), eventqueue.SoftDeleteOne:
		return handleOrganizationSubscriptionDeactivationGala(inv, payload)
	default:
		return nil
	}
}

// handleOrganizationSubscriptionDeactivationGala deactivates an organization's customer subscription when deleted
func handleOrganizationSubscriptionDeactivationGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	entInv, ok := newEntitlementInvocation(inv, payload, softDeleteAllowContext)
	if !ok {
		return nil
	}

	org, err := entInv.client.Organization.Query().Where(
		organization.And(
			organization.ID(entInv.orgID),
			organization.DeletedAtNotNil(),
		),
	).Only(entInv.Allow())
	if err != nil {
		entInv.Logger().Err(err).Str("organization_id", entInv.orgID).Msg("organization delete event unable to load organization")
		return nil
	}

	if org.StripeCustomerID == nil {
		return nil
	}

	if err := entInv.client.EntitlementManager.FindAndDeactivateCustomerSubscription(entInv.Context(), *org.StripeCustomerID); err != nil {
		entInv.Logger().Error().Err(err).Msg("failed to deactivate customer subscription")
		return err
	}

	return nil
}

// handleOrganizationCreatedGala reconciles entitlements after organization creation.
func handleOrganizationCreatedGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	entInv, ok := newEntitlementInvocation(inv, payload, orgAllowContext)
	if !ok {
		return nil
	}

	return entInv.reconcile()
}

// handleOrganizationSettingsUpdateOneGala updates Stripe customer details for billing changes.
func handleOrganizationSettingsUpdateOneGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	entInv, ok := newEntitlementInvocation(inv, payload, orgAllowContext)
	if !ok {
		return nil
	}

	orgSettingID := entInv.entityID

	orgCustomer, err := fetchOrganizationCustomerByOrgSettingID(entInv, orgSettingID)
	if err != nil {
		entInv.Logger().Err(err).Str("organization_setting_id", orgSettingID).Msg("failed to fetch organization customer")
		return err
	}

	if orgCustomer == nil || orgCustomer.StripeCustomerID == "" {
		return entInv.reconcile()
	}

	params := entitlements.GetUpdatedFields(payload.ProposedChanges, orgCustomer)

	if params != nil {
		if _, err := entInv.client.EntitlementManager.UpdateCustomer(entInv.Context(), orgCustomer.StripeCustomerID, params); err != nil {
			entInv.Logger().Err(err).Str("stripe_customer_id", orgCustomer.StripeCustomerID).Msg("failed to update stripe customer metadata")
			return err
		}
	}

	return entInv.reconcile()
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
func newEntitlementInvocation(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload, allow func(context.Context) context.Context) (*entitlementInvocation, bool) {
	if inv.Client.EntitlementManager == nil || !inv.Client.EntitlementManager.Config.IsEnabled() {
		return nil, false
	}

	if allow == nil {
		allow = orgAllowContext
	}

	allowCtx := allow(inv.Context)

	orgID := inv.EntityID

	if strings.TrimSpace(payload.MutationType) == entgen.TypeOrganizationSetting {
		setting, err := inv.Client.OrganizationSetting.Get(allowCtx, inv.EntityID)
		if err != nil {
			logx.FromContext(inv.Context).Error().Err(err).Str("organization_setting_id", inv.EntityID).Msg("failed to resolve organization from organization setting")

			return nil, false
		}

		orgID = setting.OrganizationID
	}

	return &entitlementInvocation{
		ctx:      inv.Context,
		client:   inv.Client,
		orgID:    orgID,
		entityID: inv.EntityID,
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
	if inv == nil || inv.client == nil || inv.client.EntitlementManager == nil || !inv.client.EntitlementManager.Config.IsEnabled() {
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
		unwrapped := errors.Unwrap(err)
		// if this is a constraint error, log as warning - this is common in tests and we will still have the logs
		// in production
		if unwrapped != nil && entgen.IsConstraintError(unwrapped) {
			inv.Logger().Warn().Err(err).Msgf("entitlement reconciliation failed, organization with stripe customer id already exits")

			// do not retry a constraint error
			return nil
		}

		inv.Logger().Err(err).Msg("entitlement reconciliation failed")

		return err
	}

	inv.Logger().Debug().Msg("entitlement reconciliation completed")

	return nil
}
