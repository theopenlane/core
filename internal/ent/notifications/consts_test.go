package notifications

import (
	"testing"
)

func TestGetURLPathForObject(t *testing.T) {
	base := "https://console.theopenlane.io/"
	baseWithoutSlash := "https://console.theopenlane.io"
	objectID := "123"

	tests := []struct {
		name       string
		base       string
		objectType string
		want       string
	}{
		{
			name:       "InternalPolicy",
			objectType: "InternalPolicy",
			base:       base,
			want:       "https://console.theopenlane.io/policies/123/view",
		},
		{
			name:       "Procedure",
			objectType: "Procedure",
			base:       baseWithoutSlash,
			want:       "https://console.theopenlane.io/procedures/123/view",
		},
		{
			name:       "Risk",
			objectType: "Risk",
			base:       base,
			want:       "https://console.theopenlane.io/risks/123",
		},
		{
			name:       "Task",
			objectType: "Task",
			base:       baseWithoutSlash,
			want:       "https://console.theopenlane.io/tasks?id=123",
		},
		{
			name:       "Control",
			objectType: "Control",
			base:       base,
			want:       "https://console.theopenlane.io/controls/123",
		},
		{
			name:       "Evidence",
			objectType: "Evidence",
			base:       baseWithoutSlash,
			want:       "https://console.theopenlane.io/evidence?id=123",
		},
		{
			name:       "TrustCenterNDARequest",
			objectType: "TrustCenterNDARequest",
			base:       base,
			want:       "https://console.theopenlane.io/trust-center/NDAs",
		},
		{
			name:       "UnknownType",
			objectType: "Unknown",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getURLPathForObject(base, objectID, tt.objectType)
			if got != tt.want {
				t.Errorf("getURLPathForObject() = %q, want %q", got, tt.want)
			}
		})
	}
}
