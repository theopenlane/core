package keystore

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// credentialSourceStub implements CredentialSource for tests
type credentialSourceStub struct { //nolint:unused
	getPayload                types.CredentialPayload
	mintPayload               types.CredentialPayload
	getErr                    error
	mintErr                   error
	getForIntegrationPayload  types.CredentialPayload
	mintForIntegrationPayload types.CredentialPayload
	getForIntegrationErr      error
	mintForIntegrationErr     error

	getCount                int
	mintCount               int
	getForIntegrationCount  int
	mintForIntegrationCount int

	lastGetOrgID          string
	lastGetProvider       types.ProviderType
	lastMintOrgID         string
	lastMintProvider      types.ProviderType
	lastGetIntegrationID  string
	lastMintIntegrationID string
}

func (s *credentialSourceStub) Get(_ context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error) { //nolint:unused
	s.getCount++
	s.lastGetOrgID = orgID
	s.lastGetProvider = provider

	if s.getErr != nil {
		return types.CredentialPayload{}, s.getErr
	}

	payload := s.getPayload
	if payload.Provider == types.ProviderUnknown {
		payload.Provider = provider
	}

	return payload, nil
}

func (s *credentialSourceStub) Mint(_ context.Context, orgID string, provider types.ProviderType) (types.CredentialPayload, error) { //nolint:unused
	s.mintCount++
	s.lastMintOrgID = orgID
	s.lastMintProvider = provider

	if s.mintErr != nil {
		return types.CredentialPayload{}, s.mintErr
	}

	payload := s.mintPayload
	if payload.Provider == types.ProviderUnknown {
		payload.Provider = provider
	}

	return payload, nil
}

func (s *credentialSourceStub) GetForIntegration(_ context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialPayload, error) { //nolint:unused
	s.getForIntegrationCount++
	s.lastGetOrgID = orgID
	s.lastGetProvider = provider
	s.lastGetIntegrationID = integrationID

	if s.getForIntegrationErr != nil {
		return types.CredentialPayload{}, s.getForIntegrationErr
	}

	payload := s.getForIntegrationPayload
	if payload.Provider == types.ProviderUnknown {
		payload = s.getPayload
	}
	if payload.Provider == types.ProviderUnknown {
		payload.Provider = provider
	}

	return payload, nil
}

func (s *credentialSourceStub) MintForIntegration(_ context.Context, orgID string, provider types.ProviderType, integrationID string) (types.CredentialPayload, error) { //nolint:unused
	s.mintForIntegrationCount++
	s.lastMintOrgID = orgID
	s.lastMintProvider = provider
	s.lastMintIntegrationID = integrationID

	if s.mintForIntegrationErr != nil {
		return types.CredentialPayload{}, s.mintForIntegrationErr
	}

	payload := s.mintForIntegrationPayload
	if payload.Provider == types.ProviderUnknown {
		payload = s.mintPayload
	}
	if payload.Provider == types.ProviderUnknown {
		payload.Provider = provider
	}

	return payload, nil
}
