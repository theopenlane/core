package vercel

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeVercel identifies the Vercel provider
const TypeVercel = types.ProviderType("vercel")

// Builder returns the Vercel provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeVercel)
}
