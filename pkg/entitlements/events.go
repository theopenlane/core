package entitlements

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
)

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
func (sc *StripeClient) HandleEvent(c context.Context, e *stripe.Event) error {
	switch e.Type {
	case "customer.subscription.updated", "customer.subscription.deleted", "customer.subscription.paused":
		subscription, err := unmarshalEventData[stripe.Subscription](e)
		if err == nil {
			return sc.handleSubscriptionUpdated(subscription)
		}
	case "payment_intent.succeeded":
		cust, err := unmarshalEventData[stripe.PaymentIntent](e)
		if err == nil {
			return sc.handlePaymentIntent(c, cust)
		}
	}

	return nil
}

// handleSubscriptionUpdated handles subscription updated events
func (sc *StripeClient) handleSubscriptionUpdated(s *stripe.Subscription) error {
	_, err := sc.GetSubscriptionByID(s.ID)
	if err != nil {
		return err
	}

	// TODO implement update logic; for now just confirm the lookup flow is successful and print the event

	return nil
}

// handlePaymentIntent handles payment intent events
func (sc *StripeClient) handlePaymentIntent(c context.Context, stripeCust *stripe.PaymentIntent) error {
	_, err := sc.GetCustomerByStripeID(c, stripeCust.ID)
	if err != nil {
		return err
	}

	// TODO implement payment intent logic; for now just confirm the lookup flow is successful and print the event

	return nil
}
