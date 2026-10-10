package oidclocal

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verify validates the stored OIDC credential and returns the identity it was issued for
func verify(_ context.Context, _ types.ConnectionRequest[oidcLocalCred], c Client) (InstallationMetadata, error) {
	switch {
	case c.Credential.AccessToken == "":
		return InstallationMetadata{}, ErrOAuthTokenMissing
	case c.Credential.Subject == "":
		return InstallationMetadata{}, ErrSubjectMissing
	}

	return InstallationMetadata{
		Issuer:            c.Credential.Issuer,
		Subject:           c.Credential.Subject,
		Email:             c.Credential.Email,
		PreferredUsername: c.Credential.PreferredUsername,
	}, nil
}
