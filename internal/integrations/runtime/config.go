package runtime

import (
	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationconfig "github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/keymaker"
)

// Config carries all dependencies and settings for constructing the integrations runtime.
// Provider credentials and settings are expressed through ProviderSpecs, which can be
// overridden at deploy time via the integrationproviders config key (same mechanism used
// for all providers, including OAuth clientId/clientSecret).
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
	// SuccessRedirectURL is the global fallback URL to redirect to after successful provider
	// authentication. Per-provider overrides take precedence when set in the provider spec.
	SuccessRedirectURL string
}
