package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
<<<<<<< HEAD:internal/integrations/providers/azuresecuritycenter/azuresecuritycenter.go
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
=======
	"github.com/theopenlane/core/pkg/integrations/providers"
>>>>>>> 11b940e6b (add new integration descriptors, client initialization, swing at health checks per provider):pkg/integrations/providers/azuresecuritycenter/azuresecuritycenter.go
)

// TypeAzureSecurityCenter identifies the Azure Security Center provider.
const TypeAzureSecurityCenter = types.ProviderType("azure_security_center")

// Builder returns the Azure Security Center provider builder.
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeAzureSecurityCenter,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			if spec.AuthType != "" && spec.AuthType != types.AuthKindOAuth2 {
				return nil, ErrAuthTypeMismatch
			}

			return newProvider(spec), nil
		},
	}
}
