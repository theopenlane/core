package graphapi

import (
	"errors"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/graphapi/model"
)

func TestValidateDomainScanReviewRefs(t *testing.T) {
	baseInput := model.ImportDomainScanReviewInput{
		Vendors: []*model.ImportDomainScanReviewVendorInput{
			{Ref: "vendor-1", Name: "Stripe"},
			{Ref: "vendor-2", Name: "Datadog"},
		},
		Assets: []*model.ImportDomainScanReviewAssetInput{
			{Ref: "asset-1", Name: "example.com"},
			{Ref: "asset-2", Name: "api.example.com"},
		},
	}

	tests := []struct {
		name    string
		input   model.ImportDomainScanReviewInput
		wantErr bool
	}{
		{
			name: "valid platform and system refs",
			input: func() model.ImportDomainScanReviewInput {
				in := baseInput
				in.Platforms = []*model.ImportDomainScanReviewPlatformInput{
					{Ref: "platform-1", Name: "Openlane Console", EntityRefs: []string{"vendor-1"}, AssetRefs: []string{"asset-1"}},
					{Ref: "platform-2", Name: "Admin Portal", EntityRefs: []string{"vendor-2"}, AssetRefs: []string{"asset-2"}},
				}
				in.Systems = []*model.ImportDomainScanReviewSystemInput{
					{Name: "Console", EntityRefs: []string{"vendor-1"}, AssetRefs: []string{"asset-1"}, PlatformRefs: []string{"platform-1"}},
					{Name: "Admin", EntityRefs: []string{"vendor-2"}, AssetRefs: []string{"asset-2"}, PlatformRefs: []string{"platform-1", "platform-2"}},
				}
				return in
			}(),
			wantErr: false,
		},
		{
			name: "unknown vendor ref on platform",
			input: func() model.ImportDomainScanReviewInput {
				in := baseInput
				in.Platforms = []*model.ImportDomainScanReviewPlatformInput{
					{Ref: "platform-1", Name: "Openlane", EntityRefs: []string{"does-not-exist"}},
				}
				return in
			}(),
			wantErr: true,
		},
		{
			name: "unknown asset ref on a system",
			input: func() model.ImportDomainScanReviewInput {
				in := baseInput
				in.Systems = []*model.ImportDomainScanReviewSystemInput{{Name: "Console", AssetRefs: []string{"does-not-exist"}}}
				return in
			}(),
			wantErr: true,
		},
		{
			name: "unknown platform ref on a system",
			input: func() model.ImportDomainScanReviewInput {
				in := baseInput
				in.Platforms = []*model.ImportDomainScanReviewPlatformInput{{Ref: "platform-1", Name: "Openlane"}}
				in.Systems = []*model.ImportDomainScanReviewSystemInput{{Name: "Console", PlatformRefs: []string{"does-not-exist"}}}
				return in
			}(),
			wantErr: true,
		},
		{
			name:    "no platform or systems is valid",
			input:   baseInput,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDomainScanReviewRefs(tt.input)

			if tt.wantErr {
				assert.Assert(t, err != nil)
				assert.Assert(t, errors.Is(err, ErrDomainScanReviewInvalidRef))
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestBuildDomainScanImportEnvelope(t *testing.T) {
	input := model.ImportDomainScanReviewInput{
		ScanIDs: []string{"scan-1", "scan-2"},
		Platforms: []*model.ImportDomainScanReviewPlatformInput{
			{
				Ref:         "platform-1",
				Name:        "Openlane",
				Description: lo.ToPtr("Compliance platform"),
				EntityRefs:  []string{"vendor-1"},
				AssetRefs:   []string{"asset-1"},
			},
			{
				Ref:         "platform-2",
				Name:        "Admin Portal",
				Description: lo.ToPtr("Internal admin tooling"),
				EntityRefs:  []string{"vendor-2"},
				AssetRefs:   []string{"asset-2"},
			},
		},
		Systems: []*model.ImportDomainScanReviewSystemInput{
			{
				Name:         "Console",
				Description:  lo.ToPtr("Web console"),
				EntityRefs:   []string{"vendor-1"},
				AssetRefs:    []string{"asset-1"},
				PlatformRefs: []string{"platform-1"},
			},
			{
				Name:         "Admin",
				Description:  lo.ToPtr("Admin backend"),
				EntityRefs:   []string{"vendor-2"},
				AssetRefs:    []string{"asset-2"},
				PlatformRefs: []string{"platform-1", "platform-2"},
			},
		},
		Vendors: []*model.ImportDomainScanReviewVendorInput{
			{Ref: "vendor-1", Name: "Stripe", Domain: lo.ToPtr("stripe.com"), Categories: []string{"payments"}},
			{Ref: "vendor-2", Name: "Datadog", Domain: lo.ToPtr("datadoghq.com"), Categories: []string{"monitoring"}},
		},
		Assets: []*model.ImportDomainScanReviewAssetInput{
			{Ref: "asset-1", Name: "example.com", Identifier: lo.ToPtr("example.com"), Website: lo.ToPtr("https://example.com")},
			{Ref: "asset-2", Name: "api.example.com", Identifier: lo.ToPtr("api.example.com"), Website: lo.ToPtr("https://api.example.com")},
		},
		Findings: []*model.ImportDomainScanReviewFindingInput{
			{Category: lo.ToPtr("security"), Description: lo.ToPtr("missing SPF"), Severity: lo.ToPtr("medium")},
		},
	}

	got := buildDomainScanImportEnvelope("org-1", input)

	assert.Equal(t, got.OrganizationID, "org-1")
	assert.DeepEqual(t, got.ScanIDs, []string{"scan-1", "scan-2"})

	assert.Assert(t, len(got.Vendors) == 2)
	assert.Equal(t, got.Vendors[0].Ref, "vendor-1")
	assert.Equal(t, got.Vendors[0].Domain, "stripe.com")
	assert.Equal(t, got.Vendors[1].Ref, "vendor-2")
	assert.Equal(t, got.Vendors[1].Domain, "datadoghq.com")

	assert.Assert(t, len(got.Assets) == 2)
	assert.Equal(t, got.Assets[0].Identifier, "example.com")
	assert.Equal(t, got.Assets[0].Website, "https://example.com")
	assert.Equal(t, got.Assets[1].Identifier, "api.example.com")
	assert.Equal(t, got.Assets[1].Website, "https://api.example.com")

	assert.Assert(t, len(got.Platforms) == 2)
	assert.Equal(t, got.Platforms[0].Ref, "platform-1")
	assert.Equal(t, got.Platforms[0].Description, "Compliance platform")
	assert.DeepEqual(t, got.Platforms[0].EntityRefs, []string{"vendor-1"})
	assert.DeepEqual(t, got.Platforms[0].AssetRefs, []string{"asset-1"})
	assert.Equal(t, got.Platforms[1].Ref, "platform-2")
	assert.Equal(t, got.Platforms[1].Description, "Internal admin tooling")
	assert.DeepEqual(t, got.Platforms[1].EntityRefs, []string{"vendor-2"})
	assert.DeepEqual(t, got.Platforms[1].AssetRefs, []string{"asset-2"})

	assert.Assert(t, len(got.Systems) == 2)
	assert.Equal(t, got.Systems[0].Description, "Web console")
	assert.DeepEqual(t, got.Systems[0].EntityRefs, []string{"vendor-1"})
	assert.DeepEqual(t, got.Systems[0].AssetRefs, []string{"asset-1"})
	assert.DeepEqual(t, got.Systems[0].PlatformRefs, []string{"platform-1"})
	assert.Equal(t, got.Systems[1].Description, "Admin backend")
	assert.DeepEqual(t, got.Systems[1].EntityRefs, []string{"vendor-2"})
	assert.DeepEqual(t, got.Systems[1].AssetRefs, []string{"asset-2"})
	assert.DeepEqual(t, got.Systems[1].PlatformRefs, []string{"platform-1", "platform-2"})

	assert.Assert(t, len(got.Findings) == 1)
	assert.Equal(t, got.Findings[0].Severity, "medium")
}
