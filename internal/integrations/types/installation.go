package types

import openapi "github.com/theopenlane/core/common/openapi"

// IntegrationConfig is the per-installation runtime configuration
type IntegrationConfig = openapi.IntegrationConfig

// IntegrationInstallationMetadata stores stable installation identity metadata
type IntegrationInstallationMetadata = openapi.IntegrationInstallationMetadata

// IntegrationProviderState stores provider-specific state captured during auth and config
type IntegrationProviderState = openapi.IntegrationProviderState
