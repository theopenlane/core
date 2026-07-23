package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/asset"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/internal/ent/generated/finding"
	"github.com/theopenlane/core/internal/ent/generated/platform"
	"github.com/theopenlane/core/internal/ent/generated/scan"
	"github.com/theopenlane/core/internal/ent/generated/systemdetail"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/domainscan"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// domainScanImportVendorEntityType is the entityTypeName to use when creating the entities from
// a completed scan
const domainScanImportVendorEntityType = "vendor"

// importSummary records what an accepted domain scan review resolved to, for the follow-up
// notification. The per-type fields hold every resolved record; CreatedIDs holds the subset that
// this import actually created, so a resubmission can report what was new versus already there
type importSummary struct {
	PlatformIDs     []string `json:"platform_ids,omitempty"`
	SystemDetailIDs []string `json:"system_detail_ids,omitempty"`
	EntityIDs       []string `json:"entity_ids,omitempty"`
	AssetIDs        []string `json:"asset_ids,omitempty"`
	FindingIDs      []string `json:"finding_ids,omitempty"`
	CreatedIDs      []string `json:"created_ids,omitempty"`
}

// createdCount returns how many of ids this import created rather than reused
func (s importSummary) createdCount(ids []string) int {
	return len(lo.Intersect(ids, s.CreatedIDs))
}

// resolvedCount returns how many records the import resolved in total, created or reused
func (s importSummary) resolvedCount() int {
	return len(s.EntityIDs) + len(s.AssetIDs) + len(s.PlatformIDs) + len(s.SystemDetailIDs) + len(s.FindingIDs)
}

