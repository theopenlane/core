package entitlements

import (
	"github.com/stripe/stripe-go/v81/client"
)

// StripeClient is a client for the Stripe API
type StripeClient struct {
	// Client is the Stripe client and is used for accessing all subsequent stripe objects, e.g. products, prices, etc.
	Client *client.API
	// apikey is the Stripe API key
	apikey string
	// webhooksecret is the stripe signing secret to verify webhook signatures
	webhooksecret string

	config Config
	// TODO[MKA] cleanup these fields and determine if an interface is easier to use than internal struct mapping and how to mock out the use of the stripe client
	// Customer is a ref to a generic Customer struct used to wrap Stripe customer and Openlane Organizations (typically)
	Cust Customer
	// Plans are a nomenclature for the recurring context that holds the payment information and is synonymous with Stripe subscriptions
	Plan Plan
	// Product is a stripe product; also know as a "tier"
	Product Product
	// Price holds the interval and amount to be billed
	Price Price
}

// NewStripeClient creates a new Stripe client
func NewStripeClient(opts ...StripeOptions) *StripeClient {
	sc := &StripeClient{}
	for _, opt := range opts {
		opt(sc)
	}

	sc.Client = client.New(sc.apikey, nil)

	return sc
}

// StripeOptions is a type for setting options on the Stripe client
type StripeOptions func(*StripeClient)

// WithConfig sets the config for the Stripe client
func WithConfig(config Config) StripeOptions {
	return func(sc *StripeClient) {
		sc.config = config
	}
}

// WithAPIKey sets the API key for the Stripe client
func WithAPIKey(apiKey string) StripeOptions {
	return func(sc *StripeClient) {
		sc.apikey = apiKey
	}
}

// WithWebhookSecret sets the webhook secret for the Stripe client
func WithWebhookSecret(webhookSecret string) StripeOptions {
	return func(sc *StripeClient) {
		sc.webhooksecret = webhookSecret
	}
}
