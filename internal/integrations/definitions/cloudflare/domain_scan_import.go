package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/internal/ent/generated/scan"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/domainscan"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// domainScanImportVendorEntityType is the entityTypeName to use when creating the entities from
// a completed scan
const domainScanImportVendorEntityType = "vendor"

// importSummary counts what an accepted domain scan review created, for the follow-up notification
type importSummary struct {
	PlatformIDs     []string `json:"platform_ids,omitempty"`
	SystemDetailIDs []string `json:"system_detail_ids,omitempty"`
	EntityIDs       []string `json:"entity_ids,omitempty"`
	AssetIDs        []string `json:"asset_ids,omitempty"`
	FindingIDs      []string `json:"finding_ids,omitempty"`
}

// HandleImportDomainScanReview creates the records for a reviewer-accepted domain scan report,
// all under one transaction, and sends a follow-up Notification once done
func (s domainScanSaga) HandleImportDomainScanReview(ctx context.Context, envelope operations.ImportDomainScanReviewEnvelope) error {
	ctx = logx.WithFields(ctx, logx.LogFields{
		"organization_id": envelope.OrganizationID,
		"scan_ids":        envelope.ScanIDs,
	})

	systemCtx := domainScanSystemContext(ctx, envelope.OrganizationID)

	summary, err := workflows.WithTx(systemCtx, s.services.DB(), nil, func(tx *generated.Tx) (importSummary, error) {
		return importDomainScanReview(systemCtx, tx.Client(), envelope)
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed importing accepted review")
		return err
	}

	return s.notifyDomainScanImportComplete(systemCtx, envelope.OrganizationID, summary)
}

// notifyDomainScanImportComplete sends the Notification that surfaces an import's results
func (s domainScanSaga) notifyDomainScanImportComplete(ctx context.Context, organizationID string, summary importSummary) error {
	data, err := jsonx.ToMap(summary)
	if err != nil {
		return err
	}

	body := fmt.Sprintf(
		"Imported %d vendor(s), %d asset(s), %d system(s), and %d finding(s)",
		len(summary.EntityIDs), len(summary.AssetIDs), len(summary.SystemDetailIDs), len(summary.FindingIDs),
	)

	_, err = s.services.DB().Notification.Create().
		SetOwnerID(organizationID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType("scan.import.completed").
		SetTitle("Domain scan import completed").
		SetBody(body).
		SetData(data).
		SetTopic(enums.NotificationTopicImportComplete).
		Save(ctx)

	return err
}

// importDomainScanReview creates every accepted object in order: vendors and assets first (so
// their real IDs are known), then the platform and its system details (which link back to the
// vendors/assets via ref), then findings
func importDomainScanReview(ctx context.Context, client *generated.Client, envelope operations.ImportDomainScanReviewEnvelope) (importSummary, error) {
	var summary importSummary

	assetIDByRef, err := createDomainScanAssets(ctx, client, envelope.OrganizationID, envelope.Assets, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	for _, id := range assetIDByRef {
		summary.AssetIDs = append(summary.AssetIDs, id)
	}

	vendorNameByAssetDomain, err := domainScanAssetVendorNames(ctx, client, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	entityIDByRef, err := findOrCreateDomainScanVendors(ctx, client, envelope.OrganizationID, envelope.Vendors, envelope.Assets, assetIDByRef, vendorNameByAssetDomain, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	for _, id := range entityIDByRef {
		summary.EntityIDs = append(summary.EntityIDs, id)
	}

	platformIDByRef := map[string]string{}

	if len(envelope.Platforms) > 0 {
		platformIDByRef, err = createDomainScanPlatforms(ctx, client, envelope.OrganizationID, envelope.Platforms, envelope.ScanIDs, entityIDByRef, assetIDByRef)
		if err != nil {
			return summary, err
		}

		for _, id := range platformIDByRef {
			summary.PlatformIDs = append(summary.PlatformIDs, id)
		}
	}

	if len(envelope.Systems) > 0 {
		summary.SystemDetailIDs, err = createDomainScanSystemDetails(ctx, client, envelope.OrganizationID, envelope.Systems, platformIDByRef, entityIDByRef, assetIDByRef)
		if err != nil {
			return summary, err
		}
	}

	summary.FindingIDs, err = createDomainScanFindings(ctx, client, envelope.OrganizationID, envelope.Findings, envelope.Assets, assetIDByRef, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	return summary, nil
}

// findOrCreateDomainScanVendors resolves each accepted vendor to an Entity ID, reusing an
// existing org Entity by name if one already exists, and auto-links any accepted asset the scan
// itself attributed to that vendor
func findOrCreateDomainScanVendors(ctx context.Context, client *generated.Client, ownerID string, vendors []operations.ImportDomainScanReviewVendor, assets []operations.ImportDomainScanReviewAsset, assetIDByRef, vendorNameByAssetDomain map[string]string, scanIDs []string) (map[string]string, error) {
	ids := make(map[string]string, len(vendors))
	if len(vendors) == 0 {
		return ids, nil
	}

	vendorEntityTypeID, err := client.EntityType.Query().Where(entitytype.NameEqualFold(domainScanImportVendorEntityType), entitytype.OwnerID(ownerID)).OnlyID(ctx)
	if err != nil {
		return nil, fmt.Errorf("looking up vendor entity type: %w", err)
	}

	for _, vendor := range vendors {
		matchedAssetIDs := domainScanMappingMatches(assets, assetIDByRef, func(assetDomain string) bool {
			registrable, ok := domainscan.RegistrableDomain(assetDomain)
			if !ok {
				return false
			}

			return strings.EqualFold(vendorNameByAssetDomain[registrable], vendor.Name)
		})

		entityName := vendor.LegalName
		if entityName == "" {
			entityName = vendor.Name
		}

		existing, err := client.Entity.Query().
			Where(
				entity.OwnerID(ownerID),
				entity.Or(
					entity.NameEqualFold(entityName),
					entity.NameEqualFold(vendor.Name),
					entity.DisplayNameEqualFold(vendor.Name),
				),
			).
			First(ctx)

		switch {
		case err == nil:
			ids[vendor.Ref] = existing.ID

			if len(matchedAssetIDs) > 0 {
				if err := client.Entity.UpdateOneID(existing.ID).AddAssetIDs(matchedAssetIDs...).Exec(ctx); err != nil {
					return nil, fmt.Errorf("linking assets to existing vendor %q: %w", vendor.Name, err)
				}
			}

			continue
		case !generated.IsNotFound(err):
			return nil, fmt.Errorf("looking up existing vendor %q: %w", vendor.Name, err)
		}

		input := generated.CreateEntityInput{
			Name:           &entityName,
			DisplayName:    &vendor.Name,
			OwnerID:        &ownerID,
			ApprovedForUse: lo.ToPtr(true),
			EntityTypeID:   &vendorEntityTypeID,
			ScanIDs:        scanIDs,
			Tags:           vendor.Categories,
			AssetIDs:       matchedAssetIDs,
		}

		if vendor.Domain != "" {
			input.Domains = []string{vendor.Domain}
		}

		created, err := client.Entity.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating vendor %q: %w", vendor.Name, err)
		}

		ids[vendor.Ref] = created.ID
	}

	return ids, nil
}

// domainScanAssetVendorNames maps each domain (lowercased) to the vendor name the scan itself
// attributed it to, read back from the completed scans' own dns_records
func domainScanAssetVendorNames(ctx context.Context, client *generated.Client, scanIDs []string) (map[string]string, error) {
	scans, err := client.Scan.Query().Where(scan.IDIn(scanIDs...)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("looking up scans for vendor-asset attribution: %w", err)
	}

	vendorNameByDomain := map[string]string{}

	for _, s := range scans {
		var report domainscan.ScanReport
		if err := jsonx.RoundTrip(s.Metadata, &report); err != nil || report.Assets == nil {
			continue
		}

		for _, record := range report.Assets.DNSRecords {
			if record.Domain == "" || record.Vendor == "" {
				continue
			}

			registrable, ok := domainscan.RegistrableDomain(record.Domain)
			if !ok {
				continue
			}

			vendorNameByDomain[registrable] = record.Vendor
		}
	}

	return vendorNameByDomain, nil
}

// domainScanMappingMatches returns the IDs of accepted assets whose lowercased, dot-trimmed
// identifier satisfies match
func domainScanMappingMatches(assets []operations.ImportDomainScanReviewAsset, assetIDByRef map[string]string, match func(assetDomain string) bool) []string {
	var ids []string

	for _, asset := range assets {
		if asset.Identifier == "" {
			continue
		}

		assetDomain := strings.ToLower(strings.TrimSuffix(asset.Identifier, "."))
		if !match(assetDomain) {
			continue
		}

		if id, ok := assetIDByRef[asset.Ref]; ok {
			ids = append(ids, id)
		}
	}

	return ids
}

// createDomainScanAssets creates one Asset per accepted asset
func createDomainScanAssets(ctx context.Context, client *generated.Client, ownerID string, assets []operations.ImportDomainScanReviewAsset, scanIDs []string) (map[string]string, error) {
	ids := make(map[string]string, len(assets))

	for _, asset := range assets {
		input := generated.CreateAssetInput{
			Name:       asset.Name,
			OwnerID:    &ownerID,
			Categories: asset.Categories,
			ScanIDs:    scanIDs,
			Identifier: &asset.Identifier,
			Website:    &asset.Website,
		}

		created, err := client.Asset.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating asset %q: %w", asset.Name, err)
		}

		ids[asset.Ref] = created.ID
	}

	return ids, nil
}

// createDomainScanPlatforms creates one Platform per accepted platform, each linked to the scans
// they were generated from and to whichever accepted vendors/assets the reviewer marked as
// belonging to it, returning a ref -> Platform ID map so systems can attach to them by ref
func createDomainScanPlatforms(ctx context.Context, client *generated.Client, ownerID string, platforms []operations.ImportDomainScanReviewPlatform, scanIDs []string, entityIDByRef, assetIDByRef map[string]string) (map[string]string, error) {
	ids := make(map[string]string, len(platforms))

	for _, platform := range platforms {
		input := generated.CreatePlatformInput{
			Name:             platform.Name,
			OwnerID:          &ownerID,
			GeneratedScanIDs: scanIDs,
			EntityIDs:        resolveDomainScanRefs(platform.EntityRefs, entityIDByRef),
			AssetIDs:         resolveDomainScanRefs(platform.AssetRefs, assetIDByRef),
		}

		if platform.Description != "" {
			input.Description = &platform.Description
		}

		created, err := client.Platform.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating platform %q: %w", platform.Name, err)
		}

		ids[platform.Ref] = created.ID
	}

	return ids, nil
}

// createDomainScanSystemDetails creates one SystemDetail per accepted system, parented to the
// platforms it references by ref (required by SystemDetail's policy) and linked to that system's
// own subset of accepted vendors/assets
func createDomainScanSystemDetails(ctx context.Context, client *generated.Client, ownerID string, systems []operations.ImportDomainScanReviewSystem, platformIDByRef, entityIDByRef, assetIDByRef map[string]string) ([]string, error) {
	ids := make([]string, 0, len(systems))

	for _, system := range systems {
		platformIDs := resolveDomainScanRefs(system.PlatformRefs, platformIDByRef)

		input := generated.CreateSystemDetailInput{
			SystemName:  system.Name,
			OwnerID:     &ownerID,
			PlatformIDs: platformIDs,
			EntityIDs:   resolveDomainScanRefs(system.EntityRefs, entityIDByRef),
			AssetIDs:    resolveDomainScanRefs(system.AssetRefs, assetIDByRef),
			Description: &system.Description,
		}

		created, err := client.SystemDetail.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating system detail %q: %w", system.Name, err)
		}

		ids = append(ids, created.ID)
	}

	return ids, nil
}

// createDomainScanFindings creates one Finding per accepted finding, linked to the scans it came
// from and auto-linked to any accepted asset matching its domain
func createDomainScanFindings(ctx context.Context, client *generated.Client, ownerID string, findings []operations.ImportDomainScanReviewFinding, assets []operations.ImportDomainScanReviewAsset, assetIDByRef map[string]string, scanIDs []string) ([]string, error) {
	ids := make([]string, 0, len(findings))

	for _, finding := range findings {
		domain := strings.ToLower(strings.TrimSuffix(finding.Domain, "."))

		input := generated.CreateFindingInput{
			OwnerID: &ownerID,
			ScanIDs: scanIDs,
			AssetIDs: domainScanMappingMatches(assets, assetIDByRef, func(assetDomain string) bool {
				return domain != "" && (assetDomain == domain || strings.HasSuffix(assetDomain, "."+domain))
			}),
			Category:    &finding.Category,
			Description: &finding.Description,
			Severity:    &finding.Severity,
		}

		created, err := client.Finding.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating finding: %w", err)
		}

		ids = append(ids, created.ID)
	}

	return ids, nil
}

// resolveDomainScanRefs looks up each ref in idByRef, skipping any that don't resolve. The
// resolver that emitted this envelope already validated every ref before emitting, so an
// unresolved ref here would only happen if that validation was bypassed - fail open (drop the
// link) rather than erroring the whole import over a single bad link
func resolveDomainScanRefs(refs []string, idByRef map[string]string) []string {
	ids := make([]string, 0, len(refs))

	for _, ref := range refs {
		if id, ok := idByRef[ref]; ok {
			ids = append(ids, id)
		}
	}

	return ids
}
