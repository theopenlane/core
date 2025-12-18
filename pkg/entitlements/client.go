package entitlements

import (
	"github.com/stripe/stripe-go/v84"
)

// StripeClient is a client for the Stripe API
type StripeClient struct {
	// Client is the Stripe client and is used for accessing all subsequent stripe objects, e.g. products, prices, etc.
	Client *stripe.Client
	// apikey is the Stripe API key
	apikey string
	// config is the configuration for the Stripe client
	Config *Config
	// Backends is a map of backend services
	backends *stripe.Backends
}

// NewStripeClient creates a new Stripe client
func NewStripeClient(opts ...StripeOptions) (*StripeClient, error) {
	sc := &StripeClient{}
	for _, opt := range opts {
		opt(sc)
	}

	if sc.apikey == "" {
		return nil, ErrMissingAPIKey
	}

	sc.Client = stripe.NewClient(sc.apikey, stripe.WithBackends(sc.backends))

	return sc, nil
}

// StripeOptions is a type for setting options on the Stripe client
type StripeOptions func(*StripeClient)

// WithConfig sets the config for the Stripe client
func WithConfig(config Config) StripeOptions {
	return func(sc *StripeClient) {
		sc.Config = &config
	}
}

// WithAPIKey sets the API key for the Stripe client
func WithAPIKey(apiKey string) StripeOptions {
	return func(sc *StripeClient) {
		sc.apikey = apiKey
	}
}

// WithBackends sets the backends for the Stripe client
func WithBackends(backends *stripe.Backends) StripeOptions {
	return func(sc *StripeClient) {
		sc.backends = backends
	}
}
