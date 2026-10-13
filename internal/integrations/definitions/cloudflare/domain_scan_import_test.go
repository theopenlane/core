package cloudflare

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/domainscan"
)

func TestChooseTrustCenterBrandColor(t *testing.T) {
	tt := []struct {
		name          string
		newColor      string
		existingColor string
		want          string
		isChanged     bool
	}{
		{
			name:          "accepts a new valid color",
			newColor:      "  #112233  ",
			existingColor: "#000000",
			want:          "#112233",
			isChanged:     true,
		},
		{
			name:          "valid unchanged color",
			newColor:      "#112233",
			existingColor: "#112233",
			want:          "#112233",
		},
		{
			name:          "empty color retains existing value",
			newColor:      "   ",
			existingColor: "#112233",
			want:          "#112233",
		},
		{
			name:          "invalid color retains existing value",
			newColor:      "blue",
			existingColor: "#112233",
			want:          "#112233",
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			color, ok := chooseTrustCenterBrandColor(tt.newColor, tt.existingColor)

			assert.Equal(t, color, tt.want)
			assert.Equal(t, ok, tt.isChanged)
		})
	}
}

func TestUpdateTrustcenterBrandDesignSetting(t *testing.T) {
	setting := &generated.TrustCenterSetting{
		ThemeMode:                enums.TrustCenterThemeModeAdvanced,
		PrimaryColor:             "#000001",
		ForegroundColor:          "#ffffff",
		BackgroundColor:          "#000000",
		AccentColor:              "#000005",
		SecondaryBackgroundColor: "#F54927",
		SecondaryForegroundColor: "#000006",
		Font:                     "Outfit",
	}

	changed, err := updateTrustcenterBrandDesignSetting(context.Background(), setting, domainscan.BrandDesignProfile{
		PrimaryColor: "invalid",
		Font:         "   ",
		LogoURL:      "not-a-url",
		FaviconURL:   "not-a-url",
	})

	assert.NilError(t, err)
	assert.Equal(t, changed, false)
}

