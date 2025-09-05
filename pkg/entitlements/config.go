package entitlements

type Config struct {
	// Enabled determines if the entitlements service is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// PublicStripeKey is the key for the stripe service
	PublicStripeKey string `json:"publicStripeKey" koanf:"publicStripeKey" default:""`
	// PrivateStripeKey is the key for the stripe service
	PrivateStripeKey string `json:"privateStripeKey" koanf:"privateStripeKey" default:"" sensitive:"true"`
	// StripeWebhookSecret is the secret for the stripe service
	StripeWebhookSecret string `json:"stripeWebhookSecret" koanf:"stripeWebhookSecret" default:"" sensitive:"true"`
	// StripeWebhookURL is the URL for the stripe webhook
	StripeWebhookURL string `json:"stripeWebhookURL" koanf:"stripeWebhookURL" default:"https://api.openlane.com/v1/stripe/webhook" domain:"inherit" domainPrefix:"https://api" domainSuffix:"/v1/stripe/webhook"`
	// StripeBillingPortalSuccessURL
	StripeBillingPortalSuccessURL string `json:"stripeBillingPortalSuccessURL" koanf:"stripeBillingPortalSuccessURL" default:"https://console.openlane.com/billing" domain:"inherit" domainPrefix:"https://console" domainSuffix:"/billing"`
	// StripeCancellationReturnURL is the URL for the stripe cancellation return
	StripeCancellationReturnURL string `json:"stripeCancellationReturnURL" koanf:"stripeCancellationReturnURL" default:"https://console.theopenlane.io/organization-settings/billing/subscription_canceled" domain:"inherit" domainPrefix:"https://console" domainSuffix:"/organization-settings/billing/subscription_canceled"`
	// StripeWebhookEvents is the list of events to register when creating a webhook endpoint
	StripeWebhookEvents []string `json:"stripeWebhookEvents" koanf:"stripeWebhookEvents"`
}

type ConfigOpts func(*Config)

// WithEnabled sets the enabled field
func WithEnabled(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.Enabled = enabled
	}
}

// WithPublicStripeKey sets the public stripe key
func WithPublicStripeKey(publicStripeKey string) ConfigOpts {
	return func(c *Config) {
		c.PublicStripeKey = publicStripeKey
	}
}

// WithPrivateStripeKey sets the private stripe key
func WithPrivateStripeKey(privateStripeKey string) ConfigOpts {
	return func(c *Config) {
		c.PrivateStripeKey = privateStripeKey
	}
}

// WithStripeWebhookSecret sets the stripe webhook secret
func WithStripeWebhookSecret(stripeWebhookSecret string) ConfigOpts {
	return func(c *Config) {
		c.StripeWebhookSecret = stripeWebhookSecret
	}
}

// WithStripeWebhookURL sets the stripe webhook URL
func WithStripeWebhookURL(stripeWebhookURL string) ConfigOpts {
	return func(c *Config) {
		c.StripeWebhookURL = stripeWebhookURL
	}
}

// WithStripeBillingPortalSuccessURL sets the stripe billing portal success URL
func WithStripeBillingPortalSuccessURL(stripeBillingPortalSuccessURL string) ConfigOpts {
	return func(c *Config) {
		c.StripeBillingPortalSuccessURL = stripeBillingPortalSuccessURL
	}
}

// WithStripeCancellationReturnURL sets the stripe cancellation return URL
func WithStripeCancellationReturnURL(stripeCancellationReturnURL string) ConfigOpts {
	return func(c *Config) {
		c.StripeCancellationReturnURL = stripeCancellationReturnURL
	}
}

// WithStripeWebhookEvents sets the stripe webhook events
func WithStripeWebhookEvents(events []string) ConfigOpts {
	return func(c *Config) {
		c.StripeWebhookEvents = events
	}
}

// NewConfig creates a new entitlements config
func NewConfig(opts ...ConfigOpts) *Config {
	c := &Config{}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// IsEnabled checks if the entitlements feature is enabled based on the status of the Stripe client settings
func (c *Config) IsEnabled() bool {
	return c.Enabled
}
