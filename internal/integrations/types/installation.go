package types //nolint:revive

import (
	"time"

	"github.com/theopenlane/core/common/enums"
	openapi "github.com/theopenlane/core/common/openapi"
)

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

// Installation is the ent-free view of one integration installation record handed to definition code.
// It is a point-in-time snapshot taken when the runtime builds a request; definitions read from it and
// persist changes through the runtime rather than mutating it
type Installation struct {
	// ID is the installation identifier
	ID string `json:"id"`
	// OwnerID is the organization that owns the installation
	OwnerID string `json:"ownerId,omitempty"`
	// Name is the user-facing installation name
	Name string `json:"name,omitempty"`
	// Description is the user-facing installation description
	Description string `json:"description,omitempty"`
	// DefinitionID is the canonical identifier of the installed definition
	DefinitionID string `json:"definitionId,omitempty"`
	// DefinitionVersion is the definition surface version the installation was last written under
	DefinitionVersion string `json:"definitionVersion,omitempty"`
	// Family is the definition family label recorded on the installation
	Family string `json:"family,omitempty"`
	// Status is the installation lifecycle status
	Status enums.IntegrationStatus `json:"status,omitempty"`
	// SystemOwned reports whether the installation is owned by the platform rather than a customer
	SystemOwned bool `json:"systemOwned,omitempty"`
	// PrimaryDirectory reports whether the installation is its organization's authoritative directory source
	PrimaryDirectory bool `json:"primaryDirectory,omitempty"`
	// CampaignEmail reports whether the installation is its organization's campaign email sender
	CampaignEmail bool `json:"campaignEmail,omitempty"`
	// Config is the installation configuration document
	Config openapi.IntegrationConfig `json:"config,omitempty"`
	// UserInput is the stored installation-scoped user input
	UserInput IntegrationUserInput `json:"userInput,omitempty"`
	// OperationConfig is the stored per-operation input keyed by operation name
	OperationConfig IntegrationOperationConfig `json:"operationConfig,omitempty"`
	// InstallationMetadata is the stable external identity metadata for the installation
	InstallationMetadata IntegrationInstallationMetadata `json:"installationMetadata,omitempty"`
	// ProviderState is the provider-specific state captured during auth and config
	ProviderState IntegrationProviderState `json:"providerState,omitempty"`
	// Metadata is the free-form installation metadata map
	Metadata map[string]any `json:"metadata,omitempty"`
	// ExpiresAt is when a never-connected installation expires, nil when it does not
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// CreatedAt is when the installation was created
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt is when the installation was last updated
	UpdatedAt time.Time `json:"updatedAt"`
}

// WebhookEndpoint is the ent-free view of one persisted webhook endpoint for an installation
type WebhookEndpoint struct {
	// ID is the webhook record identifier
	ID string `json:"id"`
	// OwnerID is the organization that owns the webhook
	OwnerID string `json:"ownerId,omitempty"`
	// IntegrationID is the installation the webhook belongs to
	IntegrationID string `json:"integrationId,omitempty"`
	// Name is the webhook contract name within the definition
	Name string `json:"name,omitempty"`
	// Status is the webhook lifecycle status
	Status enums.IntegrationWebhookStatus `json:"status,omitempty"`
	// EndpointID is the external endpoint identifier assigned by the provider, when any
	EndpointID *string `json:"endpointId,omitempty"`
	// EndpointURL is the inbound URL the provider delivers to, when persisted
	EndpointURL *string `json:"endpointUrl,omitempty"`
	// SecretToken is the shared verification secret for inbound deliveries
	SecretToken string `json:"-"`
	// AllowedEvents lists the provider events accepted on the endpoint
	AllowedEvents []string `json:"allowedEvents,omitempty"`
	// ExternalEventID is the provider-side subscription identifier, when any
	ExternalEventID *string `json:"externalEventId,omitempty"`
	// LastDeliveryID is the most recent processed delivery identifier
	LastDeliveryID string `json:"lastDeliveryId,omitempty"`
	// Metadata is the free-form webhook metadata map
	Metadata map[string]any `json:"metadata,omitempty"`
}
