package zitadel

import (
	"github.com/zitadel/zitadel-go/v3/pkg/client"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Zitadel integration definition
	definitionID = types.NewDefinitionRef("def_01K0ZITADEL000000000000001")
	// integration is the typed installation metadata handle for the Zitadel definition
	integration = types.NewInstallationRef(resolveInstallationMetadata)
	// zitadelPATCredentialSchema is the JSON schema for the Zitadel PAT credential
	// zitadelPATCredential is the typed runtime ref for resolving the PAT credential
	zitadelPATCredentialSchema, zitadelPATCredential = providerkit.CredentialSchema[CredentialSchema]()
	// zitadelOAuthCredentialSchema is the JSON schema for the Zitadel OAuth2 client-credentials credential
	// zitadelOAuthCredential is the typed runtime ref for resolving the OAuth credential
	zitadelOAuthCredentialSchema, zitadelOAuthCredential = providerkit.CredentialSchema[OAuthCredentialSchema]()
	// zitadelClient is the client ref for the Zitadel unified API client
	zitadelClient = types.NewClientRef[*client.Client]()
	// healthCheckSchema, healthCheckOperation is the operation ref for the health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema, directorySyncOperation is the operation ref for directory sync
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// CredentialSchema holds the Zitadel instance credentials for one installation
type CredentialSchema struct {
	// Domain is the Zitadel instance domain (e.g. my-instance.zitadel.cloud). It connects over
	// TLS by default; prefix with http:// for a self-hosted or local instance without TLS.
	Domain string `json:"domain" jsonschema:"required,title=Domain,description=Zitadel instance domain (e.g. my-instance.zitadel.cloud). Uses TLS by default; prefix with http:// for a non-TLS self-hosted instance."`
	// Token is the Zitadel Personal Access Token
	Token string `json:"token" jsonschema:"required,title=Personal Access Token"`
}

// OAuthCredentialSchema holds the Zitadel OAuth2 client-credentials for one installation
type OAuthCredentialSchema struct {
	// Domain is the Zitadel instance domain (e.g. my-instance.zitadel.cloud). It connects over
	// TLS by default; prefix with http:// for a self-hosted or local instance without TLS.
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
	// FilterExpr limits imported records to envelopes matching a CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting"`
}

// InstallationMetadata holds the stable Zitadel instance identity for one installation
type InstallationMetadata struct {
	// Domain is the Zitadel instance domain configured for this installation
	Domain string `json:"domain,omitempty"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalName: m.Domain,
	}
}