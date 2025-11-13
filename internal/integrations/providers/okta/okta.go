package okta

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/apikey"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeOkta identifies the Okta provider
const TypeOkta = types.ProviderType("okta")

// Builder returns the Okta provider builder
func Builder() providers.Builder {
	return apikey.Builder(
		TypeOkta,
		apikey.WithTokenField("apiToken"),
		apikey.WithOperations(oktaOperations()),
	)
}
