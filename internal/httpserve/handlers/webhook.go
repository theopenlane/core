package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	echo "github.com/theopenlane/echox"

	ent "github.com/theopenlane/core/internal/ent/generated"
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
			return h.handleSubscriptionUpdated(subscription)
		}
	case "payment_intent.succeeded":
		cust, err := unmarshalEventData[stripe.PaymentIntent](e)
		if err == nil {
			return h.handlePaymentIntent(c, cust)
		}
	case "customer.subscription.trial_will_end":
		cust, err := unmarshalEventData[stripe.PaymentIntent](e)
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

// getOrgByID returns the organization based on the id in the request
func (h *Handler) updateOrgSubscription(ctx context.Context, s *stripe.Subscription) (org *ent.Organization, err error) {
	// _ := h.Entitlements.MapStripeSubscription(s)
	stripeCust, err := h.Entitlements.GetCustomerByStripeID(ctx, s.Customer.ID)
	if err != nil {
		log.Error().Err(err).Msg("error obtaining customer from id")

		return nil, err
	}

	var orgSubscriptionID string

	for key, value := range stripeCust.Metadata {
		if key == "organization_subscription_id" {
			orgSubscriptionID = value
			break
		}
	}

	if orgSubscriptionID == "" {
		log.Error().Msg("organization_subscription_id not found in metadata")
		return nil, fmt.Errorf("organization_subscription_id not found in metadata")
	}

	// mapper := MapStripeToOrgSubscription(s, stripeCust)

	err = transaction.FromContext(ctx).OrgSubscription.UpdateOneID("").
		SetActive(false).
		SetProductPrice(models.Price{}).
		SetFeatures([]string{}).
		Exec(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error updating organization subscription")

		return nil, err
	}

	return org, nil
}

// handleSubscriptionUpdated handles subscription updated events
func (h *Handler) handleSubscriptionPaused(ctx context.Context, s *stripe.Subscription) (orgCust *entitlements.OrganizationCustomer, err error) {
	orgCust = &entitlements.OrganizationCustomer{}
	subs, err := h.Entitlements.GetSubscriptionByID(s.ID)

	if err != nil {
		return orgCust, err
	}

	internalSubs := h.Entitlements.MapStripeSubscription(subs)

	orgCust.Subscription = *internalSubs

	return orgCust, nil
}

// handleSubscriptionUpdated handles subscription updated events
func (h *Handler) handleSubscriptionUpdated(s *stripe.Subscription) (orgCust *entitlements.OrganizationCustomer, err error) {
	orgCust = &entitlements.OrganizationCustomer{}
	subs, err := h.Entitlements.GetSubscriptionByID(s.ID)

	if err != nil {
		return orgCust, err
	}

	internalSubs := h.Entitlements.MapStripeSubscription(subs)

	orgCust.Subscription = *internalSubs

	return orgCust, nil
}

// handlePaymentIntent handles payment intent events
func (h *Handler) handlePaymentIntent(c context.Context, stripeCust *stripe.PaymentIntent) (orgCust *entitlements.OrganizationCustomer, err error) {
	_, err = h.Entitlements.GetCustomerByStripeID(c, stripeCust.ID)
	if err != nil {
		return orgCust, err
	}

	// TODO implement payment intent logic; for now just confirm the lookup flow is successful and print the event

	return orgCust, nil
}

// handlePaymentIntent handles payment intent events
func (h *Handler) trialWillEnd(c context.Context, stripeCust *stripe.PaymentIntent) (orgCust *entitlements.OrganizationCustomer, err error) {
	_, err = h.Entitlements.GetCustomerByStripeID(c, stripeCust.ID)
	if err != nil {
		return orgCust, err
	}

	// TODO implement payment intent logic; for now just confirm the lookup flow is successful and print the event

	return orgCust, nil
}