// HandleImportDomainScanReview creates the records for a reviewer-accepted domain scan report,
// all under one transaction, and sends a follow-up Notification once done
func (s domainScanSaga) HandleImportDomainScanReview(ctx context.Context, envelope DomainScanImport) error {
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
		summary.createdCount(summary.EntityIDs),
		summary.createdCount(summary.AssetIDs),
		summary.createdCount(summary.SystemDetailIDs),
		summary.createdCount(summary.FindingIDs),
	)

	// a resubmitted review resolves to records that already exist - call that out rather than
	// reporting zeros with no explanation
	if reused := summary.resolvedCount() - len(summary.CreatedIDs); reused > 0 {
		body += fmt.Sprintf("; %d existing record(s) were linked to this scan", reused)
	}

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
func importDomainScanReview(ctx context.Context, client *generated.Client, envelope DomainScanImport) (importSummary, error) {
	var summary importSummary

	assetIDByRef, createdAssetIDs, err := findOrCreateDomainScanAssets(ctx, client, envelope.OrganizationID, envelope.Assets, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	for _, id := range assetIDByRef {
		summary.AssetIDs = append(summary.AssetIDs, id)
	}

	summary.CreatedIDs = append(summary.CreatedIDs, createdAssetIDs...)

	vendorNameByAssetDomain, err := domainScanAssetVendorNames(ctx, client, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	entityIDByRef, createdEntityIDs, err := findOrCreateDomainScanVendors(ctx, client, envelope.OrganizationID, envelope.Vendors, envelope.Assets, assetIDByRef, vendorNameByAssetDomain, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	for _, id := range entityIDByRef {
		summary.EntityIDs = append(summary.EntityIDs, id)
	}

	summary.CreatedIDs = append(summary.CreatedIDs, createdEntityIDs...)

	platformIDByRef := map[string]string{}

	if len(envelope.Platforms) > 0 {
		var createdPlatformIDs []string

		platformIDByRef, createdPlatformIDs, err = findOrCreateDomainScanPlatforms(ctx, client, envelope.OrganizationID, envelope.Platforms, envelope.ScanIDs, entityIDByRef, assetIDByRef)
		if err != nil {
			return summary, err
		}

		for _, id := range platformIDByRef {
			summary.PlatformIDs = append(summary.PlatformIDs, id)
		}

		summary.CreatedIDs = append(summary.CreatedIDs, createdPlatformIDs...)
	}

	if len(envelope.Systems) > 0 {
		var createdSystemDetailIDs []string

		summary.SystemDetailIDs, createdSystemDetailIDs, err = findOrCreateDomainScanSystemDetails(ctx, client, envelope.OrganizationID, envelope.Systems, platformIDByRef, entityIDByRef, assetIDByRef)
		if err != nil {
			return summary, err
		}

		summary.CreatedIDs = append(summary.CreatedIDs, createdSystemDetailIDs...)
	}

	var createdFindingIDs []string

	summary.FindingIDs, createdFindingIDs, err = findOrCreateDomainScanFindings(ctx, client, envelope.OrganizationID, envelope.Findings, envelope.Assets, assetIDByRef, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	summary.CreatedIDs = append(summary.CreatedIDs, createdFindingIDs...)

	return summary, nil
}

// findOrCreateDomainScanVendors resolves each accepted vendor to an Entity ID, reusing an
// existing org Entity by name if one already exists, and auto-links any accepted asset the scan
// itself attributed to that vendor
func findOrCreateDomainScanVendors(ctx context.Context, client *generated.Client, ownerID string, vendors []DomainScanImportVendor, assets []DomainScanImportAsset, assetIDByRef, vendorNameByAssetDomain map[string]string, scanIDs []string) (map[string]string, []string, error) {
	ids := make(map[string]string, len(vendors))

	var createdIDs []string

	if len(vendors) == 0 {
		return ids, nil, nil
	}

	vendorEntityTypeID, err := client.EntityType.Query().Where(entitytype.NameEqualFold(domainScanImportVendorEntityType)).OnlyID(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("looking up vendor entity type: %w", err)
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

			linkedAssetIDs, err := existing.QueryAssets().IDs(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("looking up assets linked to vendor %q: %w", vendor.Name, err)
			}

			if missing := unlinkedDomainScanIDs(matchedAssetIDs, linkedAssetIDs); len(missing) > 0 {
				if err := client.Entity.UpdateOneID(existing.ID).AddAssetIDs(missing...).Exec(ctx); err != nil {
					return nil, nil, fmt.Errorf("linking assets to existing vendor %q: %w", vendor.Name, err)
				}
			}

			continue
		case !generated.IsNotFound(err):
			return nil, nil, fmt.Errorf("looking up existing vendor %q: %w", vendor.Name, err)
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
			return nil, nil, fmt.Errorf("creating vendor %q: %w", vendor.Name, err)
		}

		ids[vendor.Ref] = created.ID
		createdIDs = append(createdIDs, created.ID)
	}

	return ids, createdIDs, nil
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
func domainScanMappingMatches(assets []DomainScanImportAsset, assetIDByRef map[string]string, match func(assetDomain string) bool) []string {
	var ids []string

	for _, a := range assets {
		if a.Identifier == "" {
			continue
		}

		assetDomain := strings.ToLower(strings.TrimSuffix(a.Identifier, "."))
		if !match(assetDomain) {
			continue
		}

		if id, ok := assetIDByRef[a.Ref]; ok {
			ids = append(ids, id)
		}
	}

	return ids
}

// findOrCreateDomainScanAssets resolves each accepted asset to an Asset ID, reusing the existing
// org Asset of the same name when the review is imported more than once, and returns the subset
// of IDs it created
func findOrCreateDomainScanAssets(ctx context.Context, client *generated.Client, ownerID string, assets []DomainScanImportAsset, scanIDs []string) (map[string]string, []string, error) {
	ids := make(map[string]string, len(assets))

	var createdIDs []string

	for _, a := range assets {
		existing, err := client.Asset.Query().
			Where(asset.NameEqualFold(a.Name)).
			First(ctx)

		switch {
		case err == nil:
			ids[a.Ref] = existing.ID

			linkedScanIDs, err := existing.QueryScans().IDs(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("looking up scans linked to asset %q: %w", a.Name, err)
			}

			if missing := unlinkedDomainScanIDs(scanIDs, linkedScanIDs); len(missing) > 0 {
				if err := client.Asset.UpdateOneID(existing.ID).AddScanIDs(missing...).Exec(ctx); err != nil {
					return nil, nil, fmt.Errorf("linking scans to existing asset %q: %w", a.Name, err)
				}
			}

			continue
		case !generated.IsNotFound(err):
			return nil, nil, fmt.Errorf("looking up existing asset %q: %w", a.Name, err)
		}

		input := generated.CreateAssetInput{
			Name:       a.Name,
			OwnerID:    &ownerID,
			Categories: a.Categories,
			ScanIDs:    scanIDs,
			Identifier: &a.Identifier,
			Website:    &a.Website,
		}

		created, err := client.Asset.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("creating asset %q: %w", a.Name, err)
		}

		ids[a.Ref] = created.ID
		createdIDs = append(createdIDs, created.ID)
	}

	return ids, createdIDs, nil
}

// unlinkedDomainScanIDs returns the desired IDs that aren't already linked, so re-importing a
// review doesn't try to insert an edge row that's already there
func unlinkedDomainScanIDs(desired, linked []string) []string {
	return lo.Without(lo.Uniq(desired), linked...)
}

// findOrCreateDomainScanPlatforms resolves each accepted platform to a Platform ID, reusing the
// existing org Platform of the same name, and links it to the scans it was generated from and to
// whichever accepted vendors/assets the reviewer marked as belonging to it. Returns a
// ref -> Platform ID map so systems can attach to them by ref, plus the subset of IDs it created
func findOrCreateDomainScanPlatforms(ctx context.Context, client *generated.Client, ownerID string, platforms []DomainScanImportPlatform, scanIDs []string, entityIDByRef, assetIDByRef map[string]string) (map[string]string, []string, error) {
	ids := make(map[string]string, len(platforms))

	var createdIDs []string

	for _, p := range platforms {
		entityIDs := resolveDomainScanRefs(p.EntityRefs, entityIDByRef)
		assetIDs := resolveDomainScanRefs(p.AssetRefs, assetIDByRef)

		existing, err := client.Platform.Query().
			Where(platform.NameEqualFold(p.Name)).
			First(ctx)

		switch {
		case err == nil:
			ids[p.Ref] = existing.ID

			if err := relinkDomainScanPlatform(ctx, client, existing, scanIDs, entityIDs, assetIDs); err != nil {
				return nil, nil, fmt.Errorf("linking to existing platform %q: %w", p.Name, err)
			}

			continue
		case !generated.IsNotFound(err):
			return nil, nil, fmt.Errorf("looking up existing platform %q: %w", p.Name, err)
		}

		input := generated.CreatePlatformInput{
			Name:             p.Name,
			OwnerID:          &ownerID,
			GeneratedScanIDs: scanIDs,
			EntityIDs:        entityIDs,
			AssetIDs:         assetIDs,
		}

		if p.Description != "" {
			input.Description = &p.Description
		}

		created, err := client.Platform.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("creating platform %q: %w", p.Name, err)
		}

		ids[p.Ref] = created.ID
		createdIDs = append(createdIDs, created.ID)
	}

	return ids, createdIDs, nil
}

// relinkDomainScanPlatform adds only the scan/vendor/asset links an existing Platform is missing
func relinkDomainScanPlatform(ctx context.Context, client *generated.Client, existing *generated.Platform, scanIDs, entityIDs, assetIDs []string) error {
	linkedScanIDs, err := existing.QueryGeneratedScans().IDs(ctx)
	if err != nil {
		return err
	}

	linkedEntityIDs, err := existing.QueryEntities().IDs(ctx)
	if err != nil {
		return err
	}

	linkedAssetIDs, err := existing.QueryAssets().IDs(ctx)
	if err != nil {
		return err
	}

	update := client.Platform.UpdateOneID(existing.ID).
		AddGeneratedScanIDs(unlinkedDomainScanIDs(scanIDs, linkedScanIDs)...).
		AddEntityIDs(unlinkedDomainScanIDs(entityIDs, linkedEntityIDs)...).
		AddAssetIDs(unlinkedDomainScanIDs(assetIDs, linkedAssetIDs)...)

	return update.Exec(ctx)
}

// findOrCreateDomainScanSystemDetails resolves each accepted system to a SystemDetail ID, reusing
// the existing org SystemDetail of the same name, parented to the platforms it references by ref
// (required by SystemDetail's policy) and linked to that system's own subset of accepted
// vendors/assets. Returns every resolved ID plus the subset it created
func findOrCreateDomainScanSystemDetails(ctx context.Context, client *generated.Client, ownerID string, systems []DomainScanImportSystem, platformIDByRef, entityIDByRef, assetIDByRef map[string]string) ([]string, []string, error) {
	ids := make([]string, 0, len(systems))

	var createdIDs []string

	for _, system := range systems {
		platformIDs := resolveDomainScanRefs(system.PlatformRefs, platformIDByRef)
		entityIDs := resolveDomainScanRefs(system.EntityRefs, entityIDByRef)
		assetIDs := resolveDomainScanRefs(system.AssetRefs, assetIDByRef)

		existing, err := client.SystemDetail.Query().
			Where(systemdetail.SystemNameEqualFold(system.Name)).
			First(ctx)

		switch {
		case err == nil:
			ids = append(ids, existing.ID)

			if err := relinkDomainScanSystemDetail(ctx, client, existing, platformIDs, entityIDs, assetIDs); err != nil {
				return nil, nil, fmt.Errorf("linking to existing system detail %q: %w", system.Name, err)
			}

			continue
		case !generated.IsNotFound(err):
			return nil, nil, fmt.Errorf("looking up existing system detail %q: %w", system.Name, err)
		}

		input := generated.CreateSystemDetailInput{
			SystemName:  system.Name,
			OwnerID:     &ownerID,
			PlatformIDs: platformIDs,
			EntityIDs:   entityIDs,
			AssetIDs:    assetIDs,
			Description: &system.Description,
		}

		created, err := client.SystemDetail.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("creating system detail %q: %w", system.Name, err)
		}

		ids = append(ids, created.ID)
		createdIDs = append(createdIDs, created.ID)
	}

	return ids, createdIDs, nil
}

// relinkDomainScanSystemDetail adds only the platform/vendor/asset links an existing SystemDetail
// is missing
func relinkDomainScanSystemDetail(ctx context.Context, client *generated.Client, existing *generated.SystemDetail, platformIDs, entityIDs, assetIDs []string) error {
	linkedPlatformIDs, err := existing.QueryPlatforms().IDs(ctx)
	if err != nil {
		return err
	}

	linkedEntityIDs, err := existing.QueryEntities().IDs(ctx)
	if err != nil {
		return err
	}

	linkedAssetIDs, err := existing.QueryAssets().IDs(ctx)
	if err != nil {
		return err
	}

	update := client.SystemDetail.UpdateOneID(existing.ID).
		AddPlatformIDs(unlinkedDomainScanIDs(platformIDs, linkedPlatformIDs)...).
		AddEntityIDs(unlinkedDomainScanIDs(entityIDs, linkedEntityIDs)...).
		AddAssetIDs(unlinkedDomainScanIDs(assetIDs, linkedAssetIDs)...)

	return update.Exec(ctx)
}

// findOrCreateDomainScanFindings resolves each accepted finding to a Finding ID, reusing the
// existing org Finding with the same category and display name, linked to the scans it came from
// and auto-linked to any accepted asset matching its domain. Returns every resolved ID plus the
// subset it created
func findOrCreateDomainScanFindings(ctx context.Context, client *generated.Client, ownerID string, findings []DomainScanImportFinding, assets []DomainScanImportAsset, assetIDByRef map[string]string, scanIDs []string) ([]string, []string, error) {
	ids := make([]string, 0, len(findings))

	var createdIDs []string

	// a single-domain review (the common case) has exactly one scan behind it, so its target
	// stands in for any finding that didn't carry its own domain
	fallbackDomain := domainScanSingleTarget(ctx, client, scanIDs)

	for _, f := range findings {
		domain := strings.ToLower(strings.TrimSuffix(f.Domain, "."))
		if domain == "" {
			domain = fallbackDomain
		}

		matchedAssetIDs := domainScanMappingMatches(assets, assetIDByRef, func(assetDomain string) bool {
			return domain != "" && (assetDomain == domain || strings.HasSuffix(assetDomain, "."+domain))
		})

		displayName := f.Category
		if domain != "" {
			displayName = fmt.Sprintf("%s (%s)", f.Category, domain)
		}

		existing, err := client.Finding.Query().
			Where(
				finding.Category(f.Category),
				finding.DisplayNameEqualFold(displayName),
			).
			First(ctx)

		switch {
		case err == nil:
			ids = append(ids, existing.ID)

			if err := relinkDomainScanFinding(ctx, client, existing, scanIDs, matchedAssetIDs); err != nil {
				return nil, nil, fmt.Errorf("linking to existing finding %q: %w", displayName, err)
			}

			continue
		case !generated.IsNotFound(err):
			return nil, nil, fmt.Errorf("looking up existing finding %q: %w", displayName, err)
		}

		input := generated.CreateFindingInput{
			OwnerID:     &ownerID,
			ScanIDs:     scanIDs,
			AssetIDs:    matchedAssetIDs,
			Category:    &f.Category,
			Description: &f.Description,
			Severity:    &f.Severity,
			DisplayName: &displayName,
		}

		created, err := client.Finding.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("creating finding: %w", err)
		}

		ids = append(ids, created.ID)
		createdIDs = append(createdIDs, created.ID)
	}

	return ids, createdIDs, nil
}

// relinkDomainScanFinding adds only the scan/asset links an existing Finding is missing
func relinkDomainScanFinding(ctx context.Context, client *generated.Client, existing *generated.Finding, scanIDs, assetIDs []string) error {
	linkedScanIDs, err := existing.QueryScans().IDs(ctx)
	if err != nil {
		return err
	}

	linkedAssetIDs, err := existing.QueryAssets().IDs(ctx)
	if err != nil {
		return err
	}

	update := client.Finding.UpdateOneID(existing.ID).
		AddScanIDs(unlinkedDomainScanIDs(scanIDs, linkedScanIDs)...).
		AddAssetIDs(unlinkedDomainScanIDs(assetIDs, linkedAssetIDs)...)

	return update.Exec(ctx)
}

// domainScanSingleTarget returns the target domain when scanIDs resolves to exactly one scan -
// ambiguous for a multi-domain review, where it returns ""
func domainScanSingleTarget(ctx context.Context, client *generated.Client, scanIDs []string) string {
	if len(scanIDs) != 1 {
		return ""
	}

	scanRecord, err := client.Scan.Get(ctx, scanIDs[0])
	if err != nil {
		return ""
	}

	return scanRecord.Target
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
