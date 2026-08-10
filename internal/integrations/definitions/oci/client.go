package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/cloudguard"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
)

// IdentityClientBuilder builds OCI Identity clients for one installation
type IdentityClientBuilder struct{}

// Build constructs the OCI Identity client for one installation
func (IdentityClientBuilder) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	provider, err := buildConfigurationProvider(req.Credentials)
	if err != nil {
		return nil, err
	}

	client, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, ErrIdentityClientCreate
	}

	return &client, nil
}

// CloudGuardClientBuilder builds OCI Cloud Guard clients for one installation
type CloudGuardClientBuilder struct{}

// Build constructs the OCI Cloud Guard client for one installation
func (CloudGuardClientBuilder) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	provider, err := buildConfigurationProvider(req.Credentials)
	if err != nil {
		return nil, err
	}

	client, err := cloudguard.NewCloudGuardClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, ErrCloudGuardClientCreate
	}

	return &client, nil
}

// buildConfigurationProvider resolves the credential and turns it into an OCI request signing configuration
func buildConfigurationProvider(bindings types.CredentialBindings) (common.ConfigurationProvider, error) {
	cred, err := resolveCredential(bindings)
	if err != nil {
		return nil, err
	}

	provider := common.NewRawConfigurationProvider(
		cred.TenancyOCID,
		cred.UserOCID,
		cred.Region,
		cred.Fingerprint,
		cred.PrivateKey,
		lo.EmptyableToPtr(cred.PrivateKeyPassphrase),
	)

	// catches an unparsable PEM key, a bad passphrase, or an unknown region before the first API call
	if _, err := common.IsConfigurationProviderValid(provider); err != nil {
		return nil, ErrConfigurationProviderInvalid
	}

	return provider, nil
}

// resolveCredential decodes OCI credential metadata from the credential bindings
func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := ociCredential.Resolve(bindings)
	if err != nil {
		return CredentialSchema{}, ErrMetadataDecode
	}

	if !ok {
		return CredentialSchema{}, ErrCredentialMetadataRequired
	}

	return cred, nil
}

// resolveCompartment returns the compartment collection is rooted at, defaulting to the tenancy
func resolveCompartment(meta CredentialSchema) string {
	if meta.CompartmentOCID != "" {
		return meta.CompartmentOCID
	}

	return meta.TenancyOCID
}
