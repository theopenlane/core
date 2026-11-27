package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/v2/internal/ent/generated/platform"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/ent/generated/systemdetail"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/v2/internal/ent/validator"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/pkg/domainscan"
	"github.com/theopenlane/core/v2/pkg/urlx"
	"github.com/theopenlane/utils/jsonx"
)

// domainScanImportVendorEntityType is the entityTypeName to use when creating the entities from
// a completed scan
const domainScanImportVendorEntityType = "vendor"

// domainScanImportVariant is the mapping variant the domain scan import ingests its records under
const domainScanImportVariant = "DOMAIN_SCAN"

// domainScanFindingSource is the finding source marking findings produced by a domain scan
const domainScanFindingSource = "domain_scan"

// domainScanImportContracts are the schemas the domain scan import ingests, in dependency order
var domainScanImportContracts = []types.IngestContract{
	{Schema: entityops.SchemaAsset.Name},
	{Schema: entityops.SchemaEntity.Name},
	{Schema: entityops.SchemaPlatform.Name},
	{Schema: entityops.SchemaSystemDetail.Name},
	{Schema: entityops.SchemaFinding.Name},
}

// assetKeysField is the payload key listing the source identifiers of the assets a record links to
const assetKeysField = "asset_keys"

// domainScanImportMappings returns the import's identity mappings, whose payloads list link target keys resolved to ids at ingest
func domainScanImportMappings() []types.MappingRegistration {
	// assets another sync wrote carry its external id, so assets link by source identifier instead
	toAssets := types.LinkRule{TargetSchema: entityops.SchemaAsset.Name, Edge: "assets", TargetField: asset.FieldSourceIdentifier, SourceList: assetKeysField, PayloadKeys: true}

	platformLinks := append(providerkit.EdgeKeyLinks(entityops.SchemaPlatform, platform.EdgeEntities), toAssets)
	systemLinks := append(providerkit.EdgeKeyLinks(entityops.SchemaSystemDetail, systemdetail.EdgePlatforms, systemdetail.EdgeEntities), toAssets)

	return []types.MappingRegistration{
		domainScanImportMapping(entityops.SchemaAsset.Name),
		domainScanImportMapping(entityops.SchemaEntity.Name, toAssets),
		domainScanImportMapping(entityops.SchemaPlatform.Name, platformLinks...),
		domainScanImportMapping(entityops.SchemaSystemDetail.Name, systemLinks...),
		domainScanImportMapping(entityops.SchemaFinding.Name, toAssets),
	}
}

// domainScanImportMapping registers one schema's domain scan variant, passing its payload through as-is
func domainScanImportMapping(schema string, links ...types.LinkRule) types.MappingRegistration {
	return types.MappingRegistration{
		Schema:  schema,
		Variant: domainScanImportVariant,
		Spec:    types.MappingOverride{FilterExpr: "true", Links: links},
	}
}

// domainScanImportPayloadSets builds the report's payload sets, assets and vendors first so later records can link to them
func domainScanImportPayloadSets(ctx context.Context, client *generated.Client, cfg DomainScanImport) ([]types.IngestPayloadSet, error) {
	vendorTypeID, err := client.EntityType.Query().Where(entitytype.NameEqualFold(domainScanImportVendorEntityType)).OnlyID(ctx)
	if err != nil {
		return nil, fmt.Errorf("looking up vendor entity type: %w", err)
	}

	vendorNameByAssetDomain, err := domainScanAssetVendorNames(ctx, client, cfg.ScanIDs)
	if err != nil {
		return nil, err
	}

	assets, err := domainScanAssetEnvelopes(cfg)
	if err != nil {
		return nil, err
	}

	vendors, err := domainScanVendorEnvelopes(cfg, vendorTypeID, vendorNameByAssetDomain)
	if err != nil {
		return nil, err
	}

	platforms, err := domainScanPlatformEnvelopes(cfg)
	if err != nil {
		return nil, err
	}

	systems, err := domainScanSystemEnvelopes(cfg)
	if err != nil {
		return nil, err
	}

	findings, err := domainScanFindingEnvelopes(cfg, domainScanHostname(domainScanSingleTarget(ctx, client, cfg.ScanIDs)))
	if err != nil {
		return nil, err
	}

	return []types.IngestPayloadSet{
		{Schema: entityops.SchemaAsset.Name, Envelopes: assets},
		{Schema: entityops.SchemaEntity.Name, Envelopes: vendors},
		{Schema: entityops.SchemaPlatform.Name, Envelopes: platforms},
		{Schema: entityops.SchemaSystemDetail.Name, Envelopes: systems},
		{Schema: entityops.SchemaFinding.Name, Envelopes: findings},
	}, nil
}

