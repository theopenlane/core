package cloudflare

import (
	"github.com/theopenlane/common/integrations/types"
	"github.com/theopenlane/core/pkg/integrations/providers"
	"github.com/theopenlane/core/pkg/integrations/providers/apikey"
)

// TypeCloudflare identifies the Cloudflare provider
const TypeCloudflare = types.ProviderType("cloudflare")

// Builder returns the Cloudflare provider builder
func Builder() providers.Builder {
	return apikey.Builder(
		TypeCloudflare,
		apikey.WithTokenField("apiToken"),
		apikey.WithOperations(cloudflareOperations()),
	)
}
