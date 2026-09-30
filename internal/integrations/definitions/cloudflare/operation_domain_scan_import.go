package cloudflare

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/domainscan"
)

// DomainScanImportVendor is one reviewer-accepted vendor, keyed by a client-assigned Ref
type DomainScanImportVendor struct {
	// Ref is a client-assigned identifier for this vendor, referenced by EntityRefs elsewhere
	Ref string `json:"ref"`
	// Name is the vendor's name
	Name string `json:"name"`
	// LegalName is the vendor's raw legal entity name, if known and different from Name
	LegalName string `json:"legalName,omitempty"`
	// Domain is the vendor's domain, if known
	Domain string `json:"domain,omitempty"`
	// Categories are the vendor's detected categories
	Categories []string `json:"categories,omitempty"`
}

// DomainScanImportAsset is one reviewer-accepted asset, keyed by a client-assigned Ref
type DomainScanImportAsset struct {
	// Ref is a client-assigned identifier for this asset, referenced by AssetRefs elsewhere
	Ref string `json:"ref"`
	// Name is the asset's display name
	Name string `json:"name"`
	// Identifier is the asset's domain, IP, or other unique identifier
	Identifier string `json:"identifier,omitempty"`
	// Website is the asset's URL, if known
	Website string `json:"website,omitempty"`
	// Categories are the asset's detected categories
	Categories []string `json:"categories,omitempty"`
}

// DomainScanImportPlatform is one accepted platform, keyed by a client-assigned Ref
type DomainScanImportPlatform struct {
	// Ref is a client-assigned identifier for this platform, referenced by PlatformRefs elsewhere
	Ref string `json:"ref"`
	// Name is the platform's name
	Name string `json:"name"`
	// Description is the platform's description
	Description string `json:"description,omitempty"`
	// EntityRefs are the Refs of accepted vendors linked to this platform
	EntityRefs []string `json:"entityRefs,omitempty"`
	// AssetRefs are the Refs of accepted assets linked to this platform
	AssetRefs []string `json:"assetRefs,omitempty"`
}

// DomainScanImportSystem is one accepted system detail, linked to its own vendors/assets/platforms
type DomainScanImportSystem struct {
	// Name is the system's name
	Name string `json:"name"`
	// Description is the system's description
	Description string `json:"description,omitempty"`
	// EntityRefs are the Refs of accepted vendors linked to this system
	EntityRefs []string `json:"entityRefs,omitempty"`
	// AssetRefs are the Refs of accepted assets linked to this system
	AssetRefs []string `json:"assetRefs,omitempty"`
	// PlatformRefs are the Refs of accepted platforms this system belongs to
	PlatformRefs []string `json:"platformRefs,omitempty"`
}

// DomainScanImportFinding is one accepted finding
type DomainScanImportFinding struct {
	// Category is the finding's category
	Category string `json:"category,omitempty"`
	// Description is the finding's description
	Description string `json:"description,omitempty"`
	// Severity is the finding's severity
	Severity string `json:"severity,omitempty"`
	// Domain is the domain this finding was raised against, if known
	Domain string `json:"domain,omitempty"`
}

// DomainScanImport imports a reviewer-accepted domain scan report into real records
type DomainScanImport struct {
	// OrganizationID is the organization the created records belong to
	OrganizationID string `json:"organizationId"`
	// ScanIDs are the Scan records the created records should link back to
	ScanIDs []string `json:"scanIds"`
	// Platforms are the accepted platforms, if any
	Platforms []DomainScanImportPlatform `json:"platforms,omitempty"`
	// Systems are the accepted system details
	Systems []DomainScanImportSystem `json:"systems,omitempty"`
	// Vendors are the accepted vendors
	Vendors []DomainScanImportVendor `json:"vendors"`
	// Assets are the accepted assets
	Assets []DomainScanImportAsset `json:"assets"`
	// Findings are the accepted findings
	Findings []DomainScanImportFinding `json:"findings,omitempty"`
	// Branding is the brand design config for a trust center environment
	Branding *domainscan.BrandDesignProfile `json:"branding,omitempty"`
}

// DomainScanImportOp is the operation ref for importing an accepted domain scan review
var DomainScanImportOp = types.OperationRefOf[DomainScanImport]().HandlesRequest(runDomainScanImport) //nolint:revive

// runDomainScanImport imports the reviewer-accepted domain scan report through the domain scan saga
func runDomainScanImport(ctx context.Context, request types.OperationRequest, cfg DomainScanImport) (json.RawMessage, error) {
	saga := domainScanSaga{services: request.Services}

	return nil, saga.HandleImportDomainScanReview(ctx, cfg)
}
