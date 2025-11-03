package hooks

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/stripe/stripe-go/v83"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	catalog "github.com/theopenlane/core/internal/entitlements/entmapping"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// handleOrganizationDelete handles the deletion of an organization and deletes the customer in Stripe.
func handleOrganizationDelete(ctx *soiree.ListenerContext[*entgen.Client, soiree.Event]) error {
	client, ok := ctx.Client()
	if !ok {
		zerolog.Ctx(ctx.Context()).Debug().Msg("failed to cast event client to entgen.Client, skipping customer deletion")

		return nil
	}

	entMgr := client.EntitlementManager
	if entMgr == nil {
		zerolog.Ctx(ctx.Context()).Debug().Msg("EntitlementManager not found on client, skipping customer deletion")

		return nil
	}

	value, ok := ctx.Properties().Get(propertyOrganizationDataKey)
	if !ok {
		zerolog.Ctx(ctx.Context()).Debug().Msg("organization delete event missing hydrated organization payload")

		return nil
	}

	data, _ := value.(*organizationEventData)
	if data == nil || data.Organization == nil {
		zerolog.Ctx(ctx.Context()).Debug().Msg("organization delete event missing hydrated organization payload")

		return nil
	}

	if data.Organization.StripeCustomerID == nil {
		return nil
	}

	if err := entMgr.FindAndDeactivateCustomerSubscription(ctx.Context(), *data.Organization.StripeCustomerID); err != nil {
		zerolog.Ctx(ctx.Context()).Error().Err(err).Msg("failed to deactivate customer subscription")

		return err
	}

	return nil
}

// handleOrganizationCreated checks for the creation of an organization subscription and creates a customer in Stripe.
func handleOrganizationCreated(ctx *soiree.ListenerContext[*entgen.Client, soiree.Event]) error {
	client, ok := ctx.Client()
	if !ok {
		zerolog.Ctx(ctx.Context()).Debug().Msg("failed to cast event client to entgen.Client, skipping customer creation")

		return nil
	}

	entMgr := client.EntitlementManager
	if entMgr == nil {
		zerolog.Ctx(ctx.Context()).Debug().Msg("EntitlementManager not found on client, skipping customer creation")

		return nil
	}

	value, ok := ctx.Properties().Get(propertyOrganizationDataKey)
	if !ok {
		zerolog.Ctx(ctx.Context()).Debug().Msg("organization create event missing hydrated organization payload")

		return nil
	}

	data, _ := value.(*organizationEventData)
	if data == nil || data.Organization == nil {
		zerolog.Ctx(ctx.Context()).Debug().Msg("organization create event missing hydrated organization payload")

		return nil
	}

	if data.Organization.PersonalOrg {
		return nil
	}

	if data.Customer == nil || data.Subscription == nil {
		zerolog.Ctx(ctx.Context()).Debug().Msg("organization event missing subscription/customer payload, skipping customer creation")

		return nil
	}

	orgCustomer := catalog.PopulatePricesForOrganizationCustomer(data.Customer, client.EntConfig.Modules.UseSandbox)

	zerolog.Ctx(ctx.Context()).Debug().Msgf("Prices attached to organization customer: %+v", orgCustomer.Prices)

	runCtx := data.AllowCtx
	if runCtx == nil {
		runCtx = ctx.Context()
	}

	if err := entMgr.CreateCustomerAndSubscription(runCtx, orgCustomer); err != nil {
		zerolog.Ctx(ctx.Context()).Err(err).Msg("Failed to create customer")

		return err
	}

	if err := updateCustomerOrgSub(runCtx, orgCustomer, client); err != nil {
		zerolog.Ctx(ctx.Context()).Err(err).Msg("Failed to map customer to org subscription")

		return err
	}

	return nil
}

// handleOrganizationSettingsUpdateOne handles the update of an organization setting and updates the customer in Stripe.
func handleOrganizationSettingsUpdateOne(ctx *soiree.ListenerContext[*entgen.Client, soiree.Event]) error {
	client, ok := ctx.Client()
	if !ok {
		zerolog.Ctx(ctx.Context()).Debug().Msg("failed to cast event client to entgen.Client, skipping customer creation")

		return nil
	}

	entMgr := client.EntitlementManager
	if entMgr == nil {
		zerolog.Ctx(ctx.Context()).Debug().Msg("EntitlementManager not found on client, skipping customer creation")

		return nil
	}

	value, ok := ctx.Properties().Get(propertyOrgSettingCustomerKey)
	if !ok {
		zerolog.Ctx(ctx.Context()).Debug().Msg("organization setting event missing hydrated customer payload, skipping customer update")

		return nil
	}

	orgCust, _ := value.(*entitlements.OrganizationCustomer)
	if orgCust == nil {
		zerolog.Ctx(ctx.Context()).Debug().Msg("organization setting event missing hydrated customer payload, skipping customer update")

		return nil
	}

	if orgCust.StripeCustomerID == "" {
		return nil
	}

	params := entitlements.GetUpdatedFields(ctx.Properties().Raw(), orgCust)
	if params == nil {
		return nil
	}

	if _, err := entMgr.UpdateCustomer(ctx.Context(), orgCust.StripeCustomerID, params); err != nil {
		zerolog.Ctx(ctx.Context()).Err(err).Msg("Failed to update customer")

		return err
	}

	return nil
}

