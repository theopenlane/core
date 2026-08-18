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

// EntitlementListeners keeps Stripe customers and subscriptions in sync with organization
// lifecycle and billing changes
func EntitlementListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaOrganization,
			Operations: []string{entityops.OpCreate},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return auth.NewWebhookCaller("")
			},
			Handle: entityops.RequireDep(handleOrganizationCreatedGala),
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaOrganization,
			Operations: []string{entityops.OpSoftDelete, entityops.OpDelete, entityops.OpDeleteOne},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return auth.NewWebhookCaller("")
			},
			ContextKeys: []func(context.Context) context.Context{entx.SkipSoftDelete},
			Handle:      entityops.RequireDep(handleOrganizationDeleteGala),
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaOrganizationSetting,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields: []string{
				organizationsetting.FieldBillingEmail,
				organizationsetting.FieldBillingPhone,
				organizationsetting.FieldBillingAddress,
			},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return auth.NewWebhookCaller("")
			},
			Handle: entityops.RequireDep(handleOrganizationSettingsUpdateOneGala),
		},
	}
}

// handleOrganizationDeleteGala deactivates a deleted organization's customer subscription;
// hard deletes resolve the stripe customer from captured pre-delete values
func handleOrganizationDeleteGala(inv entityops.Invocation, payload entityops.MutationPayload, manager *entitlements.StripeClient) error {
	if !manager.Config.IsEnabled() {
		return nil
	}

	stripeCustomerID, _ := payload.OldStringValue(organization.FieldStripeCustomerID)

	if stripeCustomerID == "" {
		org, err := inv.Client.Organization.Query().Where(
			organization.And(
				organization.ID(inv.EntityID),
				organization.DeletedAtNotNil(),
			),
		).Only(inv.Context)

		switch {
		case entgen.IsNotFound(err):
			return nil
		case err != nil:
			logx.FromContext(inv.Context).Error().Err(err).Msg("organization delete event unable to load organization")

			return err
		}

		if org.StripeCustomerID == nil {
			return nil
		}

		stripeCustomerID = *org.StripeCustomerID
	}

	if err := manager.FindAndDeactivateCustomerSubscription(inv.Context, stripeCustomerID); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to deactivate customer subscription")

		return err
	}

	return nil
}

// handleOrganizationCreatedGala reconciles entitlements after organization creation
func handleOrganizationCreatedGala(inv entityops.Invocation, _ entityops.MutationPayload, manager *entitlements.StripeClient) error {
	if !manager.Config.IsEnabled() {
		return nil
	}

	return reconcileEntitlements(inv.Context, inv.Client, manager, inv.EntityID)
}

// handleOrganizationSettingsUpdateOneGala updates Stripe customer details for billing changes
func handleOrganizationSettingsUpdateOneGala(inv entityops.Invocation, payload entityops.MutationPayload, manager *entitlements.StripeClient) error {
	if !manager.Config.IsEnabled() {
		return nil
	}

	orgSetting, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.OrganizationSetting.Get)
	if err != nil || !ok {
		return err
	}

	orgID := orgSetting.OrganizationID

	orgCustomer, err := fetchOrganizationCustomer(inv.Context, inv.Client, orgSetting)
	if err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to fetch organization customer")

		return err
	}

	if orgCustomer.StripeCustomerID == "" {
		return reconcileEntitlements(inv.Context, inv.Client, manager, orgID)
	}

	params := entitlements.GetUpdatedFields(payload.ProposedChanges, orgCustomer)

	if params != nil {
		if _, err := manager.UpdateCustomer(inv.Context, orgCustomer.StripeCustomerID, params); err != nil {
			logx.FromContext(inv.Context).Error().Err(err).Str("stripe_customer_id", orgCustomer.StripeCustomerID).Msg("failed to update stripe customer metadata")

			return err
		}
	}

	return reconcileEntitlements(inv.Context, inv.Client, manager, orgID)
}

// fetchOrganizationCustomer builds the organization customer view for a billing settings change
func fetchOrganizationCustomer(ctx context.Context, client *entgen.Client, orgSetting *entgen.OrganizationSetting) (*entitlements.OrganizationCustomer, error) {
	org, err := client.Organization.Get(ctx, orgSetting.OrganizationID)
	if err != nil {
		return nil, err
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
		StripeCustomerID:       lo.FromPtr(org.StripeCustomerID),
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
