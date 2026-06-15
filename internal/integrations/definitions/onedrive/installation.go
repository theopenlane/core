package onedrive

import (
	"context"

	"github.com/golang-jwt/jwt/v5"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives OneDrive installation metadata from the persisted access token
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, _, err := oneDriveCredential.Resolve(req.Credentials)
	if err != nil {
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

	upn, _ := claims["upn"].(string)
	if upn == "" {
		upn, _ = claims["preferred_username"].(string)
	}

	return InstallationMetadata{
		TenantID: tenantID,
		Domain:   domainFromEmail(upn),
	}, true, nil
}

// domainFromEmail extracts the domain portion from an email address
func domainFromEmail(email string) string {
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			return email[i+1:]
		}
	}

	return email
}
