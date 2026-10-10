package keycloak

import (
	"context"

	gocloak "github.com/Nerzal/gocloak/v13"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Keycloak integration definition
	definitionID = types.NewDefinitionRef("def_01K0KEYCLOAK000000000000001")
	// clientCredentials is the connection handle for the Keycloak client credentials
	clientCredentials = types.ConnectionOf[CredentialSchema]()
	// installation is the typed installation metadata handle for the Keycloak definition
	installation = types.InstallationOf[InstallationMetadata]()
	// userInput is the installation user input layout
	userInput = types.UserInputRefOf[UserInput]()
)

// Client is the Keycloak API client with the realm scope and client credentials its calls need
type Client struct {
	// GoCloak is the underlying Keycloak API client
	*gocloak.GoCloak
	// Realm is the Keycloak realm the client is scoped to
	Realm string
	// ClientID is the Keycloak client ID used to log in
	ClientID string
	// ClientSecret is the Keycloak client secret used to log in
	ClientSecret string
}

// ClientToken acquires an access token for the client credentials in the realm
func (c Client) ClientToken(ctx context.Context) (*gocloak.JWT, error) {
	return c.LoginClient(ctx, c.ClientID, c.ClientSecret, c.Realm)
}

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
