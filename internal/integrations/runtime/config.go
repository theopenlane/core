package runtime

import (
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/keymaker"
)

// Config carries all dependencies and settings for constructing the integrations runtime.
// Builders provides the set of provider builders used to construct the registry; when nil
// the catalog defaults are used. Registry may be provided directly to bypass builder
// construction entirely; intended for tests.
type Config struct {
	// Builders is the list of provider builders used to initialize the registry.
	// When nil, catalog.Builders with default (empty) operator config is used.
	// Ignored when Registry is provided directly.
	Builders []providers.Builder
	// Registry optionally provides a pre-built registry, bypassing builder construction.
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
