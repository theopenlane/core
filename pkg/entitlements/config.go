package entitlements

import "github.com/theopenlane/core/pkg/enums"

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
	// PersonalOrgSubscriptionPriceID is the price ID for the personal org subscription
	PersonalOrgSubscriptionPriceID string `json:"personalOrgSubscriptionPriceID" koanf:"personalOrgSubscriptionPriceID" default:"price_1QycPyBvxky1R7Svz0gOWnNh"`
	// StripeWebhookURL is the URL for the stripe webhook
	StripeWebhookURL string `json:"stripeWebhookURL" koanf:"stripeWebhookURL" default:"https://api.openlane.com/v1/stripe/webhook"`
	// StripeBillingPortalSuccessURL
	StripeBillingPortalSuccessURL string `json:"stripeBillingPortalSuccessURL" koanf:"stripeBillingPortalSuccessURL" default:"https://console.openlane.com/billing"`
	// StripeCancellationReturnURL is the URL for the stripe cancellation return
	StripeCancellationReturnURL string `json:"stripeCancellationReturnURL" koanf:"stripeCancellationReturnURL" default:"https://console.theopenlane.io/organization-settings/billing/subscription_canceled"`
	// SaaSPricingTiers are the pricing tiers for the SaaS product
	SaaSPricingTiers []PricingTier `json:"saasPricingTiers" koanf:"saasPricingTiers"`
	// Features are the features for the SaaS product
	Features []Feature `json:"features" koanf:"features"`
	// UsageLimits defines default usage caps per resource type
	// that new organizations should start with
	UsageLimits map[enums.UsageType]int64 `json:"usageLimits" koanf:"usageLimits"`
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

// WithTrialSubscriptionPriceID sets the trial subscription price ID
func WithTrialSubscriptionPriceID(trialSubscriptionPriceID string) ConfigOpts {
	return func(c *Config) {
		c.TrialSubscriptionPriceID = trialSubscriptionPriceID
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

// WithPersonalOrgSubscriptionPriceID sets the personal org subscription price ID
func WithPersonalOrgSubscriptionPriceID(personalOrgSubscriptionPriceID string) ConfigOpts {
	return func(c *Config) {
		c.PersonalOrgSubscriptionPriceID = personalOrgSubscriptionPriceID
	}
}

// WithStripeCancellationReturnURL sets the stripe cancellation return URL
func WithStripeCancellationReturnURL(stripeCancellationReturnURL string) ConfigOpts {
	return func(c *Config) {
		c.StripeCancellationReturnURL = stripeCancellationReturnURL
	}
}

// WithUsageLimits sets the default usage limits for new organizations
func WithUsageLimits(limits map[enums.UsageType]int64) ConfigOpts {
	return func(c *Config) {
		c.UsageLimits = limits
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
