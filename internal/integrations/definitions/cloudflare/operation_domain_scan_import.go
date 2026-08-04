package cloudflare

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/types"
)

// DomainScanImportVendor is one vendor the reviewer accepted, keyed by a client-assigned
// Ref so DomainScanImportPlatform/DomainScanImportSystem can reference it before it
// has a real Entity ID
type DomainScanImportVendor struct {
	// Ref is a client-assigned identifier for this vendor, referenced by EntityRefs elsewhere in the envelope
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

// DomainScanImportAsset is one asset the reviewer accepted, keyed by a client-assigned Ref
// so DomainScanImportPlatform/DomainScanImportSystem can reference it before it has a
// real Asset ID
type DomainScanImportAsset struct {
	// Ref is a client-assigned identifier for this asset, referenced by AssetRefs elsewhere in the envelope
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

// DomainScanImportPlatform is one accepted platform, linked to a subset of the accepted
// vendors/assets, and keyed by a client-assigned Ref so DomainScanImportSystem can
// reference it before it has a real Platform ID
type DomainScanImportPlatform struct {
	// Ref is a client-assigned identifier for this platform, referenced by PlatformRefs elsewhere in the envelope
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

// DomainScanImportSystem is one accepted system detail, linked to its own subset of the
// accepted vendors/assets/platforms
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

// DomainScanImport imports a reviewer-accepted domain scan report into real
// Platform/SystemDetail/Entity/Asset/Finding records
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
}

// Handle adapts DomainScanImport to the generic operation registration boundary
func (d DomainScanImport) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		var cfg DomainScanImport
		if err := json.Unmarshal(request.Config, &cfg); err != nil {
			return nil, ErrOperationConfigInvalid
		}

		saga := domainScanSaga{services: request.Services}

		return nil, saga.HandleImportDomainScanReview(ctx, cfg)
	}
}
