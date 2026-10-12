package types //nolint:revive

import openapi "github.com/theopenlane/core/common/openapi"

// IntegrationUserInput stores the installation-scoped user input under its layout name
type IntegrationUserInput = openapi.IntegrationUserInput

// IntegrationOperationConfig stores the installation-scoped operation input keyed by operation name
type IntegrationOperationConfig = openapi.IntegrationOperationConfig

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
