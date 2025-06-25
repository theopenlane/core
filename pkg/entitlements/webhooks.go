package entitlements

import (
	"context"
	"time"

	"github.com/stripe/stripe-go/v82"
)

// CreateWebhookEndpoint creates a webhook endpoint in Stripe and returns the resulting webhook endpoint
func (sc *StripeClient) CreateWebhookEndpoint(ctx context.Context, url string, events []string) (*stripe.WebhookEndpoint, error) {
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
