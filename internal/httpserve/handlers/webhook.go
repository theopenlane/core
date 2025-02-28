package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	echo "github.com/theopenlane/echox"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/apitoken"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/personalaccesstoken"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/models"
)

// WebhookReceiverHandler handles incoming stripe webhook events
func (h *Handler) WebhookReceiverHandler(ctx echo.Context) error {
	req := ctx.Request()
	res := ctx.Response()

	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(res.Writer, req.Body, MaxBodyBytes)

	payload, err := io.ReadAll(req.Body)
	if err != nil {
		return ctx.String(http.StatusServiceUnavailable, fmt.Errorf("problem with request. Error: %w", err).Error())
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), h.Entitlements.Config.StripeWebhookSecret)
	if err != nil {
		return ctx.String(http.StatusBadRequest, fmt.Errorf("error verifying webhook signature. Error: %w", err).Error())
	}

	exists, err := h.checkForEventID(req.Context(), event.ID)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	if !exists {
		_, err := h.Entitlements.HandleEvent(req.Context(), &event)
		if err != nil {
			return h.InternalServerError(ctx, err)
		}

		meow := "meow"

		input := ent.CreateEventInput{
			EventID:   event.ID,
			EventType: &meow,
			// TODO unmarshall event data into internal event
		}

		meowevent, err := h.createEvent(req.Context(), input)
		if err != nil {
			return h.InternalServerError(ctx, err)
		}

		log.Debug().Msgf("Internal event: %v", meowevent)
	}

	out := WebhookResponse{
		Message: "Received!",
	}

	return h.Success(ctx, out)
}

// WebhookRequest is the request object for the webhook handler
type WebhookRequest struct {
	// TODO determine if there's any request data actually need or needs to be validated given the signature verification that's already occurring
}

// WebhookResponse is the response object for the webhook handler
type WebhookResponse struct {
	Message string
}

// Validate validates the webhook request
func (r *WebhookRequest) Validate() error {
	return nil
}

// unmarshalEventData is used to unmarshal event data from a stripe.Event object into a specific type T
func unmarshalEventData[T interface{}](e *stripe.Event) (*T, error) {
	var data T

	err := json.Unmarshal(e.Data.Raw, &data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal event data")
	}

	return &data, err
}

// HandleEvent umarshals event data and triggers a corresponding function to be executed based on case match
func (h *Handler) HandleEvent(c context.Context, e *stripe.Event) (orgCust *entitlements.OrganizationCustomer, err error) {
	switch e.Type {
	case "customer.subscription.updated":
		subscription, err := unmarshalEventData[stripe.Subscription](e)
		if err == nil {
			return h.handleSubscriptionUpdated(c, subscription)
		}
	case "payment_method.attached":
		cust, err := unmarshalEventData[stripe.PaymentMethod](e)
		if err == nil {
			return h.handlePaymentMethodAdded(c, cust)
		}
	case "customer.subscription.trial_will_end":
		cust, err := unmarshalEventData[stripe.Subscription](e)
		if err == nil {
			return h.trialWillEnd(c, cust)
		}
	case "customer.subscription.deleted", "customer.subscription.paused":
		cust, err := unmarshalEventData[stripe.Subscription](e)
		if err == nil {
			return h.handleSubscriptionPaused(c, cust)
		}
	}

	return orgCust, nil
}

// InvalidateAPITokens invalidates all API tokens for an organization
func InvalidateAPITokens(c context.Context, orgID string) error {
	if err := transaction.FromContext(c).APIToken.Update().Where(apitoken.OwnerID(orgID)).
		SetIsActive(false).
		SetExpiresAt(time.Now().Add(-24 * time.Hour)).
		SetRevokedAt(time.Now()).
		SetRevokedReason("subscription paused or deleted").
		SetRevokedBy("entitlements_engine").
		Exec(c); err != nil {
		return err
	}

	return nil
}

// InvalidatePATokens invalidates all personal access tokens for an organization
func InvalidatePATokens(c context.Context, orgID string) error {
	if err := transaction.FromContext(c).PersonalAccessToken.Update().Where(personalaccesstoken.OwnerID(orgID)).
		SetIsActive(false).
		SetExpiresAt(time.Now().Add(-24 * time.Hour)).
		SetRevokedAt(time.Now()).
		SetRevokedReason("subscription paused or deleted").
		SetRevokedBy("entitlements_engine").
		Exec(c); err != nil {
		return err
	}

	return nil
}

