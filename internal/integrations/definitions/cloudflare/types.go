package cloudflare

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/domainscan"
)

var (
	// DefinitionID is the stable identifier for the Cloudflare integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0CFLARE00000000000000001")
	// installation is the typed installation metadata handle for the Cloudflare definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// cloudflareSchema is the credential schema for the Cloudflare integration definition
	cloudflareSchema, cloudflareCredential = providerkit.CredentialSchema[CredentialSchema]()
	// cloudflareClient is the client ref for the Cloudflare API client used by this definition
	cloudflareClient = types.NewClientRef[*CloudflareClient]()
	// healthDefaultOperation is the operation ref for the Cloudflare health check
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// directorySyncSchema is the operation ref for the directory account sync operation
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
	// assetSyncSchema is the operation ref for the domain asset sync operation
	assetSyncSchema, assetSyncOperation = providerkit.OperationSchema[AssetSync]()
	// findingsSyncSchema is the operation ref for the Security Center insights finding sync operation
	findingsSyncSchema, findingsSyncOperation = providerkit.OperationSchema[FindingsSync]()
	// domainScanSubmitSchema is the operation ref for submitting domains to the URL Scanner
	domainScanSubmitSchema, DomainScanSubmitOp = providerkit.OperationSchema[DomainScanSubmit]() //nolint:revive
	// domainScanPollSchema is the operation ref for polling a submitted URL Scanner result
	domainScanPollSchema, DomainScanPollOp = providerkit.OperationSchema[DomainScanPoll]() //nolint:revive
	// domainScanGatherEnrichmentSchema is the operation ref for gathering enrichment data for a domain
	domainScanGatherEnrichmentSchema, DomainScanEnrichmentOp = providerkit.OperationSchema[DomainScanGatherEnrichment]() //nolint:revive
	// domainScanBuildReportSchema is the operation ref for building the scan report from a completed URL Scanner result and gathered enrichment
	domainScanBuildReportSchema, DomainScanBuildReportOp = providerkit.OperationSchema[DomainScanBuildReport]() //nolint:revive
	// domainScanRequestSchema is the operation ref for requesting a domain scan; used both by
	// customer-facing calls (queues a pending Scan) and the system re-dispatch from the listener
	// on Scan creation (actually runs the saga for it)
	domainScanRequestSchema, DomainScanRequestOp = providerkit.OperationSchema[DomainScanRequest]() //nolint:revive
	// domainScanImportSchema is the operation ref for importing an accepted domain scan review
	domainScanImportSchema, DomainScanImportOp = providerkit.OperationSchema[DomainScanImport]() //nolint:revive
	// runtimeCloudflareSchema is the JSON schema and typed ref for the runtime Cloudflare config
	runtimeCloudflareSchema, runtimeCloudflareRef = providerkit.RuntimeSchema[RuntimeConfig]()
)

const (
	// DomainScanPerformedBy marks a Scan record as one the system should actually submit to the cloudflare domain scan job
	DomainScanPerformedBy = "openlane_domain_scan"
	// DomainScanGroupMetadataKey is the Scan.Metadata key carrying the shared group id for scans
	// created together (e.g. every domain from one organization settings update), so scans
	// submitted independently can still be recombined into a single notification once the whole
	// group reaches a terminal state
	DomainScanGroupMetadataKey = "scan_group_id"
)

// RuntimeConfig is the runtime-provisioned configuration for the operator-owned
// Cloudflare account. Sourced from koanf/environment at startup; used for system-initiated
// Cloudflare calls (e.g. onboarding domain scans) that are not tied to a customer installation
type RuntimeConfig struct {
	// APIToken is the Cloudflare API token for the operator-owned account
	APIToken string `json:"apitoken" koanf:"apitoken" jsonschema:"description=Cloudflare API token for the operator-owned account" sensitive:"true"`
	// AccountID is the Cloudflare account identifier for the operator-owned account
	AccountID string `json:"accountid" koanf:"accountid" jsonschema:"description=Cloudflare account ID for the operator-owned account"`
	// DomainScan configures vendor/technology classification and enrichment behavior for
	// onboarding domain scan reports
	DomainScan domainscan.ReportConfig `json:"domainscan" koanf:"domainscan" jsonschema:"description=Vendor/technology classification and enrichment behavior for onboarding domain scan reports"`
}

// Provisioned reports whether the runtime config has the minimum required fields to make Cloudflare API calls
func (c RuntimeConfig) Provisioned() bool {
	return c.APIToken != "" && c.AccountID != ""
}

const (
	assetSyncRegistrarPageSize = 50
	assetSyncMinIntervalHours  = 24
	assetSyncMaxIntervalDays   = 7
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// DirectorySync includes the configuration for identity accounts from Cloudflare members
	DirectorySync DirectorySync `json:"directorySync,omitempty" jsonschema:"title=Directory Account Sync"`
	// AssetSync includes the configuration for Cloudflare domains as assets
	AssetSync AssetSync `json:"assetSync,omitempty" jsonschema:"title=Cloudflare Asset Sync"`
	// FindingsSync includes the configuration for findings from Cloudflare Security Center insights
	FindingsSync FindingsSync `json:"findingSync,omitempty" jsonschema:"title=Security Insights Sync"`
}

// DirectorySync holds installation-specific configuration collected from the user
type DirectorySync struct {
	// Disable is used to disable the directory sync operation from Cloudflare
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of users and groups from Cloudflare"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.),example=Example: payload.status = 'ACTIVE'"`
}

// FindingsSync holds installation-specific configuration for Cloudflare Security Center insights
type FindingsSync struct {
	// Disable is used to disable the findings sync operation from Cloudflare
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of findings from Cloudflare Security Center insights"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting,example=Example: payload.severity == 'Critical'"`
}

// AssetSync holds installation-specific configuration for Cloudflare domain assets
type AssetSync struct {
	// Disable is used to disable the asset sync operation from Cloudflare
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of domains from Cloudflare Registrar"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting,example=Example: payload.status == 'active'"`
}

// CredentialSchema holds the Cloudflare API credentials for one installation
type CredentialSchema struct {
	// APIToken is the Cloudflare API token with permissions to read account and zone metadata
	APIToken string `json:"apiToken"          jsonschema:"required,title=API Token"`
	// AccountID is the Cloudflare account identifier used for account-scoped API calls
	AccountID string `json:"accountId,omitempty" jsonschema:"required,title=Account ID,description=Cloudflare account ID required for listing account members."`
}

// InstallationMetadata holds the stable Cloudflare account identity for one installation
type InstallationMetadata struct {
	// AccountID is the Cloudflare account identifier used for account-scoped collection
	AccountID string `json:"accountId,omitempty" jsonschema:"title=Account ID"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.AccountID,
	}
}
