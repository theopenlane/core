package audiences

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/common/enums"
)

func TestValidateFilters(t *testing.T) {
	tt := []struct {
		name         string
		audienceType enums.AudienceType
		filters      map[string]any
		hasError     bool
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
			hasError: true,
		},
		{
			name:         " employee audience",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema":     "identity_holder",
				"expression": "target.identity_holder_type == 'EMPLOYEE' && target.is_active == true",
			},
		},
		{
			name:         " contact audience",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema":     "contact",
				"expression": "target.status == 'ACTIVE' && target.email != ''",
			},
		},
		{
			name:         " multiple selector audience",
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
			name:         " audience without selector",
			audienceType: enums.AudienceTypeDynamic,
			filters:      map[string]any{},
			hasError:     true,
		},
		{
			name:         " audience with unsupported schema",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema": "asset",
			},
			hasError: true,
		},
		{
			name:         " user audience",
			audienceType: enums.AudienceTypeDynamic,
			filters:      map[string]any{"schema": "user"},
		},
		{
			name:         " group audience",
			audienceType: enums.AudienceTypeDynamic,
			filters:      map[string]any{"schema": "group", "expression": "target.name == 'All Members'"},
		},
		{
			name:         " subscriber audience",
			audienceType: enums.AudienceTypeDynamic,
			filters:      map[string]any{"schema": "subscriber"},
		},
		{
			name:         " audience does not compile expression at save time",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema":     "identity_holder",
				"expression": "target.missing_field == true",
			},
		},
		{
			name:         " audience with key match",
			audienceType: enums.AudienceTypeDynamic,
			filters: map[string]any{
				"schema": "contact",
				"key_match": map[string]any{
					"target_field": "email",
					"source_field": "email",
				},
			},
			hasError: true,
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilters(tt.audienceType, tt.filters)

			assert.Equal(t, err != nil, tt.hasError)
		})
	}
}
