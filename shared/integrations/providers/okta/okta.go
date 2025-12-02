package okta

import (
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/apikey"
	"github.com/theopenlane/shared/integrations/types"
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
