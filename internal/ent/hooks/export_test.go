package hooks

import (
	"encoding/json"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"gotest.tools/v3/assert"
)

func TestGetFinalFilters(t *testing.T) {
	type testCase struct {
		name                string
		hasOwnerField       bool
		hasSystemOwnedField bool
		exportType          enums.ExportType
		filters             string
		ownerID             string
		expectedErr         string
		expectedFilters     map[string]any
	}

	cases := []testCase{
		{
			name:            "Control export type, should add ownerID and systemOwned automatically",
			exportType:      enums.ExportTypeControl,
			filters:         `{"foo":"bar"}`,
			ownerID:         "owner-123",
			expectedFilters: map[string]any{"foo": "bar", "ownerID": "owner-123", "systemOwned": false},
		},
		{
			name:            "Trust center subprocesor should add no fields",
			exportType:      enums.ExportTypeTrustCenterSubprocessor,
			filters:         `{"foo":"bar"}`,
			ownerID:         "owner-123",
			expectedFilters: map[string]any{"foo": "bar"},
		},
		{
			name:            "Evidence should only add ownerID",
			exportType:      enums.ExportTypeEvidence,
			filters:         `{"foo":"bar"}`,
			ownerID:         "owner-123",
			expectedFilters: map[string]any{"foo": "bar", "ownerID": "owner-123"},
		},
		{
			name:            "No filters provided, should add ownerID and systemOwned",
			exportType:      enums.ExportTypeRemediation,
			filters:         "",
			ownerID:         "owner-123",
			expectedFilters: map[string]any{"ownerID": "owner-123", "systemOwned": false},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mut := &generated.ExportMutation{}
			mut.SetExportType(tc.exportType)
			mut.SetFilters(tc.filters)

			result, err := getFinalFilters(tc.filters, mut, tc.ownerID)
			if tc.expectedErr != "" {
				assert.Error(t, err, tc.expectedErr)
				return

			}

			assert.NilError(t, err)

			// parse result back to map for comparison
			var resultMap map[string]any
			err = json.Unmarshal([]byte(result), &resultMap)
			assert.NilError(t, err)

			for key, expectedValue := range tc.expectedFilters {
				actualValue, exists := resultMap[key]
				assert.Assert(t, exists, "expected key %s to exist in result", key)
				assert.Equal(t, actualValue, expectedValue, "expected value for key %s to be %v, got %v", key, expectedValue, actualValue)
			}

		})
	}
}
