package entitlements

type Config struct {
	// Enabled determines if the entitlements service is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// PublicStripeKey is the key for the stripe service
	PublicStripeKey string `json:"publicStripeKey" koanf:"publicStripeKey" default:""`
	// PrivateStripeKey is the key for the stripe service
	PrivateStripeKey string `json:"privateStripeKey" koanf:"privateStripeKey" default:""`
	// StripeWebhookSecret is the secret for the stripe service
	StripeWebhookSecret string `json:"stripeWebhookSecret" koanf:"stripeWebhookSecret" default:""`
	// TrialSubscriptionPriceID is the price ID for the trial subscription
	TrialSubscriptionPriceID string `json:"trialSubscriptionPriceID" koanf:"trialSubscriptionPriceID" default:"price_1QKLyeBvxky1R7SvaZYGWyQb"`
	// StripeWebhookURL is the URL for the stripe webhook
	StripeWebhookURL string `json:"stripeWebhookURL" koanf:"stripeWebhookURL" default:"https://api.openlane.com/v1/stripe/webhook"`
	// StripeBillingPortalSuccessURL
	StripeBillingPortalSuccessURL string `json:"stripeBillingPortalSuccessURL" koanf:"stripeBillingPortalSuccessURL" default:"https://console.openlane.com/billing"`
}

type ConfigOpts func(*Config)

func WithEnabled(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.Enabled = enabled
	}
}

func WithPublicStripeKey(publicStripeKey string) ConfigOpts {
	return func(c *Config) {
		c.PublicStripeKey = publicStripeKey
	}
}

func WithPrivateStripeKey(privateStripeKey string) ConfigOpts {
	return func(c *Config) {
		c.PrivateStripeKey = privateStripeKey
	}
}

func WithStripeWebhookSecret(stripeWebhookSecret string) ConfigOpts {
	return func(c *Config) {
		c.StripeWebhookSecret = stripeWebhookSecret
	}
}

func WithTrialSubscriptionPriceID(trialSubscriptionPriceID string) ConfigOpts {
	return func(c *Config) {
		c.TrialSubscriptionPriceID = trialSubscriptionPriceID
	}
}

func WithStripeWebhookURL(stripeWebhookURL string) ConfigOpts {
	return func(c *Config) {
		c.StripeWebhookURL = stripeWebhookURL
	}
}

func WithStripeBillingPortalSuccessURL(stripeBillingPortalSuccessURL string) ConfigOpts {
	return func(c *Config) {
		c.StripeBillingPortalSuccessURL = stripeBillingPortalSuccessURL
	}
}

func NewConfig(opts ...ConfigOpts) *Config {
	c := &Config{}
	for _, opt := range opts {
		opt(c)
	}

	return c
}
