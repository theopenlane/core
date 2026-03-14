package providerkit

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ValidateAPIKeyCredential returns a CredentialBuilderFunc that validates the apiToken field
// is present and non-empty in the credential's ProviderData.
func ValidateAPIKeyCredential() types.CredentialBuilderFunc {
	return func(_ context.Context, _ *generated.Integration, credential types.CredentialSet) (types.CredentialSet, error) {
		if _, err := APITokenFromCredential(credential); err != nil {
			return types.CredentialSet{}, err
		}

		return credential, nil
	}
}
