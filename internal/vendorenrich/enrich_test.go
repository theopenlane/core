package vendorenrich

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestBuildCandidates(t *testing.T) {
	tests := []struct {
		name    string
		vendors []map[string]any
		want    []candidate
	}{
		{
			name: "name and url both present",
			vendors: []map[string]any{
				{"name": "Stripe", "url": "https://stripe.com"},
			},
			want: []candidate{
				{vendor: map[string]any{"name": "Stripe", "url": "https://stripe.com"}, name: "Stripe", domain: "stripe.com"},
			},
		},
		{
			name: "name only, no url",
			vendors: []map[string]any{
				{"name": "Stripe"},
			},
			want: []candidate{
				{vendor: map[string]any{"name": "Stripe"}, name: "Stripe", domain: ""},
			},
		},
		{
			name: "url only, no name",
			vendors: []map[string]any{
				{"url": "https://stripe.com"},
			},
			want: []candidate{
				{vendor: map[string]any{"url": "https://stripe.com"}, name: "", domain: "stripe.com"},
			},
		},
		{
			name: "name whitespace is trimmed",
			vendors: []map[string]any{
				{"name": "  Stripe  "},
			},
			want: []candidate{
				{vendor: map[string]any{"name": "  Stripe  "}, name: "Stripe", domain: ""},
			},
		},
		{
			name: "neither name nor url is dropped",
			vendors: []map[string]any{
				{"categories": []string{"CDN"}},
			},
			want: []candidate{},
		},
		{
			name:    "no vendors",
			vendors: nil,
			want:    []candidate{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCandidates(tt.vendors)

			assert.Assert(t, is.Len(got, len(tt.want)))

			for i := range tt.want {
				assert.Check(t, is.Equal(tt.want[i].name, got[i].name))
				assert.Check(t, is.Equal(tt.want[i].domain, got[i].domain))
				assert.Check(t, is.DeepEqual(tt.want[i].vendor, got[i].vendor))
			}
		})
	}
}

func TestMatchReference(t *testing.T) {
	byDomain := reference{entityID: "by-domain", name: "domain-match", domains: []string{"stripe.com"}}
	byName := reference{entityID: "by-name", name: "Stripe"}
	byDisplayName := reference{entityID: "by-display-name", displayName: "Stripe"}
	byAlias := reference{entityID: "by-alias", name: "other", aliases: []string{"Stripe"}}

	tests := []struct {
		name       string
		candidate  candidate
		references []reference
		want       *reference
	}{
		{
			name:       "domain match wins over name/display name/alias matches",
			candidate:  candidate{name: "Stripe", domain: "stripe.com"},
			references: []reference{byName, byDisplayName, byAlias, byDomain},
			want:       &byDomain,
		},
		{
			name:       "name match wins over display name/alias matches when no domain matches",
			candidate:  candidate{name: "Stripe", domain: "unmatched.example"},
			references: []reference{byDisplayName, byAlias, byName},
			want:       &byName,
		},
		{
			name:       "display name match wins over alias match when no domain/name matches",
			candidate:  candidate{name: "Stripe"},
			references: []reference{byAlias, byDisplayName},
			want:       &byDisplayName,
		},
		{
			name:       "alias match used as last resort",
			candidate:  candidate{name: "Stripe"},
			references: []reference{byAlias},
			want:       &byAlias,
		},
		{
			name:       "no match returns nil",
			candidate:  candidate{name: "Unknown Co", domain: "unknown.example"},
			references: []reference{byName, byDisplayName, byAlias, byDomain},
			want:       nil,
		},
		{
			name:       "candidate with no name or domain returns nil",
			candidate:  candidate{},
			references: []reference{byName, byDisplayName, byAlias, byDomain},
			want:       nil,
		},
		{
			name:       "matching is case-insensitive",
			candidate:  candidate{name: "stripe"},
			references: []reference{byName},
			want:       &byName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchReference(tt.candidate, tt.references)

			if tt.want == nil {
				assert.Check(t, got == nil)
				return
			}

			assert.Assert(t, got != nil)
			assert.Check(t, is.Equal(tt.want.entityID, got.entityID))
		})
	}
}

func TestApplyReference(t *testing.T) {
	tests := []struct {
		name   string
		vendor map[string]any
		ref    reference
		want   map[string]any
	}{
		{
			name:   "display name preferred over name",
			vendor: map[string]any{"name": "stripecdn"},
			ref:    reference{entityID: "1", name: "stripe", displayName: "Stripe"},
			want:   map[string]any{"name": "Stripe", "entity_id": "1"},
		},
		{
			name:   "falls back to name when display name is empty",
			vendor: map[string]any{"name": "stripecdn"},
			ref:    reference{entityID: "1", name: "Stripe"},
			want:   map[string]any{"name": "Stripe", "entity_id": "1"},
		},
		{
			name:   "leaves name untouched when reference has neither",
			vendor: map[string]any{"name": "stripecdn"},
			ref:    reference{entityID: "1"},
			want:   map[string]any{"name": "stripecdn", "entity_id": "1"},
		},
		{
			name:   "description, domains, and logo fields overwrite scan data when present",
			vendor: map[string]any{"name": "stripecdn", "description": "guessed from wappalyzer"},
			ref: reference{
				entityID:      "1",
				name:          "Stripe",
				description:   "vetted description",
				domains:       []string{"stripe.com"},
				logoRemoteURL: "https://stripe.com/logo.png",
				logoFileID:    "file-1",
			},
			want: map[string]any{
				"name":            "Stripe",
				"entity_id":       "1",
				"description":     "vetted description",
				"domains":         []string{"stripe.com"},
				"logo_remote_url": "https://stripe.com/logo.png",
				"logo_file_id":    "file-1",
			},
		},
		{
			name:   "empty reference fields are not applied",
			vendor: map[string]any{"name": "stripecdn", "description": "kept as-is"},
			ref:    reference{entityID: "1", name: "Stripe"},
			want:   map[string]any{"name": "Stripe", "entity_id": "1", "description": "kept as-is"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyReference(tt.vendor, &tt.ref)

			assert.Check(t, is.DeepEqual(tt.want, tt.vendor))
		})
	}
}
