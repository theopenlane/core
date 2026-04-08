package oidclocal

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives OIDC installation metadata from the stored auth-managed credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, ok, err := oidcCredential.Resolve(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if !ok {
		return InstallationMetadata{}, false, nil
	}

	if cred.Subject == "" && cred.Email == "" && cred.PreferredUsername == "" && cred.Issuer == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		Issuer:            cred.Issuer,
		Subject:           cred.Subject,
		Email:             cred.Email,
		PreferredUsername: cred.PreferredUsername,
	}, true, nil
}
