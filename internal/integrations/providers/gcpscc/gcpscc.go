package gcpscc

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeGCPSCC identifies the GCP Security Command Center provider
const TypeGCPSCC = types.ProviderType("gcp_scc")

// Builder returns the GCP SCC provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeGCPSCC)
}
