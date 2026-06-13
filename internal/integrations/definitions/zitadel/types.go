package zitadel

import (
	zitadelUser "github.com/zitadel/zitadel-go/v3/pkg/client/user/v2"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Zitadel integration definition
	definitionID = types.NewDefinitionRef("def_01K0ZITADEL000000000000001")
	// integration is the typed installation metadata handle for the Zitadel definition
	// integration = types.NewInstallationRef(resolveInstallationMetadata)
	// zitadelCredentialSchema is the JSON schema for the Zitadel credential
	// zitadelCredential is the typed runtime ref for resolving the credential
	zitadelCredentialSchema, zitadelCredential = providerkit.CredentialSchema[CredentialSchema]()
	// zitadelClient is the client ref for the Zitadel user service API client
	zitadelClient = types.NewClientRef[*zitadelUser.Client]()
	// healthCheckSchema, healthCheckOperation is the operation ref for the health check
	// healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema, directorySyncOperation is the operation ref for directory sync
	// directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// CredentialSchema holds the Zitadel instance credentials for one installation
type CredentialSchema struct {
	// Domain is the Zitadel instance domain (e.g. https://my-instance.zitadel.cloud)
	Domain string `json:"domain" jsonschema:"required,title=Domain"`
	// Token is the Zitadel Personal Access Token
	Token string `json:"token" jsonschema:"required,title=Personal Access Token"`
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