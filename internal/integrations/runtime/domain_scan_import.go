package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// domainScanImportVendorEntityType is the EntityType.Name assigned to vendors created from an accepted domain scan review
const domainScanImportVendorEntityType = "vendor"

// importSummary counts what an accepted domain scan review created, for the follow-up notification
type importSummary struct {
	PlatformIDs     []string `json:"platform_ids,omitempty"`
	SystemDetailIDs []string `json:"system_detail_ids,omitempty"`
	EntityIDs       []string `json:"entity_ids,omitempty"`
	AssetIDs        []string `json:"asset_ids,omitempty"`
	FindingIDs      []string `json:"finding_ids,omitempty"`
}

// HandleImportDomainScanReview creates the real Platform/SystemDetail/Entity/Asset/Finding
// records for a reviewer-accepted domain scan report, all under one transaction, then sends a
// follow-up Notification once done - the GraphQL mutation that emitted this envelope already
// returned before this ran, so this is how the caller learns the import actually finished and
// what got created
func (r *Runtime) HandleImportDomainScanReview(ctx context.Context, envelope operations.ImportDomainScanReviewEnvelope) error {
	ctx = logx.WithFields(ctx, logx.LogFields{
		"organization_id": envelope.OrganizationID,
		"scan_ids":        envelope.ScanIDs,
	})

	systemCtx := domainScanSystemContext(ctx, envelope.OrganizationID)

	summary, err := workflows.WithTx(systemCtx, r.DB(), nil, func(tx *generated.Tx) (importSummary, error) {
		return importDomainScanReview(systemCtx, tx.Client(), envelope)
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed importing accepted review")
		return err
	}

	return r.notifyDomainScanImportComplete(systemCtx, envelope.OrganizationID, summary)
}

// importDomainScanReview creates every accepted object in order: vendors and assets first (so
// their real IDs are known), then the platform and its system details (which link back to the
// vendors/assets via ref), then findings
func importDomainScanReview(ctx context.Context, client *generated.Client, envelope operations.ImportDomainScanReviewEnvelope) (importSummary, error) {
	var summary importSummary

	if len(envelope.Systems) > 0 && len(envelope.Platforms) == 0 {
		return summary, fmt.Errorf("%w: systems were submitted without any platform to attach them to", errDomainScanImportInvalid)
	}

	// assets are created before vendors so a vendor whose domain matches an asset's identifier
	// can be auto-linked to it on creation, without the reviewer having to draw that link explicitly
	assetIDByRef, err := createDomainScanAssets(ctx, client, envelope.OrganizationID, envelope.Assets, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	for _, id := range assetIDByRef {
		summary.AssetIDs = append(summary.AssetIDs, id)
	}

	entityIDByRef, err := findOrCreateDomainScanVendors(ctx, client, envelope.OrganizationID, envelope.Vendors, envelope.Assets, assetIDByRef, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	for _, id := range entityIDByRef {
		summary.EntityIDs = append(summary.EntityIDs, id)
	}

	if len(envelope.Platforms) > 0 {
		platformIDByRef, err := createDomainScanPlatforms(ctx, client, envelope.OrganizationID, envelope.Platforms, envelope.ScanIDs, entityIDByRef, assetIDByRef)
		if err != nil {
			return summary, err
		}

		for _, id := range platformIDByRef {
			summary.PlatformIDs = append(summary.PlatformIDs, id)
		}

		summary.SystemDetailIDs, err = createDomainScanSystemDetails(ctx, client, envelope.OrganizationID, envelope.Systems, platformIDByRef, entityIDByRef, assetIDByRef)
		if err != nil {
			return summary, err
		}
	}

	summary.FindingIDs, err = createDomainScanFindings(ctx, client, envelope.OrganizationID, envelope.Findings, envelope.ScanIDs)
	if err != nil {
		return summary, err
	}

	return summary, nil
}

// errDomainScanImportInvalid is returned when an accepted domain scan review is structurally invalid
var errDomainScanImportInvalid = fmt.Errorf("domain scan import: invalid review")

// findOrCreateDomainScanVendors resolves each accepted vendor to an Entity ID, reusing an
// existing org Entity by name if one already exists, and auto-links any accepted asset whose
// domain belongs to the vendor
func findOrCreateDomainScanVendors(ctx context.Context, client *generated.Client, ownerID string, vendors []operations.ImportDomainScanReviewVendor, assets []operations.ImportDomainScanReviewAsset, assetIDByRef map[string]string, scanIDs []string) (map[string]string, error) {
	ids := make(map[string]string, len(vendors))
	if len(vendors) == 0 {
		return ids, nil
	}

	vendorEntityType, err := client.EntityType.Query().Where(entitytype.NameEqualFold(domainScanImportVendorEntityType), entitytype.OwnerID(ownerID)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("looking up vendor entity type: %w", err)
	}

	for _, vendor := range vendors {
		matchedAssetIDs := domainScanVendorAssetMatches(vendor.Domain, assets, assetIDByRef)

		existing, err := client.Entity.Query().
			Where(
				entity.OwnerID(ownerID),
				entity.Or(entity.NameEqualFold(vendor.Name), entity.DisplayNameEqualFold(vendor.Name)),
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
			Name:           &vendor.Name,
			OwnerID:        &ownerID,
			ApprovedForUse: lo.ToPtr(true),
			EntityTypeID:   &vendorEntityType.ID,
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

// domainScanVendorAssetMatches returns asset IDs whose domain matches or is a subdomain of vendorDomain
func domainScanVendorAssetMatches(vendorDomain string, assets []operations.ImportDomainScanReviewAsset, assetIDByRef map[string]string) []string {
	if vendorDomain == "" {
		return nil
	}

	vendorDomain = strings.ToLower(strings.TrimSuffix(vendorDomain, "."))

	var ids []string

	for _, asset := range assets {
		if asset.Identifier == "" {
			continue
		}

		assetDomain := strings.ToLower(strings.TrimSuffix(asset.Identifier, "."))
		if assetDomain != vendorDomain && !strings.HasSuffix(assetDomain, "."+vendorDomain) {
			continue
		}

		if id, ok := assetIDByRef[asset.Ref]; ok {
			ids = append(ids, id)
		}
	}

	return ids
}

// createDomainScanAssets creates one Asset per accepted asset, returning a ref -> Asset ID map
func createDomainScanAssets(ctx context.Context, client *generated.Client, ownerID string, assets []operations.ImportDomainScanReviewAsset, scanIDs []string) (map[string]string, error) {
	ids := make(map[string]string, len(assets))

	for _, asset := range assets {
		input := generated.CreateAssetInput{
			Name:       asset.Name,
			OwnerID:    &ownerID,
			Categories: asset.Categories,
			ScanIDs:    scanIDs,
		}

		if asset.Identifier != "" {
			input.Identifier = &asset.Identifier
		}

		if asset.Website != "" {
			input.Website = &asset.Website
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
		if len(platformIDs) == 0 {
			return nil, fmt.Errorf("%w: system %q has no platform to attach to", errDomainScanImportInvalid, system.Name)
		}

		input := generated.CreateSystemDetailInput{
			SystemName:  system.Name,
			OwnerID:     &ownerID,
			PlatformIDs: platformIDs,
			EntityIDs:   resolveDomainScanRefs(system.EntityRefs, entityIDByRef),
			AssetIDs:    resolveDomainScanRefs(system.AssetRefs, assetIDByRef),
		}

		if system.Description != "" {
			input.Description = &system.Description
		}

		created, err := client.SystemDetail.Create().SetInput(input).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating system detail %q: %w", system.Name, err)
		}

		ids = append(ids, created.ID)
	}

	return ids, nil
}

// createDomainScanFindings creates one Finding per accepted finding, linked to the scans it came from
func createDomainScanFindings(ctx context.Context, client *generated.Client, ownerID string, findings []operations.ImportDomainScanReviewFinding, scanIDs []string) ([]string, error) {
	ids := make([]string, 0, len(findings))

	for _, finding := range findings {
		input := generated.CreateFindingInput{
			OwnerID: &ownerID,
			ScanIDs: scanIDs,
		}

		if finding.Category != "" {
			input.Category = &finding.Category
		}

		if finding.Description != "" {
			input.Description = &finding.Description
		}

		if finding.Severity != "" {
			input.Severity = &finding.Severity
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

// notifyDomainScanImportComplete sends the Notification that surfaces an import's results, since
// the GraphQL mutation that triggered it already returned before this ran
func (r *Runtime) notifyDomainScanImportComplete(ctx context.Context, organizationID string, summary importSummary) error {
	data, err := jsonx.ToMap(summary)
	if err != nil {
		return err
	}

	body := fmt.Sprintf(
		"Imported %d vendor(s), %d asset(s), %d system(s), and %d finding(s)",
		len(summary.EntityIDs), len(summary.AssetIDs), len(summary.SystemDetailIDs), len(summary.FindingIDs),
	)

	_, err = r.DB().Notification.Create().
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
