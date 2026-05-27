package keycloak

import (
	gocloak "github.com/Nerzal/gocloak/v13"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Keycloak integration definition
	definitionID = types.NewDefinitionRef("def_01K0KEYCLOAK000000000000001")
	// integration is the typed installation metadata handle for the Keycloak definition
	integration = types.NewInstallationRef(resolveInstallationMetadata)
	// keycloakCredentialSchema is the JSON schema for the Keycloak credential
	// keycloakCredential is the typed runtime ref for resolving the credential
	keycloakCredentialSchema, keycloakCredential = providerkit.CredentialSchema[CredentialSchema]()
	// keycloakClient is the client ref for the Keycloak API client
	keycloakClient = types.NewClientRef[*gocloak.GoCloak]()
	// healthCheckSchema, healthCheckOperation is the operation ref for the health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema, directorySyncOperation is the operation ref for directory sync
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// CredentialSchema holds the Keycloak instance credentials for one installation
type CredentialSchema struct {
	// BaseURL is the base URL of the Keycloak instance
	BaseURL string `json:"baseUrl" jsonschema:"required,title=Base URL"`
	// Realm is the Keycloak realm to sync
	Realm string `json:"realm" jsonschema:"required,title=Realm"`
	// ClientID is the Keycloak client ID
	ClientID string `json:"clientId" jsonschema:"required,title=Client ID"`
	// ClientSecret is the Keycloak client secret
	ClientSecret string `json:"clientSecret" jsonschema:"required,title=Client Secret"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// PrimaryDirectory marks this installation as the authoritative source for identity holder sync
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory,description=Mark this as the authoritative source for identity holder enrichment and lifecycle"`
	// DisableGroupSync when true only syncs users, skipping groups and memberships
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users disable group and membership sync operations"`
	// FilterExpr limits imported records to envelopes matching a CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting, example=Example: payload.type == 'internal'"`
}

// InstallationMetadata holds the stable Keycloak realm identity for one installation
type InstallationMetadata struct {
	// RealmID is the stable UUID of the Keycloak realm
	RealmID string `json:"realmId,omitempty"`
	// RealmName is the realm name of the Keycloak instance
	RealmName string `json:"realmName,omitempty"`
	// DisplayName is the human readable display name of the realm
	DisplayName string `json:"displayName,omitempty"`
	// KeycloakVersion is the version of the Keycloak instance
	KeycloakVersion string `json:"keycloakVersion,omitempty"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	name := m.DisplayName
	if name == "" {
		name = m.RealmName
	}

	return types.IntegrationInstallationIdentity{
		ExternalName: name,
		ExternalID:   m.RealmID,
	}
}