// domainScanImportDetail is the notification detail saying what else the import did and how to retry an incomplete one
func domainScanImportDetail(brandingUpdated, incomplete bool) string {
	return strings.Join(lo.Compact([]string{
		lo.Ternary(brandingUpdated, "updated trust center branding", ""),
		lo.Ternary(incomplete, "import the review again to retry", ""),
	}), ", ")
}

// domainScanAssetEnvelopes builds one asset payload per accepted asset, keyed by its source identifier
func domainScanAssetEnvelopes(cfg DomainScanImport) ([]types.MappingEnvelope, error) {
	envelopes := make([]types.MappingEnvelope, 0, len(cfg.Assets))

	for _, a := range cfg.Assets {
		key := domainScanAssetKey(a)

		envelope, err := providerkit.MarshalEnvelopeVariant(domainScanImportVariant, key, generated.CreateAssetInput{
			Name:             a.Name,
			Categories:       a.Categories,
			ScanIDs:          cfg.ScanIDs,
			Identifier:       lo.EmptyableToPtr(a.Identifier),
			SourceIdentifier: &key,
			Website:          lo.EmptyableToPtr(a.Website),
		}, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return envelopes, nil
}

// domainScanVendorEnvelopes builds one vendor payload per accepted vendor, linking the assets the scan attributed to it
func domainScanVendorEnvelopes(cfg DomainScanImport, vendorTypeID string, vendorNameByAssetDomain map[string]string) ([]types.MappingEnvelope, error) {
	envelopes := make([]types.MappingEnvelope, 0, len(cfg.Vendors))

	for _, vendor := range cfg.Vendors {
		key := domainScanVendorKey(vendor)

		envelope, err := domainScanEnvelope(key, generated.CreateEntityInput{
			Name:           new(domainScanVendorName(vendor)),
			DisplayName:    &vendor.Name,
			ExternalID:     &key,
			Domains:        lo.Compact([]string{domainScanHostname(vendor.Domain)}),
			ApprovedForUse: new(true),
			EntityTypeID:   &vendorTypeID,
			ScanIDs:        cfg.ScanIDs,
			Tags:           vendor.Categories,
		}, map[string][]string{assetKeysField: domainScanVendorAssetKeys(cfg.Assets, vendorNameByAssetDomain, vendor)})
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return envelopes, nil
}

// domainScanPlatformEnvelopes builds one platform payload per accepted platform, keyed by name and linked to
// the scans it came from and the accepted vendors and assets it references
func domainScanPlatformEnvelopes(cfg DomainScanImport) ([]types.MappingEnvelope, error) {
	assetKeys := domainScanAssetKeysByRef(cfg)
	envelopes := make([]types.MappingEnvelope, 0, len(cfg.Platforms))

	for _, p := range cfg.Platforms {
		envelope, err := domainScanEnvelope(p.Name, generated.CreatePlatformInput{
			Name:             p.Name,
			Description:      lo.EmptyableToPtr(p.Description),
			GeneratedScanIDs: cfg.ScanIDs,
		}, lo.Assign(domainScanVendorLinkKeys(entityops.SchemaPlatform, platform.EdgeEntities, cfg, p.EntityRefs), map[string][]string{
			assetKeysField: resolveDomainScanRefs(p.AssetRefs, assetKeys),
		}))
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return envelopes, nil
}

// domainScanSystemEnvelopes builds one system detail payload per accepted system, keyed by name and parented
// to the platforms it references, as SystemDetail's policy requires
func domainScanSystemEnvelopes(cfg DomainScanImport) ([]types.MappingEnvelope, error) {
	assetKeys := domainScanAssetKeysByRef(cfg)
	platformNames := lo.SliceToMap(cfg.Platforms, func(p DomainScanImportPlatform) (string, string) { return p.Ref, p.Name })
	envelopes := make([]types.MappingEnvelope, 0, len(cfg.Systems))

	for _, system := range cfg.Systems {
		envelope, err := domainScanEnvelope(system.Name, generated.CreateSystemDetailInput{
			SystemName:  system.Name,
			Description: lo.EmptyableToPtr(system.Description),
		}, lo.Assign(domainScanVendorLinkKeys(entityops.SchemaSystemDetail, systemdetail.EdgeEntities, cfg, system.EntityRefs), map[string][]string{
			providerkit.EdgeNameKey(entityops.SchemaSystemDetail, systemdetail.EdgePlatforms): resolveDomainScanRefs(system.PlatformRefs, platformNames),
			assetKeysField: resolveDomainScanRefs(system.AssetRefs, assetKeys),
		}))
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return envelopes, nil
}

// domainScanEnvelope encodes a create input under the import variant, adding each non-empty list of link keys
func domainScanEnvelope(key string, input any, links map[string][]string) (types.MappingEnvelope, error) {
	envelope, err := providerkit.MarshalEnvelopeVariant(domainScanImportVariant, key, input, ErrPayloadEncode)
	if err != nil {
		return envelope, err
	}

	for field, keys := range links {
		if len(keys) == 0 {
			continue
		}

		if envelope.Payload, _, err = jsonx.SetObjectKey(envelope.Payload, field, keys); err != nil {
			return envelope, ErrPayloadEncode
		}
	}

	return envelope, nil
}

// domainScanFindingEnvelopes builds one finding payload per accepted finding, linked to the accepted assets under its domain
func domainScanFindingEnvelopes(cfg DomainScanImport, fallbackDomain string) ([]types.MappingEnvelope, error) {
	envelopes := make([]types.MappingEnvelope, 0, len(cfg.Findings))

	for _, f := range cfg.Findings {
		domain := lo.CoalesceOrEmpty(domainScanHostname(f.Domain), fallbackDomain)

		displayName := f.Category
		if domain != "" {
			displayName = fmt.Sprintf("%s (%s)", f.Category, domain)
		}

		// scans give findings no id of their own, so a resubmitted review matches on this one
		externalID := domainScanFindingSource + ":" + displayName

		envelope, err := domainScanEnvelope(externalID, generated.CreateFindingInput{
			ExternalID:  &externalID,
			Source:      new(domainScanFindingSource),
			DisplayName: &displayName,
			Category:    lo.EmptyableToPtr(f.Category),
			Description: lo.EmptyableToPtr(f.Description),
			Severity:    lo.EmptyableToPtr(f.Severity),
			ScanIDs:     cfg.ScanIDs,
		}, map[string][]string{assetKeysField: domainScanAssetKeys(cfg.Assets, func(assetDomain string) bool {
			return domain != "" && (assetDomain == domain || strings.HasSuffix(assetDomain, "."+domain))
		})})
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return envelopes, nil
}

// domainScanHostname normalizes a hostname, empty when the value is not one
func domainScanHostname(value string) string {
	hostname, err := urlx.NormalizeHostname(value)
	if err != nil {
		return ""
	}

	return hostname
}

// domainScanAssetKey is the source identifier an accepted asset is upserted and linked by
func domainScanAssetKey(a DomainScanImportAsset) string {
	return lo.CoalesceOrEmpty(domainScanHostname(a.Identifier), strings.TrimSpace(a.Identifier), strings.TrimSpace(a.Name))
}

// domainScanVendorKey is the external id an accepted vendor is upserted and linked by, its domain or else its name
func domainScanVendorKey(vendor DomainScanImportVendor) string {
	return lo.CoalesceOrEmpty(domainScanHostname(vendor.Domain), strings.ToLower(strings.TrimSpace(vendor.Name)))
}

// domainScanVendorName is the name an accepted vendor is created and linked by
func domainScanVendorName(vendor DomainScanImportVendor) string {
	return lo.CoalesceOrEmpty(vendor.LegalName, vendor.Name)
}

// domainScanVendorLinkKeys lists the external ids and names linking schema's edge to the referenced vendors
func domainScanVendorLinkKeys(schema *entityops.Schema, edge string, cfg DomainScanImport, refs []string) map[string][]string {
	keys := lo.SliceToMap(cfg.Vendors, func(v DomainScanImportVendor) (string, string) { return v.Ref, domainScanVendorKey(v) })
	names := lo.SliceToMap(cfg.Vendors, func(v DomainScanImportVendor) (string, string) { return v.Ref, domainScanVendorName(v) })

	return map[string][]string{
		providerkit.EdgeExternalIDKey(schema, edge): resolveDomainScanRefs(refs, keys),
		providerkit.EdgeNameKey(schema, edge):       resolveDomainScanRefs(refs, names),
	}
}

// domainScanAssetKeysByRef maps each accepted asset's ref to its link key
func domainScanAssetKeysByRef(cfg DomainScanImport) map[string]string {
	return lo.SliceToMap(cfg.Assets, func(a DomainScanImportAsset) (string, string) { return a.Ref, domainScanAssetKey(a) })
}

// domainScanVendorAssetKeys returns the keys of accepted assets the scan itself attributed to the vendor
func domainScanVendorAssetKeys(assets []DomainScanImportAsset, vendorNameByAssetDomain map[string]string, vendor DomainScanImportVendor) []string {
	return domainScanAssetKeys(assets, func(assetDomain string) bool {
		registrable, ok := domainscan.RegistrableDomain(assetDomain)
		if !ok {
			return false
		}

		return strings.EqualFold(vendorNameByAssetDomain[registrable], vendor.Name)
	})
}

// domainScanAssetKeys returns the link keys of accepted assets whose lowercased, dot-trimmed
// identifier satisfies match
func domainScanAssetKeys(assets []DomainScanImportAsset, match func(assetDomain string) bool) []string {
	var keys []string

	for _, a := range assets {
		assetDomain := domainScanHostname(a.Identifier)
		if assetDomain == "" || !match(assetDomain) {
			continue
		}

		keys = append(keys, domainScanAssetKey(a))
	}

	return keys
}

func applyBrandingToTrustCenter(ctx context.Context, client *generated.Client, brandDesign domainscan.BrandDesignProfile) (bool, error) {
	envs := []enums.TrustCenterEnvironment{}

	if brandDesign.ApplyToPreviewTrustcenter {
		envs = append(envs, enums.TrustCenterEnvironmentPreview)
	}

	if brandDesign.ApplyToLiveTrustcenter {
		envs = append(envs, enums.TrustCenterEnvironmentLive)
	}

	if len(envs) == 0 {
		return false, nil
	}

	return workflows.WithTx(ctx, client, nil, func(ctx context.Context, tx *generated.Tx) (bool, error) {
		tc, err := tx.TrustCenter.Query().First(ctx)
		if generated.IsNotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, fmt.Errorf("could not find trustcenter for brand design: %w", err)
		}

		var isUpdated bool

		settings, err := tx.TrustCenterSetting.Query().
			Where(
				trustcentersetting.TrustCenterIDEQ(tc.ID),
				trustcentersetting.EnvironmentIn(envs...)).
			All(ctx)

		if err != nil {
			return false, fmt.Errorf("could not fetch trust center settings for brand design: %w", err)
		}

		for _, setting := range settings {

			ok, err := updateTrustcenterBrandDesignSetting(ctx, setting, brandDesign)
			if err != nil {
				return false, err
			}

			isUpdated = isUpdated || ok
		}

		return isUpdated, nil
	})
}

func chooseTrustCenterBrandColor(newColor, existingColor string) (string, bool) {
	color := strings.TrimSpace(newColor)
	if color == "" || validator.HexColorValidator(color) != nil {
		return existingColor, false
	}

	return color, color != existingColor
}

func updateTrustcenterBrandDesignSetting(ctx context.Context, setting *generated.TrustCenterSetting, brandDesign domainscan.BrandDesignProfile) (bool, error) {
	input := generated.UpdateTrustCenterSettingInput{}
	isChanged := false

	primaryColor, shouldUpdate := chooseTrustCenterBrandColor(brandDesign.PrimaryColor, setting.PrimaryColor)
	if shouldUpdate {
		input.PrimaryColor = &primaryColor
		isChanged = true
	}

	foregroundColor, shouldUpdate := chooseTrustCenterBrandColor(brandDesign.ForegroundColor, setting.ForegroundColor)
	if shouldUpdate {
		input.ForegroundColor = &foregroundColor
		isChanged = true
	}

	backgroundColor, shouldUpdate := chooseTrustCenterBrandColor(brandDesign.BackgroundColor, setting.BackgroundColor)
	if shouldUpdate {
		input.BackgroundColor = &backgroundColor
		isChanged = true
	}

	accentColor, shouldUpdate := chooseTrustCenterBrandColor(brandDesign.AccentColor, setting.AccentColor)
	if shouldUpdate {
		input.AccentColor = &accentColor
		isChanged = true
	}

	secondaryBackgroundColor, shouldUpdate := chooseTrustCenterBrandColor(brandDesign.SecondaryBackgroundColor, setting.SecondaryBackgroundColor)
	if shouldUpdate {
		input.SecondaryBackgroundColor = &secondaryBackgroundColor
		isChanged = true
	}

	secondaryForegroundColor, shouldUpdate := chooseTrustCenterBrandColor(brandDesign.SecondaryForegroundColor, setting.SecondaryForegroundColor)
	if shouldUpdate {
		input.SecondaryForegroundColor = &secondaryForegroundColor
		isChanged = true
	}

	palettes := []string{
		brandDesign.PrimaryColor,
		brandDesign.ForegroundColor,
		brandDesign.BackgroundColor,
		brandDesign.AccentColor,
		brandDesign.SecondaryBackgroundColor,
		brandDesign.SecondaryForegroundColor,
	}

	// check if this new palette imports have at least one valid color code
	hasAcceptedColor := lo.SomeBy(palettes, func(color string) bool {
		color = strings.TrimSpace(color)
		return color != "" && validator.HexColorValidator(color) == nil
	})

	if font := strings.TrimSpace(brandDesign.Font); font != "" && font != setting.Font {
		input.Font = &font
		isChanged = true
	}

	if setting.LogoLocalFileID == nil {
		if logoURL := brandDesign.LogoURL; validator.ValidateURL()(logoURL) == nil {
			if setting.LogoRemoteURL == nil || logoURL != *setting.LogoRemoteURL {
				input.LogoRemoteURL = &logoURL
				isChanged = true
			}
		}
	}

	if setting.FaviconLocalFileID == nil {
		if faviconURL := brandDesign.FaviconURL; validator.ValidateURL()(faviconURL) == nil {
			if setting.FaviconRemoteURL == nil || faviconURL != *setting.FaviconRemoteURL {
				input.FaviconRemoteURL = &faviconURL
				isChanged = true
			}
		}
	}

	newPalettes := []string{
		primaryColor,
		foregroundColor,
		backgroundColor,
		accentColor,
		secondaryBackgroundColor,
		secondaryForegroundColor,
	}

	hasCompletePalette := lo.EveryBy(newPalettes, func(color string) bool {
		return validator.HexColorValidator(color) == nil
	})

	// only convert to advanced mode if we have the complete color palette
	if hasAcceptedColor && hasCompletePalette {
		if setting.ThemeMode != enums.TrustCenterThemeModeAdvanced {
			input.ThemeMode = new(enums.TrustCenterThemeModeAdvanced)
			isChanged = true
		}
	}

	if !isChanged {
		return false, nil
	}

	_, err := setting.Update().SetInput(input).Save(ctx)
	if err != nil {
		return false, fmt.Errorf("updating %s trust center brand design: %w", setting.Environment, err)
	}

	return true, nil
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