func TestResolveDomainScanRefs(t *testing.T) {
	idByRef := map[string]string{
		"vendor-1": "entity-abc",
		"vendor-2": "entity-def",
	}

	tests := []struct {
		name string
		refs []string
		want []string
	}{
		{
			name: "resolves known refs in order",
			refs: []string{"vendor-1", "vendor-2"},
			want: []string{"entity-abc", "entity-def"},
		},
		{
			name: "drops unresolved refs rather than erroring",
			refs: []string{"vendor-1", "does-not-exist"},
			want: []string{"entity-abc"},
		},
		{
			name: "no refs returns empty slice",
			refs: nil,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveDomainScanRefs(tt.refs, idByRef)

			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestDomainScanAssetKeys(t *testing.T) {
	assets := []DomainScanImportAsset{
		{Ref: "a1", Identifier: "CDN.iubenda.com."},
		{Ref: "a2", Identifier: "iubenda.com"},
		{Ref: "a3", Identifier: "theopenlane.io"},
		{Ref: "a4", Identifier: ""},
	}

	tests := []struct {
		name   string
		domain string
		want   []string
	}{
		{
			name:   "matches subdomain and exact domain",
			domain: "iubenda.com",
			want:   []string{"cdn.iubenda.com", "iubenda.com"},
		},
		{
			name:   "no match for unrelated domain",
			domain: "theopenlane.io",
			want:   []string{"theopenlane.io"},
		},
		{
			name:   "empty domain matches nothing",
			domain: "",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain := tt.domain
			got := domainScanAssetKeys(assets, func(assetDomain string) bool {
				return domain != "" && (assetDomain == domain || strings.HasSuffix(assetDomain, "."+domain))
			})

			assert.Check(t, cmp.DeepEqual(got, tt.want))
		})
	}
}

func TestDomainScanVendorEnvelopes(t *testing.T) {
	cfg := DomainScanImport{
		ScanIDs: []string{"scan-1"},
		Vendors: []DomainScanImportVendor{
			{Ref: "v1", Name: "Acme", Domain: "Acme.com."},
			{Ref: "v2", Name: "Globex", LegalName: "Globex Corp", Domain: "globex.com"},
			{Ref: "v3", Name: "No Domain"},
		},
	}

	envelopes, err := domainScanVendorEnvelopes(cfg, "type-vendor", nil, map[string]string{"globex.com": "entity-globex"})
	assert.NilError(t, err)
	assert.Check(t, cmp.Len(envelopes, 3))

	assert.Check(t, cmp.Equal(envelopes[0].Variant, domainScanImportVariant))
	assert.Check(t, cmp.Equal(entityops.FieldValue(envelopes[0].Payload, "external_id"), "acme.com"))
	assert.Check(t, cmp.Equal(entityops.FieldValue(envelopes[0].Payload, entityops.FieldIngestTargetID), ""))
	assert.Check(t, strings.Contains(string(envelopes[0].Payload), `"domains":["acme.com"]`))

	assert.Check(t, cmp.Equal(entityops.FieldValue(envelopes[1].Payload, entityops.FieldIngestTargetID), "entity-globex"))
	assert.Check(t, cmp.Equal(entityops.FieldValue(envelopes[1].Payload, "name"), "Globex Corp"))
	assert.Check(t, !strings.Contains(string(envelopes[1].Payload), `"domains"`))

	assert.Check(t, cmp.Equal(entityops.FieldValue(envelopes[2].Payload, "external_id"), "no domain"))
}

func TestDomainScanPlatformAndSystemEnvelopesCarryLinkKeys(t *testing.T) {
	cfg := DomainScanImport{
		ScanIDs: []string{"scan-1"},
		Vendors: []DomainScanImportVendor{{Ref: "v1", Name: "Acme", Domain: "acme.com"}},
		Assets:  []DomainScanImportAsset{{Ref: "a1", Name: "app", Identifier: "App.Acme.com"}},
		Platforms: []DomainScanImportPlatform{
			{Ref: "p1", Name: "Production", EntityRefs: []string{"v1", "missing"}, AssetRefs: []string{"a1"}},
		},
		Systems: []DomainScanImportSystem{
			{Name: "Billing", PlatformRefs: []string{"p1"}, EntityRefs: []string{"v1"}},
		},
	}

	platforms, err := domainScanPlatformEnvelopes(cfg)
	assert.NilError(t, err)
	assert.Check(t, cmp.Len(platforms, 1))
	assert.Check(t, strings.Contains(string(platforms[0].Payload), `"entity_ids":["acme.com"]`))
	assert.Check(t, strings.Contains(string(platforms[0].Payload), `"asset_ids":["app.acme.com"]`))
	assert.Check(t, strings.Contains(string(platforms[0].Payload), `"generated_scan_ids":["scan-1"]`))

	systems, err := domainScanSystemEnvelopes(cfg)
	assert.NilError(t, err)
	assert.Check(t, cmp.Len(systems, 1))
	assert.Check(t, strings.Contains(string(systems[0].Payload), `"platform_ids":["Production"]`))
	assert.Check(t, strings.Contains(string(systems[0].Payload), `"entity_ids":["acme.com"]`))
}

func TestDomainScanFindingEnvelopes(t *testing.T) {
	cfg := DomainScanImport{
		ScanIDs: []string{"scan-1"},
		Assets:  []DomainScanImportAsset{{Ref: "a1", Name: "app", Identifier: "app.acme.com"}},
		Findings: []DomainScanImportFinding{
			{Category: "Missing DMARC", Domain: "Acme.com."},
			{Category: "Weak TLS"},
		},
	}

	envelopes, err := domainScanFindingEnvelopes(cfg, "acme.com")
	assert.NilError(t, err)
	assert.Check(t, cmp.Len(envelopes, 2))

	assert.Check(t, cmp.Equal(entityops.FieldValue(envelopes[0].Payload, "source"), domainScanFindingSource))
	assert.Check(t, cmp.Equal(entityops.FieldValue(envelopes[0].Payload, "display_name"), "Missing DMARC (acme.com)"))
	assert.Check(t, strings.Contains(string(envelopes[0].Payload), `"asset_ids":["app.acme.com"]`))

	assert.Check(t, cmp.Equal(entityops.FieldValue(envelopes[1].Payload, "display_name"), "Weak TLS (acme.com)"))
}
