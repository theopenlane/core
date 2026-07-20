package runtime

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/operations"
)

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

func TestDomainScanVendorAssetMatches(t *testing.T) {
	assets := []operations.ImportDomainScanReviewAsset{
		{Ref: "a1", Identifier: "cdn.iubenda.com"},
		{Ref: "a2", Identifier: "iubenda.com"},
		{Ref: "a3", Identifier: "theopenlane.io"},
		{Ref: "a4", Identifier: ""},
	}
	assetIDByRef := map[string]string{
		"a1": "asset-1",
		"a2": "asset-2",
		"a3": "asset-3",
	}

	tests := []struct {
		name         string
		vendorDomain string
		want         []string
	}{
		{
			name:         "matches subdomain and exact domain",
			vendorDomain: "iubenda.com",
			want:         []string{"asset-1", "asset-2"},
		},
		{
			name:         "no match for unrelated domain",
			vendorDomain: "theopenlane.io",
			want:         []string{"asset-3"},
		},
		{
			name:         "empty vendor domain matches nothing",
			vendorDomain: "",
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domainScanVendorAssetMatches(tt.vendorDomain, assets, assetIDByRef)

			assert.DeepEqual(t, got, tt.want)
		})
	}
}
