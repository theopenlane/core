package githubapp

import (
	"time"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the GitHub App integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0GHAPP000000000000000001")
	// installation is the typed installation metadata handle for the GitHub App definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// gitHubAppCredential is the credential schema for GitHub App credentials
	_, gitHubAppCredential = providerkit.CredentialSchema[githubAppCredential]()
	// GitHubClient is the client ref for the GitHub GraphQL client used by this definition
	gitHubClient = types.NewClientRef[GraphQLClient]()
	// InstallationEventsWebhook is the webhook ref for GitHub App installation-scoped deliveries
	InstallationEventsWebhook = types.NewWebhookRef("installation.events")
	// PingWebhookEvent is the webhook event ref for GitHub ping events
	pingWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("ping")
	// InstallationCreatedWebhookEvent is the webhook event ref for GitHub installation created events
	installationCreatedWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("installation.created")
	// InstallationDeletedWebhookEvent is the webhook event ref for GitHub installation deleted events
	installationDeletedWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("installation.deleted")
	// DependabotAlertWebhookEvent is the webhook event ref for Dependabot alert events
	dependabotAlertWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("dependabot_alert")
	// CodeScanningAlertWebhookEvent is the webhook event ref for code scanning alert events
	codeScanningAlertWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("code_scanning_alert")
	// SecretScanningAlertWebhookEvent is the webhook event ref for secret scanning alert events
	secretScanningAlertWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("secret_scanning_alert")
	// healthDefaultOperation is the operation ref for the GitHub App health check
	healthCheckSchema, healthDefaultOperation = providerkit.OperationSchema[HealthCheck]()
	// repositorySyncSchema is the operation schema for the GitHub repository sync operation
	repositorySyncSchema, repositorySyncOperation = providerkit.OperationSchema[RepositorySync]()
	// vulnerabilityCollectSchema is the operation schema for the GitHub vulnerability collection operation
	vulnerabilityCollectSchema, vulnerabilityCollectOperation = providerkit.OperationSchema[VulnerabilitySync]()
	// directorySyncSchema is the operation schema for the GitHub directory sync operationß
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

const (
	// githubAlertTypeDependabot is the variant name for Dependabot webhook alert payloads
	githubAlertTypeDependabot = "dependabot"
	// githubAlertTypeDependabotPoll is the variant name for Dependabot alerts collected via GraphQL poll
	githubAlertTypeDependabotPoll = "dependabot_poll"
	// githubAlertTypeCodeScanning is the variant name for code scanning alert payloads
	githubAlertTypeCodeScanning = "code_scanning"
	// githubAlertTypeSecretScan is the variant name for secret scanning alert payloads
	githubAlertTypeSecretScan = "secret_scanning"
)

// githubAppCredential is the credential payload stored in CredentialSet.Data
type githubAppCredential struct {
	// AppID is the GitHub App identifier used to mint installation tokens
	AppID int64 `json:"appId"`
	// InstallationID is the installation selected for this credential
	InstallationID int64 `json:"installationId"`
	// AccessToken is the current installation access token
	AccessToken string `json:"accessToken"`
	// Expiry is the token expiry timestamp when available
	Expiry *time.Time `json:"expiry,omitempty"`
	// OrganizationName is the organization this was installed in, needed for disconnect when installed in an organization
	OrganizationName string `json:"organizationName,omitempty"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// VulnerabilitySync includes the configuration for findings from GitHub Security
	VulnerabilitySync VulnerabilitySyncConfig `json:"findingSync,omitempty" jsonschema:"title=GitHub Security Hub Sync"`
	// DirectorySync includes the configuration for identity accounts from GitHub organization members
	DirectorySync DirectorySync `json:"directorySync,omitempty" jsonschema:"title=Directory Account Sync"`
	// RepositorySync included the configuration of repos as assets from GitHub
	RepositorySync RepositorySync `json:"repositorySync,omitempty" jsonschema:"title=Repository Account Sync"`
}

type DirectorySync struct {
	// Disable is used to disable the directory sync operation from GitHub
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of users and groups from GitHub"`
	// DisableGroupSync will just sync users and no groups or group memberships
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users from GitHub, disable groups sync operations"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting.,example=Example: payload.Org == 'my-org'"`
}

type VulnerabilitySyncConfig struct {
	// Disable is used to disable the directory sync operation from GitHub
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of vulnerabilities from Github Security"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting.,example=Example: payload.state == 'open'"`
}

type RepositorySync struct {
	// Disable is used to disable the directory sync operation from GitHub
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of vulnerabilities from Github Security"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting.,example=Example: payload.state == 'open'"`
}

// InstallationMetadata holds the stable GitHub App installation identity attributes
type InstallationMetadata struct {
	// InstallationID is the GitHub App installation identifier
	InstallationID string `json:"installationId,omitempty" jsonschema:"title=installation ID"`
	// OrganizationName is the Organization the Github App was installed into, if empty it is installed in a personal org
	OrganizationName string `json:"organizationName,omitempty"  jsonschema:"title=organization"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID:   m.InstallationID,
		ExternalName: m.OrganizationName,
	}
}
