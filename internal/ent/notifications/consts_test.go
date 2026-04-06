package notifications

import (
	"testing"
)

func TestGetURLPathForObject(t *testing.T) {
	objectID := "123"

	tests := []struct {
		name       string
		objectType string
		want       string
	}{
		{
			name:       "InternalPolicy",
			objectType: "InternalPolicy",
			want:       "policies/123/view",
		},
		{
			name:       "Procedure",
			objectType: "Procedure",
			want:       "procedures/123/view",
		},
		{
			name:       "Risk",
			objectType: "Risk",
			want:       "risks/123",
		},
		{
			name:       "Task",
			objectType: "Task",
			want:       "tasks?id=123",
		},
		{
			name:       "Control",
			objectType: "Control",
			want:       "controls/123",
		},
		{
			name:       "Evidence",
			objectType: "Evidence",
			want:       "evidence?id=123",
		},
		{
			name:       "TrustCenterNDARequest",
			objectType: "TrustCenterNDARequest",
			want:       "trust-center/NDAs",
		},
		{
			name:       "UnknownType",
			objectType: "Unknown",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getURLPathForObject(objectID, tt.objectType)
			if got != tt.want {
				t.Errorf("getURLPathForObject() = %q, want %q", got, tt.want)
			}
		})
	}
}