// handleSubscriptionUpdated handles subscription updated events
func (h *Handler) handleSubscriptionPaused(ctx context.Context, s *stripe.Subscription) (orgCust *entitlements.OrganizationCustomer, err error) {
	orgSubs, err := SyncOrgSubscriptionWithStripe(ctx, s, nil)
	if err != nil {
		return orgCust, err
	}

	if err := transaction.FromContext(ctx).APIToken.Update().Where(apitoken.OwnerID(orgSubs.ID)).
		SetIsActive(false).
		SetExpiresAt(time.Now().Add(-24 * time.Hour)).
		SetRevokedAt(time.Now()).
		SetRevokedReason("subscription paused or deleted").
		SetRevokedBy("entitlements_engine").
		Exec(ctx); err != nil {
		return orgCust, err
	}

	return orgCust, nil
}

// handleSubscriptionUpdated handles subscription updated events
func (h *Handler) handleSubscriptionUpdated(ctx context.Context, s *stripe.Subscription) (orgCust *entitlements.OrganizationCustomer, err error) {
	_, err = SyncOrgSubscriptionWithStripe(ctx, s, nil)
	if err != nil {
		return orgCust, err
	}

	return nil, nil
}

// handlePaymentIntent handles payment intent events
func (h *Handler) handlePaymentMethodAdded(ctx context.Context, stripeCust *stripe.PaymentMethod) (orgCust *entitlements.OrganizationCustomer, err error) {

	return orgCust, nil
}

// handlePaymentIntent handles payment intent events
func (h *Handler) trialWillEnd(c context.Context, stripeCust *stripe.Subscription) (orgCust *entitlements.OrganizationCustomer, err error) {
	_, err = h.Entitlements.GetCustomerByStripeID(c, stripeCust.ID)
	if err != nil {
		return orgCust, err
	}

	// TODO implement payment intent logic; for now just confirm the lookup flow is successful and print the event

	return orgCust, nil
}

func getOrgSubscription(ctx context.Context, subscription *stripe.Subscription) (*ent.OrgSubscription, error) {
	orgSubscription, err := transaction.FromContext(ctx).OrgSubscription.Query().Where(orgsubscription.StripeSubscriptionID(subscription.ID)).Only(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to find OrgSubscription")
		return nil, err
	}

	return orgSubscription, nil
}

// SyncOrgSubscriptionWithStripe updates the internal OrgSubscription record with data from Stripe
func SyncOrgSubscriptionWithStripe(ctx context.Context, subscription *stripe.Subscription, customer *stripe.Customer) (*ent.OrgSubscription, error) {
	orgSubscription, err := getOrgSubscription(ctx, subscription)
	if err != nil {
		return nil, err
	}

	// map stripe data to internal OrgSubscription
	stripeOrgSubscription := MapStripeToOrgSubscription(subscription, MapStripeCustomer(customer))

	// Check if any fields have changed before saving the updated OrgSubscription
	changed := false

	if orgSubscription.StripeSubscriptionStatus != stripeOrgSubscription.StripeSubscriptionStatus {
		orgSubscription.StripeSubscriptionStatus = stripeOrgSubscription.StripeSubscriptionStatus
		changed = true

		log.Debug().Msgf("Stripe subscription status changed to %s", orgSubscription.StripeSubscriptionStatus)
	}

	if !equalSlices(orgSubscription.Features, stripeOrgSubscription.Features) {
		orgSubscription.Features = stripeOrgSubscription.Features
		changed = true

		log.Debug().Msgf("Features changed to %v", orgSubscription.Features)
	}

	if orgSubscription.ProductTier != stripeOrgSubscription.ProductTier {
		orgSubscription.ProductTier = stripeOrgSubscription.ProductTier
		changed = true

		log.Debug().Msgf("Product tier changed to %s", orgSubscription.ProductTier)
	}

	if orgSubscription.ProductPrice != stripeOrgSubscription.ProductPrice {
		orgSubscription.ProductPrice = stripeOrgSubscription.ProductPrice
		changed = true

		log.Debug().Msgf("Product price changed to %v", orgSubscription.ProductPrice)
	}

	if orgSubscription.ExpiresAt != stripeOrgSubscription.ExpiresAt {
		orgSubscription.ExpiresAt = stripeOrgSubscription.ExpiresAt
		changed = true

		log.Debug().Msgf("Subscription expires at %v", orgSubscription.ExpiresAt)
	}

	if !equalSlices(orgSubscription.FeatureLookupKeys, stripeOrgSubscription.FeatureLookupKeys) {
		orgSubscription.FeatureLookupKeys = stripeOrgSubscription.FeatureLookupKeys
		changed = true

		log.Debug().Msgf("Feature lookup keys changed to %v", orgSubscription.FeatureLookupKeys)
	}

	if orgSubscription.TrialExpiresAt != stripeOrgSubscription.TrialExpiresAt {
		orgSubscription.TrialExpiresAt = stripeOrgSubscription.TrialExpiresAt
		changed = true

		log.Debug().Msgf("Trial expires at %v", orgSubscription.TrialExpiresAt)
	}

	if orgSubscription.DaysUntilDue != stripeOrgSubscription.DaysUntilDue {
		orgSubscription.DaysUntilDue = stripeOrgSubscription.DaysUntilDue
		changed = true

		log.Debug().Msgf("Days until due changed to %s", *orgSubscription.DaysUntilDue)
	}

	if orgSubscription.PaymentMethodAdded != stripeOrgSubscription.PaymentMethodAdded {
		orgSubscription.PaymentMethodAdded = stripeOrgSubscription.PaymentMethodAdded
		changed = true

		log.Debug().Msgf("Payment method added changed to %t", *orgSubscription.PaymentMethodAdded)
	}

	if orgSubscription.Active != stripeOrgSubscription.Active {
		orgSubscription.Active = stripeOrgSubscription.Active
		changed = true

		log.Debug().Msgf("Subscription active status changed to %t", orgSubscription.Active)
	}

	if changed {
		// Save the updated OrgSubscription
		err = transaction.FromContext(ctx).OrgSubscription.UpdateOne(orgSubscription).Exec(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to update OrgSubscription")
			return nil, err
		}
	}

	return orgSubscription, nil
}

