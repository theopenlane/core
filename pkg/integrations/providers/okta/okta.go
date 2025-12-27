package okta

import (
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/pkg/integrations/providers"
	"github.com/theopenlane/core/pkg/integrations/providers/apikey"
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
