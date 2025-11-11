package okta

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeOkta identifies the Okta provider
const TypeOkta = types.ProviderType("okta")

// Builder returns the Okta provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeOkta)
}
