package onedrive

import (
	"context"
	"net/mail"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// resolveInstallationMetadata derives OneDrive installation metadata from the persisted access token
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, _, err := oneDriveCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("onedrive: error decoding credentials")

		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return InstallationMetadata{}, false, nil
	}

	token, _, err := new(jwt.Parser).ParseUnverified(cred.AccessToken, jwt.MapClaims{})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("onedrive: error parsing JWT")

		return InstallationMetadata{}, false, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return InstallationMetadata{}, false, nil
	}

	tenantID, ok := claims["tid"].(string)
	if !ok || tenantID == "" {
		return InstallationMetadata{}, false, nil
	}

	upn, ok := claims["upn"].(string)
	if !ok || upn == "" {
		upn, _ = claims["preferred_username"].(string)
	}

	return InstallationMetadata{
		TenantID: tenantID,
		Domain:   domainFromEmail(upn),
	}, true, nil
}

// domainFromEmail extracts the domain portion from an email address
func domainFromEmail(email string) string {
	if email == "" {
		return ""
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return ""
	}

	_, domain, ok := strings.Cut(addr.Address, "@")
	if !ok || domain == "" {
		return ""
	}

	return domain
}
