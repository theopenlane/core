package githubapp

import (
	"time"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the GitHub App integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0GHAPP000000000000000001")
	// installation is the typed installation metadata handle for the GitHub App definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// gitHubAppCredential is the credential slot for GitHub App installation credentials
	gitHubAppCredential = types.CredentialRefOf[githubAppCredential]()
	// gitHubClient is the client ref for the GitHub GraphQL client used by this definition
	gitHubClient = types.ClientRefOf[GraphQLClient]()
	// InstallationEventsWebhook is the webhook ref for GitHub App installation-scoped deliveries
	InstallationEventsWebhook = types.NewWebhookRef("installation.events")
	// pingWebhookEvent is the webhook event ref for GitHub ping events
	pingWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("ping")
	// installationCreatedWebhookEvent is the webhook event ref for GitHub installation created events
	installationCreatedWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("installation.created")
	// installationDeletedWebhookEvent is the webhook event ref for GitHub installation deleted events
	installationDeletedWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("installation.deleted")
	// dependabotAlertWebhookEvent is the webhook event ref for Dependabot alert events
	dependabotAlertWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("dependabot_alert")
	// codeScanningAlertWebhookEvent is the webhook event ref for code scanning alert events
	codeScanningAlertWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("code_scanning_alert")
	// secretScanningAlertWebhookEvent is the webhook event ref for secret scanning alert events
	secretScanningAlertWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("secret_scanning_alert")
)

const (
	// githubAlertTypeDependabot is the variant name for Dependabot webhook alert payloads
	githubAlertTypeDependabot = "dependabot"
	// githubAlertTypeDependabotPoll is the variant name for Dependabot alerts collected via poll
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
	// OrganizationName is the organization this was installed in
	OrganizationName string `json:"organizationName,omitempty"`
}

// DirectorySync controls the directory sync operation
type DirectorySync struct {
	types.OperationSettings
	// DisableGroupSync will just sync users and no groups or group memberships
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users from GitHub, disable groups sync operations"`
}

// VulnerabilitySync controls the vulnerability collect operation
type VulnerabilitySync struct {
	types.OperationSettings
	// MaxRepos caps the number of repositories scanned during one run
	MaxRepos int `json:"maxRepos,omitempty" jsonschema:"title=Max Repositories,description=Optional cap on the number of repositories to scan."`
}

// RepositorySync controls the repository sync operation
type RepositorySync struct {
	types.OperationSettings
}

// InstallationMetadata holds the stable GitHub App installation identity attributes
type InstallationMetadata struct {
	// InstallationID is the GitHub App installation identifier
	InstallationID string `json:"installationId,omitempty" jsonschema:"title=installation ID"`
	// OrganizationName is the organization the GitHub App was installed into, empty for a personal org
	OrganizationName string `json:"organizationName,omitempty"  jsonschema:"title=organization"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID:   m.InstallationID,
		ExternalName: m.OrganizationName,
	}
}
