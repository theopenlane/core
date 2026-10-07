package cloudflare

import (
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/domainscan"
)

var (
	// DefinitionID is the stable identifier for the Cloudflare integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0CFLARE00000000000000001")
	// installation is the typed installation metadata handle for the Cloudflare definition
	installation = types.InstallationOf[InstallationMetadata]()
	// apiToken is the connection for the Cloudflare API token
	apiToken = types.ConnectionOf[CredentialSchema]()
)

const (
	// DomainScanPerformedBy marks a Scan record for submission to the cloudflare domain scan job
	DomainScanPerformedBy = "openlane_domain_scan"

	// DomainScanBrandDesignOnlyMetadataKey selects the brand-design-only scan path
	DomainScanBrandDesignOnlyMetadataKey = "brandDesignOnly"

	// DomainScanApplyBrandDesignToPreviewMetadataKey applies extracted brand design to preview only
	DomainScanApplyBrandDesignToPreviewMetadataKey = "applyBrandDesignToPreview"

	// DomainScanApplyBrandDesignToLiveMetadataKey applies extracted brand design to live only
	DomainScanApplyBrandDesignToLiveMetadataKey = "applyBrandDesignToLive"

	// DomainScanGroupMetadataKey is the Scan.Metadata key carrying the shared group id for scans
	DomainScanGroupMetadataKey = "scan_group_id"

	// mainFindingSyncKey is the key main's client config stored the findings sync section under
	// TODO: remove with providerkit.UpgradeFromSection once every installation has been upgraded off main's client config
	mainFindingSyncKey = "findingSync"
)

// RuntimeConfig is the runtime-provisioned configuration for the operator-owned Cloudflare account
type RuntimeConfig struct {
	// APIToken is the Cloudflare API token for the operator-owned account
	APIToken string `json:"apitoken" koanf:"apitoken" jsonschema:"description=Cloudflare API token for the operator-owned account" sensitive:"true"`
	// AccountID is the Cloudflare account identifier for the operator-owned account
	AccountID string `json:"accountid" koanf:"accountid" jsonschema:"description=Cloudflare account ID for the operator-owned account"`
	// DomainScan configures vendor/technology classification for onboarding domain scan reports
	DomainScan domainscan.ReportConfig `json:"domainscan" koanf:"domainscan" jsonschema:"description=Vendor/technology classification and enrichment behavior for onboarding domain scan reports"`
}

// Provisioned reports whether the runtime config has the fields required for Cloudflare API calls
func (c RuntimeConfig) Provisioned() bool {
	return c.APIToken != "" && c.AccountID != ""
}

const (
	assetSyncRegistrarPageSize = 50
	assetSyncMinIntervalHours  = 24
	assetSyncMaxIntervalDays   = 7
)

// DirectorySync holds installation-specific configuration for Cloudflare account members
type DirectorySync struct {
	types.OperationSettings
}

// FindingsSync holds installation-specific configuration for Cloudflare Security Center insights
type FindingsSync struct {
	types.OperationSettings
}

// AssetSync holds installation-specific configuration for Cloudflare domain assets
type AssetSync struct {
	types.OperationSettings
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
