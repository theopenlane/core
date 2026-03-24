package providerkit

import (
	"testing"
)

func TestCelMapExpr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entries []CelMapEntry
		want    string
	}{
		{
			name:    "empty entries returns empty object",
			entries: nil,
			want:    "{}",
		},
		{
			name: "single entry",
			entries: []CelMapEntry{
				{Key: "severity", Expr: "payload.severity"},
			},
			want: "{\n  \"severity\": payload.severity\n}",
		},
		{
			name: "multiple entries separated by commas",
			entries: []CelMapEntry{
				{Key: "name", Expr: "resource"},
				{Key: "level", Expr: "variant"},
			},
			want: "{\n  \"name\": resource,\n  \"level\": variant\n}",
		},
		{
			name: "key with special characters is quoted",
			entries: []CelMapEntry{
				{Key: "field.name", Expr: "payload.x"},
			},
			want: "{\n  \"field.name\": payload.x\n}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := CelMapExpr(tc.entries)
			if got != tc.want {
				t.Fatalf("CelMapExpr() = %q, want %q", got, tc.want)
			}
		})
	}
}
