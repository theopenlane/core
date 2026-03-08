package keystore

import (
	"context"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/types"
)

// credentialSourceStub implements CredentialSource for tests
type credentialSourceStub struct { //nolint:unused
	getPayload                models.CredentialSet
	mintPayload               models.CredentialSet
	getErr                    error
	mintErr                   error
	getForIntegrationPayload  models.CredentialSet
	mintForIntegrationPayload models.CredentialSet
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

func (s *credentialSourceStub) Get(_ context.Context, orgID string, provider types.ProviderType) (models.CredentialSet, error) { //nolint:unused
	s.getCount++
	s.lastGetOrgID = orgID
	s.lastGetProvider = provider

	if s.getErr != nil {
		return models.CredentialSet{}, s.getErr
	}

	return types.CloneCredentialSet(s.getPayload), nil
}

func (s *credentialSourceStub) Mint(_ context.Context, orgID string, provider types.ProviderType) (models.CredentialSet, error) { //nolint:unused
	s.mintCount++
	s.lastMintOrgID = orgID
	s.lastMintProvider = provider

	if s.mintErr != nil {
		return models.CredentialSet{}, s.mintErr
	}

	return types.CloneCredentialSet(s.mintPayload), nil
}

func (s *credentialSourceStub) GetForIntegration(_ context.Context, orgID string, provider types.ProviderType, integrationID string) (models.CredentialSet, error) { //nolint:unused
	s.getForIntegrationCount++
	s.lastGetOrgID = orgID
	s.lastGetProvider = provider
	s.lastGetIntegrationID = integrationID

	if s.getForIntegrationErr != nil {
		return models.CredentialSet{}, s.getForIntegrationErr
	}

	payload := s.getForIntegrationPayload
	if types.IsCredentialSetEmpty(payload) {
		payload = s.getPayload
	}

	return types.CloneCredentialSet(payload), nil
}

func (s *credentialSourceStub) MintForIntegration(_ context.Context, orgID string, provider types.ProviderType, integrationID string) (models.CredentialSet, error) { //nolint:unused
	s.mintForIntegrationCount++
	s.lastMintOrgID = orgID
	s.lastMintProvider = provider
	s.lastMintIntegrationID = integrationID

	if s.mintForIntegrationErr != nil {
		return models.CredentialSet{}, s.mintForIntegrationErr
	}

	payload := s.mintForIntegrationPayload
	if types.IsCredentialSetEmpty(payload) {
		payload = s.mintPayload
	}

	return types.CloneCredentialSet(payload), nil
}
