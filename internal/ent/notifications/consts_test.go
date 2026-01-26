package notifications

import (
	"testing"
)

func TestGetURLPathForObject(t *testing.T) {
	base := "/base"
	objectID := "123"

	tests := []struct {
		name       string
		objectType string
		want       string
	}{
		{
			name:       "InternalPolicy",
			objectType: "InternalPolicy",
			want:       base + "/policies/123/view",
		},
		{
			name:       "Procedure",
			objectType: "Procedure",
			want:       base + "/procedures/123/view",
		},
		{
			name:       "Risk",
			objectType: "Risk",
			want:       base + "/risks/123",
		},
		{
			name:       "Task",
			objectType: "Task",
			want:       base + "/tasks?id=123",
		},
		{
			name:       "Control",
			objectType: "Control",
			want:       base + "/controls/123",
		},
		{
			name:       "Evidence",
			objectType: "Evidence",
			want:       base + "/evidence?id=123",
		},
		{
			name:       "TrustCenterNDARequest",
			objectType: "TrustCenterNDARequest",
			want:       base + "/trust-center/NDAs",
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
