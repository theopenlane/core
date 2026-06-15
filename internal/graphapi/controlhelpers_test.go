package graphapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func strPtr(s string) *string { return &s }

func TestGetStandardRefCodes(t *testing.T) {
	tests := []struct {
		name     string
		data     []string
		expected map[string][]string
		wantErr  bool
	}{
		{
			name:     "Empty input",
			data:     []string{},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "Valid input with single standard",
			data: []string{
				"ISO 27001::A.5.1.1",
				"ISO 27001::A.5.1.2",
			},
			expected: map[string][]string{
				"ISO 27001": {"A.5.1.1", "A.5.1.2"},
			},
			wantErr: false,
		},
		{
			name: "Valid input with multiple standards",
			data: []string{
				"ISO27001::A.5.1.1",
				"NIST800-53::AC-1",
				"ISO27001::A.6.1.1",
				"NIST800-53::AC-2",
			},
			expected: map[string][]string{
				"ISO27001":   {"A.5.1.1", "A.6.1.1"},
				"NIST800-53": {"AC-1", "AC-2"},
			},
			wantErr: false,
		},
		{
			name: "Invalid format - missing colon",
			data: []string{
				"ISO27001A.5.1.1",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Invalid format - too many colons",
			data: []string{
				"ISO27001::A.5.1.1::extra",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Mixed valid and invalid inputs",
			data: []string{
				"ISO27001:A.5.1.1",
				"Invalid-Format",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getStandardRefCodes(tt.data)

			if tt.wantErr {
				assert.ErrorIs(t, err, common.ErrInvalidInput)
				assert.Check(t, result == nil)

				return
			}
			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(result, tt.expected))
		})
	}
}

func TestNormalizeFramework(t *testing.T) {
	tests := []struct {
		name      string
		framework *string
		expected  string
	}{
		{
			name:      "nil framework returns CUSTOM",
			framework: nil,
			expected:  customFramework,
		},
		{
			name:      "non-nil framework returns its value",
			framework: strPtr("ISO27001"),
			expected:  "ISO27001",
		},
		{
			name:      "empty string framework returns empty string",
			framework: strPtr(""),
			expected:  "",
		},
		{
			name:      "SOC2 framework",
			framework: strPtr("SOC2"),
			expected:  "SOC2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeFramework(tt.framework)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFrameworkName(t *testing.T) {
	tests := []struct {
		name     string
		control  *generated.Control
		expected string
	}{
		{
			name:     "nil reference framework returns CUSTOM",
			control:  &generated.Control{ReferenceFramework: nil},
			expected: customFramework,
		},
		{
			name:     "non-nil reference framework returns its value",
			control:  &generated.Control{ReferenceFramework: strPtr("NIST800-53")},
			expected: "NIST800-53",
		},
		{
			name:     "SOC2 reference framework",
			control:  &generated.Control{ReferenceFramework: strPtr("SOC2")},
			expected: "SOC2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFrameworkName(tt.control)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateMapControlKey(t *testing.T) {
	tests := []struct {
		name      string
		refCode   string
		framework *string
		expected  string
	}{
		{
			name:      "nil framework uses CUSTOM",
			refCode:   "CC1.1",
			framework: nil,
			expected:  fmt.Sprintf("CC1.1::%s", customFramework),
		},
		{
			name:      "non-nil framework uses framework value",
			refCode:   "CC1.1",
			framework: strPtr("SOC2"),
			expected:  "CC1.1::SOC2",
		},
		{
			name:      "empty ref code",
			refCode:   "",
			framework: strPtr("ISO27001"),
			expected:  "::ISO27001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateMapControlKey(tt.refCode, tt.framework)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSameControl(t *testing.T) {
	tests := []struct {
		name       string
		refCode    string
		framework  *string
		mappedCtrl *generated.Control
		expected   bool
	}{
		{
			name:       "different ref codes returns false",
			refCode:    "A.5.1",
			framework:  nil,
			mappedCtrl: &generated.Control{RefCode: "B.1.1", ReferenceFramework: nil},
			expected:   false,
		},
		{
			name:       "same ref code, both nil frameworks returns true",
			refCode:    "A.5.1",
			framework:  nil,
			mappedCtrl: &generated.Control{RefCode: "A.5.1", ReferenceFramework: nil},
			expected:   true,
		},
		{
			name:       "same ref code, same non-nil framework returns true",
			refCode:    "CC1.1",
			framework:  strPtr("SOC2"),
			mappedCtrl: &generated.Control{RefCode: "CC1.1", ReferenceFramework: strPtr("SOC2")},
			expected:   true,
		},
		{
			name:       "same ref code, different frameworks returns false",
			refCode:    "CC1.1",
			framework:  strPtr("SOC2"),
			mappedCtrl: &generated.Control{RefCode: "CC1.1", ReferenceFramework: strPtr("ISO27001")},
			expected:   false,
		},
		{
			name:       "same ref code, caller nil framework but control has framework returns false",
			refCode:    "A.5.1",
			framework:  nil,
			mappedCtrl: &generated.Control{RefCode: "A.5.1", ReferenceFramework: strPtr("ISO27001")},
			expected:   false,
		},
		{
			name:       "same ref code, caller has framework but control nil returns false",
			refCode:    "A.5.1",
			framework:  strPtr("ISO27001"),
			mappedCtrl: &generated.Control{RefCode: "A.5.1", ReferenceFramework: nil},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameControl(tt.refCode, tt.framework, tt.mappedCtrl)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSameSubcontrol(t *testing.T) {
	tests := []struct {
		name       string
		refCode    string
		framework  *string
		mappedCtrl *generated.Subcontrol
		expected   bool
	}{
		{
			name:       "different ref codes returns false",
			refCode:    "SC-1",
			framework:  nil,
			mappedCtrl: &generated.Subcontrol{RefCode: "SC-2", ReferenceFramework: nil},
			expected:   false,
		},
		{
			name:       "same ref code, both nil frameworks returns true",
			refCode:    "SC-1",
			framework:  nil,
			mappedCtrl: &generated.Subcontrol{RefCode: "SC-1", ReferenceFramework: nil},
			expected:   true,
		},
		{
			name:       "same ref code, same non-nil framework returns true",
			refCode:    "SC-1",
			framework:  strPtr("NIST800-53"),
			mappedCtrl: &generated.Subcontrol{RefCode: "SC-1", ReferenceFramework: strPtr("NIST800-53")},
			expected:   true,
		},
		{
			name:       "same ref code, different frameworks returns false",
			refCode:    "SC-1",
			framework:  strPtr("SOC2"),
			mappedCtrl: &generated.Subcontrol{RefCode: "SC-1", ReferenceFramework: strPtr("ISO27001")},
			expected:   false,
		},
		{
			name:       "same ref code, caller nil but control has framework returns false",
			refCode:    "SC-1",
			framework:  nil,
			mappedCtrl: &generated.Subcontrol{RefCode: "SC-1", ReferenceFramework: strPtr("SOC2")},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameSubcontrol(tt.refCode, tt.framework, tt.mappedCtrl)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMappedControlInfo(t *testing.T) {
	ctx := context.Background()
	framework := strPtr("SOC2")

	tests := []struct {
		name               string
		ctrl               *generated.Control
		refCode            string
		framework          *string
		includeSystemOwned bool
		expectNil          bool
		expectedKey        string
	}{
		{
			name:      "nil control returns nil",
			ctrl:      nil,
			refCode:   "CC1.1",
			framework: framework,
			expectNil: true,
		},
		{
			name:      "same control returns nil",
			ctrl:      &generated.Control{RefCode: "CC1.1", ReferenceFramework: strPtr("SOC2")},
			refCode:   "CC1.1",
			framework: framework,
			expectNil: true,
		},
		{
			name:        "non-system-owned control is returned directly",
			ctrl:        &generated.Control{RefCode: "CC2.1", ReferenceFramework: strPtr("SOC2"), SystemOwned: false},
			refCode:     "CC1.1",
			framework:   framework,
			expectNil:   false,
			expectedKey: "CC2.1::SOC2",
		},
		{
			name:        "non-system-owned control with nil framework uses CUSTOM key",
			ctrl:        &generated.Control{RefCode: "CC2.1", ReferenceFramework: nil, SystemOwned: false},
			refCode:     "CC1.1",
			framework:   framework,
			expectNil:   false,
			expectedKey: fmt.Sprintf("CC2.1::%s", customFramework),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMappedControlInfo(ctx, tt.ctrl, tt.refCode, tt.framework, tt.includeSystemOwned)

			if tt.expectNil {
				assert.Check(t, result == nil)
				return
			}

			assert.Assert(t, result != nil)
			_, ok := result[tt.expectedKey]
			assert.Check(t, ok, "expected key %q not found in result", tt.expectedKey)
			assert.Equal(t, tt.ctrl, result[tt.expectedKey])
		})
	}
}

func TestGetMappedSubcontrolInfo(t *testing.T) {
	ctx := context.Background()
	framework := strPtr("NIST800-53")

	tests := []struct {
		name               string
		ctrl               *generated.Subcontrol
		refCode            string
		framework          *string
		includeSystemOwned bool
		expectNil          bool
		expectedKey        string
	}{
		{
			name:      "nil subcontrol returns nil",
			ctrl:      nil,
			refCode:   "AC-1",
			framework: framework,
			expectNil: true,
		},
		{
			name:      "same subcontrol returns nil",
			ctrl:      &generated.Subcontrol{RefCode: "AC-1", ReferenceFramework: strPtr("NIST800-53")},
			refCode:   "AC-1",
			framework: framework,
			expectNil: true,
		},
		{
			name:        "non-system-owned subcontrol is returned directly",
			ctrl:        &generated.Subcontrol{RefCode: "AC-2", ReferenceFramework: strPtr("NIST800-53"), SystemOwned: false},
			refCode:     "AC-1",
			framework:   framework,
			expectNil:   false,
			expectedKey: "AC-2::NIST800-53",
		},
		{
			name:        "non-system-owned subcontrol with nil framework uses CUSTOM key",
			ctrl:        &generated.Subcontrol{RefCode: "AC-2", ReferenceFramework: nil, SystemOwned: false},
			refCode:     "AC-1",
			framework:   framework,
			expectNil:   false,
			expectedKey: fmt.Sprintf("AC-2::%s", customFramework),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMappedSubcontrolInfo(ctx, tt.ctrl, tt.refCode, tt.framework, tt.includeSystemOwned)

			if tt.expectNil {
				assert.Check(t, result == nil)
				return
			}

			assert.Assert(t, result != nil)
			_, ok := result[tt.expectedKey]
			assert.Check(t, ok, "expected key %q not found in result", tt.expectedKey)
			assert.Equal(t, tt.ctrl, result[tt.expectedKey])
		})
	}
}

func TestPrepMappedControlQuery(t *testing.T) {
	orgID := ulids.New().String()
	authedCtx := auth.WithCaller(context.Background(), &auth.Caller{
		OrganizationIDs: []string{orgID},
	})

	parentID := ulids.New().String()

	tests := []struct {
		name            string
		ctx             context.Context
		refCode         string
		framework       *string
		parentControlID *string
		wantErr         bool
		wantPredicates  int
	}{
		{
			name:            "no auth context returns error",
			ctx:             context.Background(),
			refCode:         "CC1.1",
			framework:       strPtr("SOC2"),
			parentControlID: nil,
			wantErr:         true,
		},
		{
			name:            "control query with nil framework adds nil framework predicate",
			ctx:             authedCtx,
			refCode:         "CC1.1",
			framework:       nil,
			parentControlID: nil,
			wantErr:         false,
			wantPredicates:  1,
		},
		{
			name:            "control query with non-nil framework",
			ctx:             authedCtx,
			refCode:         "CC1.1",
			framework:       strPtr("SOC2"),
			parentControlID: nil,
			wantErr:         false,
			wantPredicates:  1,
		},
		{
			name:            "subcontrol query with parent control ID",
			ctx:             authedCtx,
			refCode:         "SC-1",
			framework:       strPtr("NIST800-53"),
			parentControlID: &parentID,
			wantErr:         false,
			wantPredicates:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := prepMappedControlQuery(tt.ctx, tt.refCode, tt.framework, tt.parentControlID)

			if tt.wantErr {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tt.wantPredicates, len(result))
		})
	}
}
