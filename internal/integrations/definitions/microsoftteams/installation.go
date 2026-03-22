package microsoftteams

import (
	"context"

	"github.com/golang-jwt/jwt/v5"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// resolveInstallationMetadata derives Microsoft tenant metadata from the persisted access token when available
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	var cred teamsCred
	if err := jsonx.UnmarshalIfPresent(req.Credential.Data, &cred); err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return InstallationMetadata{}, false, nil
	}

	token, _, err := new(jwt.Parser).ParseUnverified(cred.AccessToken, jwt.MapClaims{})
	if err != nil {
		return InstallationMetadata{}, false, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return InstallationMetadata{}, false, nil
	}

	tenantID, _ := claims["tid"].(string)
	if tenantID == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		TenantID: tenantID,
	}, true, nil
}
