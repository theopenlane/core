package providerkit

import (
	"testing"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

func TestCelMapExpr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entries []entityops.MappingEntry
		want    string
	}{
		{
			name:    "empty entries returns empty object",
			entries: nil,
			want:    "{}",
		},
		{
			name: "single entry",
			entries: []entityops.MappingEntry{
				{Key: "severity", Expr: "payload.severity"},
			},
			want: "{\n  \"severity\": dyn(payload.severity)\n}",
		},
		{
			name: "multiple entries separated by commas",
			entries: []entityops.MappingEntry{
				{Key: "name", Expr: "resource"},
				{Key: "level", Expr: "variant"},
			},
			want: "{\n  \"name\": dyn(resource),\n  \"level\": dyn(variant)\n}",
		},
		{
			name: "key with special characters is quoted",
			entries: []entityops.MappingEntry{
				{Key: "field.name", Expr: "payload.x"},
			},
			want: "{\n  \"field.name\": dyn(payload.x)\n}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := CelMapExpr(tc.entries...)
			if got != tc.want {
				t.Fatalf("CelMapExpr() = %q, want %q", got, tc.want)
			}
		})
	}
}
