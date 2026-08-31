package audiences

import (
	"testing"

	"github.com/theopenlane/core/common/enums"
)

func TestValidateAudienceFilters(t *testing.T) {
	tests := []struct {
		name         string
		audienceType enums.AudienceType
		filters      map[string]any
		wantErr      bool
	}{
		{
			name:         "manual audience without filters",
			audienceType: enums.AudienceTypeManual,
		},
		{
			name:         "manual audience with filters",
			audienceType: enums.AudienceTypeManual,
			filters: map[string]any{
				"schema": "contact",
			},
			wantErr: true,
		},
		{
			name:         "dynamic employee audience",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema":     "identity_holder",
				"expression": "target.identity_holder_type == 'EMPLOYEE' && target.is_active == true",
			},
		},
		{
			name:         "dynamic contact audience",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema":     "contact",
				"expression": "target.status == 'ACTIVE' && target.email != ''",
			},
		},
		{
			name:         "dynamic multiple selector audience",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"selectors": []map[string]any{
					{
						"schema":     "contact",
						"expression": "target.email != ''",
					},
					{
						"schema":     "identity_holder",
						"expression": "target.email != ''",
					},
				},
			},
		},
		{
			name:         "dynamic audience without selector",
			audienceType: enums.AudienceTypeDynamic,
			filters:      map[string]any{},
			wantErr:      true,
		},
		{
			name:         "dynamic audience with unsupported schema",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema": "user",
			},
			wantErr: true,
		},
		{
			name:         "dynamic audience does not compile expression at save time",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema":     "identity_holder",
				"expression": "target.missing_field == true",
			},
		},
		{
			name:         "dynamic audience with key match",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema": "contact",
				"key_match": map[string]any{
					"target_field": "email",
					"source_field": "email",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAudienceFilters(tt.audienceType, tt.filters)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateAudienceFilters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
