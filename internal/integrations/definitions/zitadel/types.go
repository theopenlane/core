package zitadel

import (
	"github.com/zitadel/zitadel-go/v3/pkg/client"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Zitadel integration definition
	definitionID = types.NewDefinitionRef("def_01K0ZITADEL000000000000001")
	// integration is the typed installation metadata handle for the Zitadel definition
	integration = types.NewInstallationRef(resolveInstallationMetadata)
	// zitadelPATCredential is the typed runtime ref for resolving the PAT credential
	zitadelPATCredential = types.CredentialRefOf[CredentialSchema]()
	// zitadelOAuthCredential is the typed runtime ref for resolving the OAuth credential
	zitadelOAuthCredential = types.CredentialRefOf[OAuthCredentialSchema]()
	// zitadelClient is the client ref for the Zitadel unified API client
	zitadelClient = types.ClientRefOf[*client.Client]().Using(zitadelPATCredential).Using(zitadelOAuthCredential)
	// zitadelPATConnection is the connection mode selected by the PAT credential slot
	zitadelPATConnection = types.NewConnectionRef(zitadelPATCredential).Enables(zitadelClient)
	// zitadelOAuthConnection is the connection mode selected by the OAuth client-credentials slot
	zitadelOAuthConnection = types.NewConnectionRef(zitadelOAuthCredential).Enables(zitadelClient)
	// userInput is the installation user input layout, replacing the flat v1 layout
	userInput = types.NewUserInputRef[UserInput]("zitadel").Replacing(types.NewUserInputRef[oldUserInput]("zitadel-v1"), func(old oldUserInput) UserInput {
		return UserInput{PrimaryDirectory: old.PrimaryDirectory, DirectorySync: DirectorySync{FilterExpr: old.FilterExpr}}
	})
)

// CredentialSchema holds the Zitadel instance credentials for one installation
type CredentialSchema struct {
	// Domain is the Zitadel instance domain (e.g. my-instance.zitadel.cloud)
	Domain string `json:"domain" jsonschema:"required,title=Domain,description=Zitadel instance domain (e.g. my-instance.zitadel.cloud). Uses TLS by default; prefix with http:// for a non-TLS self-hosted instance."`
	// Token is the Zitadel Personal Access Token
	Token string `json:"token" jsonschema:"required,title=Personal Access Token"`
}

// OAuthCredentialSchema holds the Zitadel OAuth2 client-credentials for one installation
type OAuthCredentialSchema struct {
	// Domain is the Zitadel instance domain (e.g. my-instance.zitadel.cloud)
	Domain string `json:"domain" jsonschema:"required,title=Domain,description=Zitadel instance domain (e.g. my-instance.zitadel.cloud). Uses TLS by default; prefix with http:// for a non-TLS self-hosted instance."`
	// ClientID is the Zitadel service user client ID used for the client-credentials grant
	ClientID string `json:"clientId" jsonschema:"required,title=Client ID"`
	// ClientSecret is the Zitadel service user client secret used for the client-credentials grant
	ClientSecret string `json:"clientSecret" jsonschema:"required,title=Client Secret"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// PrimaryDirectory marks this installation as the authoritative source for identity holder sync
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory,description=Mark this as the authoritative source for identity holder enrichment and lifecycle"`
	// DirectorySync configures the directory sync operation
	DirectorySync DirectorySync `json:"directorySync,omitempty" jsonschema:"title=Directory Sync"`
}

// DirectorySync configures collection of Zitadel directory users
type DirectorySync struct {
	// Switch turns the directory sync off for the installation
	types.Switch
	// FilterExpr limits imported records to envelopes matching a CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting"`
}

// oldUserInput is the flat v1 installation user input layout
type oldUserInput struct {
	// PrimaryDirectory marks this installation as the authoritative source for identity holder sync
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory,description=Mark this as the authoritative source for identity holder enrichment and lifecycle"`
	// FilterExpr limits imported records to envelopes matching a CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting"`
}

// InstallationMetadata holds the stable Zitadel instance identity for one installation
type InstallationMetadata struct {
	// Domain is the Zitadel instance domain configured for this installation
	Domain string `json:"domain,omitempty"`
	// InstanceID is the immutable id of the Zitadel instance the credential is scoped to
	InstanceID string `json:"instanceId,omitempty"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID:   m.InstanceID,
		ExternalName: m.Domain,
	}
}
