package entitlements

import (
	"context"
	"time"

	"github.com/stripe/stripe-go/v82"
)

// SupportedEventTypes defines the Stripe events the webhook handler listens for.
var SupportedEventTypes = []stripe.EventType{
	stripe.EventTypeCustomerSubscriptionUpdated,
	stripe.EventTypeCustomerSubscriptionDeleted,
	stripe.EventTypeCustomerSubscriptionPaused,
	stripe.EventTypeCustomerSubscriptionTrialWillEnd,
	stripe.EventTypePaymentMethodAttached,
}

// SupportedEventTypeStrings returns SupportedEventTypes as a slice of strings.
func SupportedEventTypeStrings() []string {
	events := make([]string, len(SupportedEventTypes))
	for i, e := range SupportedEventTypes {
		events[i] = string(e)
	}

	return events
}

// CreateWebhookEndpoint creates a webhook endpoint in Stripe and returns the resulting webhook endpoint
func (sc *StripeClient) CreateWebhookEndpoint(ctx context.Context, url string, events []string) (*stripe.WebhookEndpoint, error) {
	if len(events) == 0 {
		switch {
		case len(sc.Config.StripeWebhookEvents) > 0:
			events = sc.Config.StripeWebhookEvents
		default:
			events = SupportedEventTypeStrings()
		}
	}

	params := &stripe.WebhookEndpointCreateParams{
		URL:           stripe.String(url),
		EnabledEvents: stripe.StringSlice(events),
	}

	start := time.Now()
	endpoint, err := sc.Client.V1WebhookEndpoints.Create(ctx, params)
	duration := time.Since(start).Seconds()

	status := StatusSuccess
	if err != nil {
		status = StatusError
	}

	stripeRequestCounter.WithLabelValues("webhook_endpoints", status).Inc()
	stripeRequestDuration.WithLabelValues("webhook_endpoints", status).Observe(duration)

	if err != nil {
		return nil, err
	}

	return endpoint, nil
}
