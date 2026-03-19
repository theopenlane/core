package githubapp

import "github.com/theopenlane/core/internal/integrations/types"

var (
	// DefinitionID is the stable identifier for the GitHub App integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0GHAPP000000000000000001")
	// GitHubClient is the client ref for the GitHub GraphQL client used by this definition
	GitHubClient = types.NewClientRef[GraphQLClient]()
	// HealthDefaultOperation is the operation ref for the GitHub App health check
	HealthDefaultOperation = types.NewOperationRef[struct{}]("health.default")
	// RepositorySyncOperation is the operation ref for the repository sync operation
	RepositorySyncOperation = types.NewOperationRef[struct{}]("repository.sync")
	// VulnerabilityCollectOperation is the operation ref for the vulnerability collection operation
	VulnerabilityCollectOperation = types.NewOperationRef[VulnerabilityCollectConfig]("vulnerability.collect")
	// InstallationEventsWebhook is the webhook ref for GitHub App installation-scoped deliveries
	InstallationEventsWebhook = types.NewWebhookRef("installation.events")
	// PingWebhookEvent is the webhook event ref for GitHub ping events
	PingWebhookEvent = types.NewWebhookEventRef[struct{}]("ping")
	// InstallationCreatedWebhookEvent is the webhook event ref for GitHub installation created events
	InstallationCreatedWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("installation.created")
	// DependabotAlertWebhookEvent is the webhook event ref for Dependabot alert events
	DependabotAlertWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("dependabot_alert")
	// CodeScanningAlertWebhookEvent is the webhook event ref for code scanning alert events
	CodeScanningAlertWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("code_scanning_alert")
	// SecretScanningAlertWebhookEvent is the webhook event ref for secret scanning alert events
	SecretScanningAlertWebhookEvent = types.NewWebhookEventRef[githubWebhookEnvelope]("secret_scanning_alert")
	// Installation is the typed installation metadata handle for the GitHub App definition
	Installation = types.NewInstallationRef[InstallationMetadata](resolveInstallationMetadata)
)

// Slug is the unique identifier for the GitHub App integration
const Slug = "github_app"

const (
	// githubAlertTypeDependabot is the variant name for Dependabot security alert payloads
	githubAlertTypeDependabot = "dependabot"
	// githubAlertTypeCodeScanning is the variant name for code scanning alert payloads
	githubAlertTypeCodeScanning = "code_scanning"
	// githubAlertTypeSecretScan is the variant name for secret scanning alert payloads
	githubAlertTypeSecretScan = "secret_scanning"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// RepositoryFilter limits repository collection to matching repositories
	RepositoryFilter string `json:"repositoryFilter,omitempty" jsonschema:"title=Repository Filter Expression"`
	// SecurityOnly limits collection to security-focused data
	SecurityOnly bool `json:"securityOnly,omitempty" jsonschema:"title=Collect Security Signals Only"`
}

// InstallationMetadata holds the stable GitHub App installation identity attributes
type InstallationMetadata struct {
	// InstallationID is the GitHub App installation identifier
	InstallationID string `json:"installationId,omitempty" jsonschema:"title=Installation ID"`
}