// updateCustomerOrgSub maps the customer fields to the organization subscription and update the organization subscription in the database
func updateCustomerOrgSub(ctx context.Context, customer *entitlements.OrganizationCustomer, client any) error {
	if customer == nil || customer.OrganizationSubscriptionID == "" {
		zerolog.Ctx(ctx).Error().Msg("organization subscription ID is empty on customer, unable to update organization subscription")

		return ErrNoSubscriptions
	}

	// update the expiration date based on the subscription status
	// if the subscription is trialing, set the expiration date to the trial end date
	// otherwise, set the expiration date to the end date if it exists
	trialExpiresAt := time.Unix(0, 0)
	if customer.Status == string(stripe.SubscriptionStatusTrialing) {
		trialExpiresAt = time.Unix(customer.TrialEnd, 0)
	}

	expiresAt := time.Unix(0, 0)
	if customer.EndDate > 0 {
		expiresAt = time.Unix(customer.EndDate, 0)
	}

	active := customer.Status == string(stripe.SubscriptionStatusActive) || customer.Status == string(stripe.SubscriptionStatusTrialing)

	c := client.(*entgen.Client)

	err := c.Organization.UpdateOneID(customer.OrganizationID).
		SetStripeCustomerID(customer.StripeCustomerID).
		Exec(ctx)
	if err != nil {
		return err
	}

	update := c.OrgSubscription.UpdateOneID(customer.OrganizationSubscriptionID).
		SetStripeSubscriptionID(customer.StripeSubscriptionID).
		SetStripeSubscriptionStatus(customer.Subscription.Status).
		SetActive(active)

	// ensure the correct expiration date is set based on the subscription status
	// if the subscription is trialing, set the expiration date to the trial end date
	// otherwise, set the expiration date to the end date
	if customer.Status == string(stripe.SubscriptionStatusTrialing) {
		update.SetTrialExpiresAt(trialExpiresAt)
	} else {
		update.SetExpiresAt(expiresAt)
	}

	return update.Exec(ctx)
}

// updateOrgCustomerWithSubscription updates the organization customer with the subscription data
// by querying the organization and organization settings
func updateOrgCustomerWithSubscription(ctx context.Context, orgSubs *entgen.OrgSubscription,
	o *entitlements.OrganizationCustomer, org *entgen.Organization) (*entitlements.OrganizationCustomer, error) {
	if orgSubs == nil || org == nil {
		return nil, ErrNoSubscriptions
	}

	if org.Edges.Setting != nil {
		o.OrganizationSettingsID = org.Edges.Setting.ID
	} else {
		zerolog.Ctx(ctx).Debug().Msgf("Organization setting is nil for organization ID %s", orgSubs.OwnerID)
	}

	o.OrganizationID = org.ID
	o.OrganizationName = org.Name
	o.OrganizationSettingsID = org.Edges.Setting.ID
	o.Email = org.Edges.Setting.BillingEmail

	return o, nil
}

// fetchOrganizationCustomerByOrgSettingID fetches the organization customer data based on the organization setting ID
func fetchOrganizationCustomerByOrgSettingID(ctx context.Context, orgSettingID string, client any) (*entitlements.OrganizationCustomer, error) {
	orgSetting, err := client.(*entgen.Client).OrganizationSetting.Get(ctx, orgSettingID)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msgf("Failed to fetch organization setting ID %s", orgSettingID)

		return nil, err
	}

	org, err := client.(*entgen.Client).Organization.
		Query().
		Where(organization.ID(orgSetting.OrganizationID)).
		Only(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msgf("Failed to fetch organization by organization setting ID %s", orgSettingID)

		return nil, err
	}

	stripeCustomerID := ""
	if org.StripeCustomerID != nil {
		stripeCustomerID = *org.StripeCustomerID
	}

	return &entitlements.OrganizationCustomer{
		OrganizationID:         org.ID,
		OrganizationName:       org.Name,
		StripeCustomerID:       stripeCustomerID,
		OrganizationSettingsID: orgSetting.ID,
		ContactInfo: entitlements.ContactInfo{
			Email:      orgSetting.BillingEmail,
			Phone:      orgSetting.BillingPhone,
			Line1:      &orgSetting.BillingAddress.Line1,
			Line2:      &orgSetting.BillingAddress.Line2,
			City:       &orgSetting.BillingAddress.City,
			State:      &orgSetting.BillingAddress.State,
			Country:    &orgSetting.BillingAddress.Country,
			PostalCode: &orgSetting.BillingAddress.PostalCode,
		},
	}, nil
}
