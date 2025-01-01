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
	// config is the configuration for the Stripe client
	Config Config
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
		sc.Config = config
	}
}

// WithAPIKey sets the API key for the Stripe client
func WithAPIKey(apiKey string) StripeOptions {
	return func(sc *StripeClient) {
		sc.apikey = apiKey
	}
}
