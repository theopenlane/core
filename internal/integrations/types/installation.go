package types //nolint:revive

import openapi "github.com/theopenlane/core/common/openapi"

// IntegrationConfig is the per-installation runtime configuration
type IntegrationConfig = openapi.IntegrationConfig

// IntegrationInstallationMetadata stores stable installation identity metadata
type IntegrationInstallationMetadata = openapi.IntegrationInstallationMetadata

// IntegrationInstallationIdentity is the normalized, provider-agnostic installation identity
type IntegrationInstallationIdentity = openapi.IntegrationInstallationIdentity

// IntegrationProviderState stores provider-specific state captured during auth and config
type IntegrationProviderState = openapi.IntegrationProviderState

// InstallationIdentifiable producs normalized display identity for the UI
type InstallationIdentifiable interface {
	// InstallationIdentity returns the normalized identity fields for UI display
	InstallationIdentity() IntegrationInstallationIdentity
}
