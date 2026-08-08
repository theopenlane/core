package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaEntitlementListeners registers entitlement mutation listeners on Gala.
func RegisterGalaEntitlementListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return registerMutationListeners(g,
		entityops.MutationListener{
			Schema:     entgen.TypeOrganization,
			Label:      "entitlements_created",
			Operations: []string{ent.OpCreate.String()},
			Caller: func(*auth.Caller, entityops.MutationPayload) *auth.Caller {
				return auth.NewWebhookCaller("")
			},
			Elevate: orgAllowContext,
			Handle:  handleOrganizationCreatedGala,
		},
		entityops.MutationListener{
			Schema: entgen.TypeOrganization,
			Label:  "entitlements_deleted",
			Operations: []string{
				ent.OpDelete.String(),
				ent.OpDeleteOne.String(),
				gala.SoftDeleteOne,
			},
			Caller: func(*auth.Caller, entityops.MutationPayload) *auth.Caller {
				return auth.NewWebhookCaller("")
			},
			Elevate: softDeleteAllowContext,
			Handle:  handleOrganizationDeleteGala,
		},
		entityops.MutationListener{
			Schema: entgen.TypeOrganizationSetting,
			Label:  "billing",
			Operations: []string{
				ent.OpUpdate.String(),
				ent.OpUpdateOne.String(),
			},
			Fields: []string{"billing_email", "billing_phone", "billing_address"},
			Caller: func(*auth.Caller, entityops.MutationPayload) *auth.Caller {
				return auth.NewWebhookCaller("")
			},
			Elevate: orgAllowContext,
			Handle:  handleOrganizationSettingsUpdateOneGala,
		},
	)
}

// handleOrganizationDeleteGala deactivates an organization's customer subscription when deleted.
func handleOrganizationDeleteGala(inv entityops.Invocation, _ entityops.MutationPayload) error {
	entInv, ok := newEntitlementInvocation(inv)
	if !ok {
		return nil
	}

	entInv.ctx = logx.WithFields(entInv.ctx, map[string]any{"organization_id": entInv.orgID})

	cleanupContext := entgen.NewContext(entInv.Context(), entInv.client)
	if err := entgen.OrganizationEdgeCleanup(cleanupContext, entInv.orgID); err != nil {
		entInv.Logger().Error().Err(err).Msg("failed to cascade delete organization edges")
		return err
	}

	org, err := entInv.client.Organization.Query().Where(
		organization.And(
			organization.ID(entInv.orgID),
			organization.DeletedAtNotNil(),
		),
	).Only(entInv.Context())
	if err != nil {
		entInv.Logger().Err(err).Msg("organization delete event unable to load organization")
		return nil
	}

	if org.StripeCustomerID == nil {
		return nil
	}

	if err := entInv.manager.FindAndDeactivateCustomerSubscription(entInv.Context(), *org.StripeCustomerID); err != nil {
		entInv.Logger().Error().Err(err).Msg("failed to deactivate customer subscription")
		return err
	}

	return nil
}

// handleOrganizationCreatedGala reconciles entitlements after organization creation.
func handleOrganizationCreatedGala(inv entityops.Invocation, _ entityops.MutationPayload) error {
	entInv, ok := newEntitlementInvocation(inv)
	if !ok {
		return nil
	}

	return entInv.reconcile()
}

// handleOrganizationSettingsUpdateOneGala updates Stripe customer details for billing changes.
func handleOrganizationSettingsUpdateOneGala(inv entityops.Invocation, payload entityops.MutationPayload) error {
	entInv, ok := newEntitlementInvocation(inv)
	if !ok {
		return nil
	}

	entInv.ctx = logx.WithFields(entInv.ctx, map[string]any{"organization_setting_id": inv.EntityID})

	orgSetting, err := entInv.client.OrganizationSetting.Get(entInv.Context(), inv.EntityID)
	if err != nil {
		entInv.Logger().Error().Err(err).Msg("failed to resolve organization from organization setting")

		return nil
	}

	entInv.orgID = orgSetting.OrganizationID

	orgCustomer, err := fetchOrganizationCustomer(entInv, orgSetting)
	if err != nil {
		entInv.Logger().Err(err).Msg("failed to fetch organization customer")
		return err
	}

	if orgCustomer.StripeCustomerID == "" {
		return entInv.reconcile()
	}

	proposedChanges, err := jsonx.Decode[map[string]any](payload.ProposedChanges)
	if err != nil {
		proposedChanges = nil
	}

	params := entitlements.GetUpdatedFields(proposedChanges, orgCustomer)

	if params != nil {
		if _, err := entInv.manager.UpdateCustomer(entInv.Context(), orgCustomer.StripeCustomerID, params); err != nil {
			entInv.Logger().Err(err).Str("stripe_customer_id", orgCustomer.StripeCustomerID).Msg("failed to update stripe customer metadata")
			return err
		}
	}

	return entInv.reconcile()
}

// entitlementInvocation bundles all data needed by entitlement listeners.
type entitlementInvocation struct {
	ctx     context.Context
	client  *entgen.Client
	manager *entitlements.StripeClient
	orgID   string
}

// Context returns the listener context associated with the invocation.
func (inv *entitlementInvocation) Context() context.Context {
	return inv.ctx
}

// Logger returns a contextual logger for the invocation.
func (inv *entitlementInvocation) Logger() *zerolog.Logger {
	return logx.FromContext(inv.Context())
}

// orgAllowContext bypasses privacy rules for entitlement logic against an organization.
func orgAllowContext(ctx context.Context, _ entityops.MutationPayload) context.Context {
	return privacy.DecisionContext(ctx, privacy.Allow)
}

// softDeleteAllowContext extends orgAllowContext to include soft-deleted records.
func softDeleteAllowContext(ctx context.Context, payload entityops.MutationPayload) context.Context {
	return entx.SkipSoftDelete(orgAllowContext(ctx, payload))
}

// newEntitlementInvocation gathers prerequisites for entitlement mutation handling; the
// organization defaults to the mutated entity and setting-scoped handlers override it
// after resolving their setting
func newEntitlementInvocation(inv entityops.Invocation) (*entitlementInvocation, bool) {
	manager, err := do.Invoke[*entitlements.StripeClient](inv.Injector)
	if err != nil || manager == nil || !manager.Config.IsEnabled() {
		return nil, false
	}

	return &entitlementInvocation{
		ctx:     inv.Context,
		client:  inv.Client,
		manager: manager,
		orgID:   inv.EntityID,
	}, true
}

// fetchOrganizationCustomer builds the organization customer view for a billing settings change
func fetchOrganizationCustomer(inv *entitlementInvocation, orgSetting *entgen.OrganizationSetting) (*entitlements.OrganizationCustomer, error) {
	org, err := inv.client.Organization.Query().
		Where(organization.ID(orgSetting.OrganizationID)).
		Only(inv.Context())
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
		reconciler.WithStripeClient(inv.manager),
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
