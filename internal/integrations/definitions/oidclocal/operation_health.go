package oidclocal

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// checkHealth validates the stored OIDC credential and returns a small identity summary
func checkHealth(_ context.Context, req types.OperationRequest) (json.RawMessage, error) {
	cred, ok, err := oidcCredential.Resolve(req.Credentials)
	if err != nil {
		return nil, ErrCredentialDecode
	}

	if !ok || cred.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	if cred.Subject == "" {
		return nil, ErrSubjectMissing
	}

	return providerkit.EncodeResult(HealthCheck{
		Issuer:            cred.Issuer,
		Subject:           cred.Subject,
		Email:             cred.Email,
		PreferredUsername: cred.PreferredUsername,
		Groups:            cred.Groups,
	}, ErrResultEncode)
}
