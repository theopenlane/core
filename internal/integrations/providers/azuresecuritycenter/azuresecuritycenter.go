package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
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
