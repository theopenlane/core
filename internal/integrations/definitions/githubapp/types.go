package githubapp

import (
	"time"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the GitHub App integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0GHAPP000000000000000001")
	// GitHubClient is the client ref for the GitHub GraphQL client used by this definition
	gitHubClient = types.NewClientRef[GraphQLClient]()
	// healthDefaultOperation is the operation ref for the GitHub App health check
	healthCheckSchema, healthDefaultOperation = providerkit.OperationSchema[HealthCheck]()
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
	// installation is the typed installation metadata handle for the GitHub App definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)

	gitHubAppCredentialSchema, gitHubAppCredential            = providerkit.CredentialSchema[githubAppCredential]()
	repositorySyncSchema, repositorySyncOperation             = providerkit.OperationSchema[RepositorySync]()
	vulnerabilityCollectSchema, vulnerabilityCollectOperation = providerkit.OperationSchema[VulnerabilityCollectConfig]()
	directorySyncSchema, directorySyncOperation               = providerkit.OperationSchema[DirectorySync]()
)

const (
	// githubAlertTypeDependabot is the variant name for Dependabot security alert payloads
	githubAlertTypeDependabot = "dependabot"
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
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}

// InstallationMetadata holds the stable GitHub App installation identity attributes
type InstallationMetadata struct {
	// InstallationID is the GitHub App installation identifier
	InstallationID string `json:"installationId,omitempty" jsonschema:"title=installation ID"`
}
