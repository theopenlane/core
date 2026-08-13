package hooks

import (
	"context"
	"errors"

	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// EntitlementListeners returns the entitlement mutation listeners
func EntitlementListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entgen.TypeOrganization,
			Label:      "entitlements_created",
			Operations: []string{entityops.OpCreate},
			Caller:     webhookCaller,
			Handle:     handleOrganizationCreatedGala,
		},
		entityops.MutationListener{
			Schema:      entgen.TypeOrganization,
			Label:       "entitlements_deleted",
			Operations:  []string{entityops.OpSoftDelete, entityops.OpDelete, entityops.OpDeleteOne},
			Caller:      webhookCaller,
			ContextKeys: []func(context.Context) context.Context{entx.SkipSoftDelete},
			Handle:      handleOrganizationDeleteGala,
		},
		entityops.MutationListener{
			Schema:     entgen.TypeOrganizationSetting,
			Label:      "billing",
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields: []string{
				organizationsetting.FieldBillingEmail,
				organizationsetting.FieldBillingPhone,
				organizationsetting.FieldBillingAddress,
			},
			Caller: webhookCaller,
			Handle: handleOrganizationSettingsUpdateOneGala,
		},
	}
}

// webhookCaller runs entitlement listeners under a webhook service caller
func webhookCaller(*auth.Caller, entityops.MutationPayload) *auth.Caller {
	return auth.NewWebhookCaller("")
}

// handleOrganizationDeleteGala deactivates an organization's customer subscription when deleted.
// Hard deletes resolve the stripe customer from the captured pre-delete values; soft deletes
// load the still-present row
func handleOrganizationDeleteGala(inv entityops.Invocation, payload entityops.MutationPayload) error {
	manager, ok := gala.Resolve[*entitlements.StripeClient](inv.Context, inv.Injector, "entitlements_deleted")
	if !ok || !manager.Config.IsEnabled() {
		return nil
	}

	ctx := inv.Context

	stripeCustomerID, _ := payload.OldStringValue(organization.FieldStripeCustomerID)

	if stripeCustomerID == "" {
		org, err := inv.Client.Organization.Query().Where(
			organization.And(
				organization.ID(inv.EntityID),
				organization.DeletedAtNotNil(),
			),
		).Only(ctx)

		switch {
		case entgen.IsNotFound(err):
			return nil
		case err != nil:
			logx.FromContext(ctx).Error().Err(err).Msg("organization delete event unable to load organization")

			return err
		}

		if org.StripeCustomerID == nil {
			return nil
		}

		stripeCustomerID = *org.StripeCustomerID
	}

	if err := manager.FindAndDeactivateCustomerSubscription(ctx, stripeCustomerID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to deactivate customer subscription")

		return err
	}

	return nil
}

// handleOrganizationCreatedGala reconciles entitlements after organization creation
func handleOrganizationCreatedGala(inv entityops.Invocation, _ entityops.MutationPayload) error {
	manager, ok := gala.Resolve[*entitlements.StripeClient](inv.Context, inv.Injector, "entitlements_created")
	if !ok || !manager.Config.IsEnabled() {
		return nil
	}

	return reconcileEntitlements(inv.Context, inv.Client, manager, inv.EntityID)
}

// handleOrganizationSettingsUpdateOneGala updates Stripe customer details for billing changes
func handleOrganizationSettingsUpdateOneGala(inv entityops.Invocation, payload entityops.MutationPayload) error {
	manager, ok := gala.Resolve[*entitlements.StripeClient](inv.Context, inv.Injector, "billing")
	if !ok || !manager.Config.IsEnabled() {
		return nil
	}

	ctx := inv.Context

	orgSetting, ok, err := entityops.LoadEntity(ctx, inv.EntityID, inv.Client.OrganizationSetting.Get)
	if err != nil || !ok {
		return err
	}

	orgID := orgSetting.OrganizationID

	orgCustomer, err := fetchOrganizationCustomer(ctx, inv.Client, orgSetting)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to fetch organization customer")

		return err
	}

	if orgCustomer.StripeCustomerID == "" {
		return reconcileEntitlements(ctx, inv.Client, manager, orgID)
	}

	params := entitlements.GetUpdatedFields(payload.ProposedMap(), orgCustomer)

	if params != nil {
		if _, err := manager.UpdateCustomer(ctx, orgCustomer.StripeCustomerID, params); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("stripe_customer_id", orgCustomer.StripeCustomerID).Msg("failed to update stripe customer metadata")

			return err
		}
	}

	return reconcileEntitlements(ctx, inv.Client, manager, orgID)
}

// fetchOrganizationCustomer builds the organization customer view for a billing settings change
func fetchOrganizationCustomer(ctx context.Context, client *entgen.Client, orgSetting *entgen.OrganizationSetting) (*entitlements.OrganizationCustomer, error) {
	org, err := client.Organization.Get(ctx, orgSetting.OrganizationID)
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

// reconcileEntitlements runs entitlement reconciliation for the given organization
func reconcileEntitlements(ctx context.Context, client *entgen.Client, manager *entitlements.StripeClient, orgID string) error {
	rec, err := reconciler.New(
		reconciler.WithDB(client),
		reconciler.WithStripeClient(manager),
	)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to construct entitlement reconciler")

		return err
	}

	if _, err := rec.Reconcile(ctx, []string{orgID}); err != nil {
		unwrapped := errors.Unwrap(err)
		// if this is a constraint error, log as warning - this is common in tests and we will still have the logs
		// in production
		if unwrapped != nil && entgen.IsConstraintError(unwrapped) {
			logx.FromContext(ctx).Warn().Err(err).Msg("entitlement reconciliation failed, organization with stripe customer id already exists")

			// do not retry a constraint error
			return nil
		}

		logx.FromContext(ctx).Error().Err(err).Msg("entitlement reconciliation failed")

		return err
	}

	logx.FromContext(ctx).Debug().Msg("entitlement reconciliation completed")

	return nil
}
