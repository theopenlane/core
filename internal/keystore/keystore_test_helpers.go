package keystore

import (
	"context"

	"github.com/theopenlane/shared/integrations/types"
)

// credentialSourceStub implements CredentialSource for tests
type credentialSourceStub struct { //nolint:unused
	getPayload  types.CredentialPayload
	mintPayload types.CredentialPayload
	getErr      error
	mintErr     error

	getCount  int
	mintCount int

	lastGetOrgID     string
	lastGetProvider  types.ProviderType
	lastMintOrgID    string
	lastMintProvider types.ProviderType
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
