package oidclocal

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Handle adapts the OIDC health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return func(_ context.Context, request types.OperationRequest) (json.RawMessage, error) {
		return h.Run(request.Credentials)
	}
}

// Run validates the stored OIDC credential and returns a small identity summary
func (HealthCheck) Run(bindings types.CredentialBindings) (json.RawMessage, error) {
	cred, ok, err := oidcCredential.Resolve(bindings)
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
