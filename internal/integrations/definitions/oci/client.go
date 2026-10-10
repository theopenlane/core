package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/cloudguard"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// buildClient constructs the OCI Identity and Cloud Guard clients with the tenancy scope for one installation
func buildClient(_ context.Context, req types.ConnectionRequest[CredentialSchema]) (Client, error) {
	provider, err := buildConfigurationProvider(req.Credential)
	if err != nil {
		return Client{}, err
	}

	identityClient, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	if err != nil {
		return Client{}, ErrIdentityClientCreate
	}

	cloudGuardClient, err := cloudguard.NewCloudGuardClientWithConfigurationProvider(provider)
	if err != nil {
		return Client{}, ErrCloudGuardClientCreate
	}

	return Client{
		Identity:        &identityClient,
		CloudGuard:      &cloudGuardClient,
		TenancyOCID:     req.Credential.TenancyOCID,
		CompartmentOCID: req.Credential.CompartmentOCID,
		Region:          req.Credential.Region,
	}, nil
}

// buildConfigurationProvider builds an OCI request signing configuration
func buildConfigurationProvider(cred CredentialSchema) (common.ConfigurationProvider, error) {
	provider := common.NewRawConfigurationProvider(
		cred.TenancyOCID,
		cred.UserOCID,
		cred.Region,
		cred.Fingerprint,
		cred.PrivateKey,
		lo.EmptyableToPtr(cred.PrivateKeyPassphrase),
	)

	if _, err := common.IsConfigurationProviderValid(provider); err != nil {
		return nil, ErrConfigurationProviderInvalid
	}

	return provider, nil
}
