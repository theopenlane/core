package cloudflare

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/domainscan"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/logx"
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

// domainScanImportPolicy is the execution policy the domain scan import runs and ingests under
var domainScanImportPolicy = types.ExecutionPolicy{SkipRunRecord: true}

// DomainScanImportOp is the operation ref for importing an accepted domain scan review
// it declares no ingest contracts so it runs as the reviewer who dispatched it, not as the integration
var DomainScanImportOp = types.OperationPayloadOf[DomainScanImport]().
	Policy(domainScanImportPolicy).
	CustomerSelectable(false).
	Internal()

// runDomainScanImport ingests the reviewer-accepted report, applies its branding, and notifies the organization
func runDomainScanImport(ctx context.Context, request types.OperationRequest, cfg DomainScanImport) (json.RawMessage, error) {
	ctx = logx.WithFields(ctx, logx.LogFields{"scan_ids": cfg.ScanIDs})

	result, importErr := importDomainScan(ctx, request, cfg)

	var (
		brandingUpdated bool
		brandingErr     error
	)

	if cfg.Branding != nil && importErr == nil {
		brandingUpdated, brandingErr = applyBrandingToTrustCenter(ctx, request.DB, *cfg.Branding)
	}

	detail := domainScanImportDetail(brandingUpdated, importErr != nil || result.Failed > 0)

	notifyErr := operations.NotifyImportOutcome(domainScanSystemContext(ctx, cfg.OrganizationID), "Domain scan", result, importErr, detail, "")

	return nil, errors.Join(importErr, brandingErr, notifyErr)
}

// importDomainScan ingests the report's records as the calling reviewer
func importDomainScan(ctx context.Context, request types.OperationRequest, cfg DomainScanImport) (operations.IngestResult, error) {
	payloadSets, err := domainScanImportPayloadSets(ctx, request.DB, cfg)
	if err != nil {
		return operations.IngestResult{}, err
	}

	oc, _ := gala.OperationContextFromContext(ctx)

	ingest := operations.IngestContext{
		Registry:     request.Services.Definitions(),
		DB:           request.DB,
		Runtime:      request.Services.Gala(),
		Integration:  request.Integration,
		DefinitionID: DefinitionID.ID(),
	}

	return operations.ProcessPayloadSets(ctx, ingest, DomainScanImportOp.Name(), domainScanImportContracts, domainScanImportPolicy, payloadSets, operations.IngestOptionsFromOperationContext(oc))
}
