package hooks

import (
	"context"
	"errors"

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
	"github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaEntitlementListeners registers entitlement mutation listeners on Gala.
func RegisterGalaEntitlementListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return eventqueue.RegisterMutationListeners(g,
		eventqueue.MutationListener{
			Schema:     entgen.TypeOrganization,
			Name:       "entitlements.organization_created",
			Operations: []string{ent.OpCreate.String()},
			Handle:     handleOrganizationCreatedGala,
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeOrganization,
			Name:   "entitlements.organization_deleted",
			Operations: []string{
				ent.OpDelete.String(),
				ent.OpDeleteOne.String(),
				eventqueue.SoftDeleteOne,
			},
			Handle: handleOrganizationDeleteGala,
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

// handleOrganizationDeleteGala deactivates an organization's customer subscription when deleted.
func handleOrganizationDeleteGala(inv eventqueue.Invocation, _ eventqueue.MutationGalaPayload) error {
	entInv, ok := newEntitlementInvocation(inv, softDeleteAllowContext)
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
func handleOrganizationCreatedGala(inv eventqueue.Invocation, _ eventqueue.MutationGalaPayload) error {
	entInv, ok := newEntitlementInvocation(inv, orgAllowContext)
	if !ok {
		return nil
	}

	return entInv.reconcile()
}

// handleOrganizationSettingsUpdateOneGala updates Stripe customer details for billing changes.
func handleOrganizationSettingsUpdateOneGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	entInv, ok := newEntitlementInvocation(inv, orgAllowContext)
	if !ok {
		return nil
	}

	orgSetting, err := entInv.client.OrganizationSetting.Get(entInv.Allow(), inv.EntityID)
	if err != nil {
		entInv.Logger().Error().Err(err).Str("organization_setting_id", inv.EntityID).Msg("failed to resolve organization from organization setting")

		return nil
	}

	entInv.orgID = orgSetting.OrganizationID

	orgCustomer, err := fetchOrganizationCustomer(entInv, orgSetting)
	if err != nil {
		entInv.Logger().Err(err).Str("organization_setting_id", orgSetting.ID).Msg("failed to fetch organization customer")
		return err
	}

	if orgCustomer.StripeCustomerID == "" {
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

// entitlementInvocation bundles all data needed by entitlement listeners.
type entitlementInvocation struct {
	ctx    context.Context
	client *entgen.Client
	orgID  string
	allow  context.Context
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

// newEntitlementInvocation gathers prerequisites for entitlement mutation handling; the
// organization defaults to the mutated entity and setting-scoped handlers override it
// after resolving their setting
func newEntitlementInvocation(inv eventqueue.Invocation, allow func(context.Context) context.Context) (*entitlementInvocation, bool) {
	if inv.Client.EntitlementManager == nil || !inv.Client.EntitlementManager.Config.IsEnabled() {
		return nil, false
	}

	return &entitlementInvocation{
		ctx:    inv.Context,
		client: inv.Client,
		orgID:  inv.EntityID,
		allow:  allow(inv.Context),
	}, true
}

// fetchOrganizationCustomer builds the organization customer view for a billing settings change
func fetchOrganizationCustomer(inv *entitlementInvocation, orgSetting *entgen.OrganizationSetting) (*entitlements.OrganizationCustomer, error) {
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
	rec, err := reconciler.New(
		reconciler.WithDB(inv.client),
		reconciler.WithStripeClient(inv.client.EntitlementManager),
	)
	if err != nil {
		inv.Logger().Err(err).Msg("unable to construct entitlement reconciler")

		return err
	}

	if _, err := rec.Reconcile(inv.Context(), []string{inv.orgID}); err != nil {
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
