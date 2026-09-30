package cloudflare

import (
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/domainscan"
)

var (
	// DefinitionID is the stable identifier for the Cloudflare integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0CFLARE00000000000000001")
	// installation is the typed installation metadata handle for the Cloudflare definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// cloudflareCredential is the typed credential slot for the Cloudflare API token
	cloudflareCredential = types.CredentialRefOf[CredentialSchema]()
	// cloudflareClient is the client ref for the Cloudflare API client used by this definition
	cloudflareClient = types.ClientRefOf[*CloudflareClient]().Using(cloudflareCredential)
	// cloudflareConnection is the connection mode selected by the Cloudflare API token credential
	cloudflareConnection = types.NewConnectionRef(cloudflareCredential).Enables(cloudflareClient)
	// userInput is the installation user input layout for the Cloudflare definition
	userInput = types.NewUserInputRef[UserInput]("cloudflare")
)

const (
	// DomainScanPerformedBy marks a Scan record as one the system should actually submit to the cloudflare domain scan job
	DomainScanPerformedBy = "openlane_domain_scan"

	// DomainScanBrandDesignOnlyMetadataKey selects the brand-design-only scan path
	DomainScanBrandDesignOnlyMetadataKey = "brandDesignOnly"

	// DomainScanApplyBrandDesignToPreviewMetadataKey makes sure we apply the extracted brand design to only the preview trustcenter environment
	DomainScanApplyBrandDesignToPreviewMetadataKey = "applyBrandDesignToPreview"

	// DomainScanApplyBrandDesignToLiveMetadataKey makes sure we apply the extracted brand design to only the live trustcenter environment
	DomainScanApplyBrandDesignToLiveMetadataKey = "applyBrandDesignToLive"

	// DomainScanGroupMetadataKey is the Scan.Metadata key carrying the shared group id for scans created together (e.g. every domain from one organization settings update), so scans submitted independently can still be recombined into a single notification once the whole group reaches a terminal state
	DomainScanGroupMetadataKey = "scan_group_id"
)

// RuntimeConfig is the runtime-provisioned configuration for the operator-owned Cloudflare account
type RuntimeConfig struct {
	// APIToken is the Cloudflare API token for the operator-owned account
	APIToken string `json:"apitoken" koanf:"apitoken" jsonschema:"description=Cloudflare API token for the operator-owned account" sensitive:"true"`
	// AccountID is the Cloudflare account identifier for the operator-owned account
	AccountID string `json:"accountid" koanf:"accountid" jsonschema:"description=Cloudflare account ID for the operator-owned account"`
	// DomainScan configures vendor/technology classification and enrichment behavior for onboarding domain scan reports
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
	// Disable switches the directory sync operation off for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of account members from Cloudflare"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.),example=Example: payload.status = 'ACTIVE'"`
}

// FindingsSync holds installation-specific configuration for Cloudflare Security Center insights
type FindingsSync struct {
	// Disable switches the findings sync operation off for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of security insights from Cloudflare"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting,example=Example: payload.severity == 'Critical'"`
}

// AssetSync holds installation-specific configuration for Cloudflare domain assets
type AssetSync struct {
	// Disable switches the asset sync operation off for the installation
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of domains from Cloudflare"`
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
