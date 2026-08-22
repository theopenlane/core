package fossa

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// resolveInstallationMetadata derives the FOSSA organization identity from the bound credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	if _, bound := req.Credentials.Resolve(fossaCredential.ID()); !bound {
		return InstallationMetadata{}, true, nil
	}

	cred, ok, err := fossaCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("fossa: error resolving credential")

		return InstallationMetadata{}, false, ErrCredentialMetadataInvalid
	}

	if !ok {
		return InstallationMetadata{}, ok, nil
	}

	baseURL := baseURLOrDefault(cred.BaseURL)

	client, err := newAPIClient(baseURL, cred.APIToken)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	org, err := client.organization(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("fossa: error fetching organization details")

		return InstallationMetadata{}, false, ErrOrganizationFetchFailed
	}

	return InstallationMetadata{
		OrganizationID: org.identifier(),
		BaseURL:        baseURL,
		Subscription:   org.Subscription,
	}, true, nil
}
