package entitlements

type Config struct {
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// PublicStripeKey is the key for the stripe service
	PublicStripeKey string `json:"publicStripeKey" koanf:"publicStripeKey" default:""`
	// PrivateStripeKey is the key for the stripe service
	PrivateStripeKey string `json:"privateStripeKey" koanf:"privateStripeKey" default:""`
	// StripeWebhookSecret is the secret for the stripe service
	StripeWebhookSecret string `json:"stripeWebhookSecret" koanf:"stripeWebhookSecret" default:""`
}