// MapStripeCustomer maps a stripe.Customer to an OrganizationCustomer
func MapStripeCustomer(c *stripe.Customer) *entitlements.OrganizationCustomer {
	var orgID, orgSettingsID, orgSubscriptionID string

	for k, v := range c.Metadata {
		switch k {
		case "organization_id":
			orgID = v
		case "organization_settings_id":
			orgSettingsID = v
		case "organization_subscription_id":
			orgSubscriptionID = v
		}
	}

	paymentAdded := false
	if c.Sources.Data != nil {
		paymentAdded = true
	}

	return &entitlements.OrganizationCustomer{
		StripeCustomerID:           c.ID,
		OrganizationID:             orgID,
		OrganizationName:           c.Metadata["organization_name"],
		OrganizationSettingsID:     orgSettingsID,
		OrganizationSubscriptionID: orgSubscriptionID,
		PaymentMethodAdded:         paymentAdded,
	}
}

// MapStripeToOrgSubscription maps a stripe.Subscription and OrganizationCustomer to an ent.OrgSubscription
func MapStripeToOrgSubscription(subscription *stripe.Subscription, customer *entitlements.OrganizationCustomer) *ent.OrgSubscription {
	productName := ""
	productPrice := models.Price{}

	if len(subscription.Items.Data) == 1 {
		productName = subscription.Items.Data[0].Price.Product.Name
		productPrice.Amount = subscription.Items.Data[0].Price.UnitAmountDecimal
		productPrice.Currency = string(subscription.Items.Data[0].Price.Currency)
		productPrice.Interval = string(subscription.Items.Data[0].Price.Recurring.Interval)
	}

	active := false
	if subscription.Status == "active" || subscription.Status == "trialing" {
		active = true
	}

	orgSubscription := &ent.OrgSubscription{
		StripeSubscriptionID:     subscription.ID,
		TrialExpiresAt:           timePtr(time.Unix(subscription.TrialEnd, 0)),
		StripeSubscriptionStatus: string(subscription.Status),
		Active:                   active,
		ProductTier:              productName,
		ProductPrice:             productPrice,
		Features:                 customer.Features,
		FeatureLookupKeys:        customer.FeatureNames,
		DaysUntilDue:             int64ToStringPtr(subscription.DaysUntilDue),
	}

	return orgSubscription
}

// int64ToStringPtr converts an int64 to a *string
func int64ToStringPtr(i int64) *string {
	s := fmt.Sprintf("%d", i)
	return &s
}

// equalSlices checks if two slices are equal
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// timePtr returns a pointer to the given time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}
