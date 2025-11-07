package entitlements

import (
	"context"
	"time"

	"github.com/stripe/stripe-go/v83"
)

const (
	// webhookEndpointsListLimit is the maximum number of webhook endpoints to retrieve in a single list operation
	webhookEndpointsListLimit = 100
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
func (sc *StripeClient) CreateWebhookEndpoint(ctx context.Context, url string, events []string, apiVersion string, disabled bool) (*stripe.WebhookEndpoint, error) {
	if len(events) == 0 {
		switch {
		case sc.Config != nil && len(sc.Config.StripeWebhookEvents) > 0:
			events = sc.Config.StripeWebhookEvents
		default:
			events = SupportedEventTypeStrings()
		}
	}

	webhookURL := url
	if apiVersion != "" {
		webhookURL = addVersionParam(url, apiVersion)
	}

	params := &stripe.WebhookEndpointCreateParams{
		URL:           stripe.String(webhookURL),
		EnabledEvents: stripe.StringSlice(events),
	}

	if apiVersion != "" {
		params.APIVersion = stripe.String(apiVersion)
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

	if disabled {
		endpoint, err = sc.DisableWebhookEndpoint(ctx, endpoint.ID)
		if err != nil {
			return nil, err
		}
	}

	return endpoint, nil
}

// ListWebhookEndpoints lists all webhook endpoints in Stripe
func (sc *StripeClient) ListWebhookEndpoints(ctx context.Context) ([]*stripe.WebhookEndpoint, error) {
	params := &stripe.WebhookEndpointListParams{}
	params.Limit = stripe.Int64(webhookEndpointsListLimit)

	start := time.Now()
	iter := sc.Client.V1WebhookEndpoints.List(ctx, params)

	var endpoints []*stripe.WebhookEndpoint
	var lastErr error

	for endpoint, err := range iter {
		if err != nil {
			lastErr = err
			break
		}
		endpoints = append(endpoints, endpoint)
	}

	duration := time.Since(start).Seconds()

	status := StatusSuccess
	if lastErr != nil {
		status = StatusError
	}

	stripeRequestCounter.WithLabelValues("webhook_endpoints_list", status).Inc()
	stripeRequestDuration.WithLabelValues("webhook_endpoints_list", status).Observe(duration)

	if lastErr != nil {
		return nil, lastErr
	}

	return endpoints, nil
}

// GetWebhookEndpoint retrieves a webhook endpoint by ID from Stripe
func (sc *StripeClient) GetWebhookEndpoint(ctx context.Context, id string) (*stripe.WebhookEndpoint, error) {
	start := time.Now()
	endpoint, err := sc.Client.V1WebhookEndpoints.Retrieve(ctx, id, &stripe.WebhookEndpointRetrieveParams{})
	duration := time.Since(start).Seconds()

	status := StatusSuccess
	if err != nil {
		status = StatusError
	}

	stripeRequestCounter.WithLabelValues("webhook_endpoint_get", status).Inc()
	stripeRequestDuration.WithLabelValues("webhook_endpoint_get", status).Observe(duration)

	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

// UpdateWebhookEndpoint updates a webhook endpoint in Stripe
func (sc *StripeClient) UpdateWebhookEndpoint(ctx context.Context, id string, params *stripe.WebhookEndpointUpdateParams) (*stripe.WebhookEndpoint, error) {
	start := time.Now()
	endpoint, err := sc.Client.V1WebhookEndpoints.Update(ctx, id, params)
	duration := time.Since(start).Seconds()

	status := StatusSuccess
	if err != nil {
		status = StatusError
	}

	stripeRequestCounter.WithLabelValues("webhook_endpoint_update", status).Inc()
	stripeRequestDuration.WithLabelValues("webhook_endpoint_update", status).Observe(duration)

	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

// DisableWebhookEndpoint disables a webhook endpoint in Stripe
func (sc *StripeClient) DisableWebhookEndpoint(ctx context.Context, id string) (*stripe.WebhookEndpoint, error) {
	params := &stripe.WebhookEndpointUpdateParams{
		Disabled: stripe.Bool(true),
	}

	return sc.UpdateWebhookEndpoint(ctx, id, params)
}

// EnableWebhookEndpoint enables a webhook endpoint in Stripe
func (sc *StripeClient) EnableWebhookEndpoint(ctx context.Context, id string) (*stripe.WebhookEndpoint, error) {
	params := &stripe.WebhookEndpointUpdateParams{
		Disabled: stripe.Bool(false),
	}

	return sc.UpdateWebhookEndpoint(ctx, id, params)
}

// DeleteWebhookEndpoint deletes a webhook endpoint in Stripe
func (sc *StripeClient) DeleteWebhookEndpoint(ctx context.Context, id string) (*stripe.WebhookEndpoint, error) {
	start := time.Now()
	endpoint, err := sc.Client.V1WebhookEndpoints.Delete(ctx, id, &stripe.WebhookEndpointDeleteParams{})
	duration := time.Since(start).Seconds()

	status := StatusSuccess
	if err != nil {
		status = StatusError
	}

	stripeRequestCounter.WithLabelValues("webhook_endpoint_delete", status).Inc()
	stripeRequestDuration.WithLabelValues("webhook_endpoint_delete", status).Observe(duration)

	if err != nil {
		return nil, err
	}

	return endpoint, nil
}
