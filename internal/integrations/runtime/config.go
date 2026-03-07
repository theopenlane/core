package runtime

import (
	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationconfig "github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/keymaker"
)

// Config carries all dependencies and settings for constructing the integrations runtime.
// The registry is built internally from ProviderSpecs and GitHubApp configuration.
// Registry may be provided directly to override the internal build; intended for tests.
type Config struct {
	// ProviderSpecs contains the declarative provider configurations keyed by provider name.
	// Ignored when Registry is provided directly.
	ProviderSpecs map[string]integrationconfig.ProviderSpec
	// Registry optionally provides a pre-built registry, bypassing ProviderSpecs construction.
	// Intended for test use only.
	Registry *registry.Registry
	// DB provides persistence for credentials and integration records.
	DB *ent.Client
	// AuthStateStore optionally overrides the in-memory OAuth activation session store.
	AuthStateStore keymaker.AuthStateStore
	// GitHubApp holds GitHub App integration configuration.
	GitHubApp GitHubAppConfig
	// OAuth holds OAuth provider integration configuration.
	OAuth OAuthConfig
}

// GitHubAppConfig holds configuration for the GitHub App integration.
type GitHubAppConfig struct {
	// Enabled toggles the GitHub App integration handlers.
	Enabled bool
	// AppID is the GitHub App ID used for JWT signing.
	AppID string
	// AppSlug is the GitHub App slug used for the install URL.
	AppSlug string
	// PrivateKey is the PEM-encoded GitHub App private key.
	PrivateKey string
	// WebhookSecret is the shared secret used to validate GitHub webhooks.
	WebhookSecret string
	// SuccessRedirectURL is the URL to redirect to after successful installation.
	SuccessRedirectURL string
}

// OAuthConfig holds configuration for integration OAuth providers.
type OAuthConfig struct {
	// Enabled toggles initialization of the integration provider registry.
	Enabled bool
	// SuccessRedirectURL is the URL to redirect to after successful OAuth integration.
	SuccessRedirectURL string
}
