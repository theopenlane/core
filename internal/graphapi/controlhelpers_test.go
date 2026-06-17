package graphapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
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

func TestIsSameControlInfo(t *testing.T) {
	soc2 := "SOC2"
	iso := "ISO27001"

	tests := []struct {
		name          string
		refCode       string
		framework     *string
		mappedControl *model.ControlInfo
		expected      bool
	}{
		{
			name:          "different ref codes",
			refCode:       "CC1.1",
			framework:     strPtr("SOC2"),
			mappedControl: &model.ControlInfo{RefCode: "CC2.1", ReferenceFramework: &soc2},
			expected:      false,
		},
		{
			name:          "same ref code both nil framework",
			refCode:       "CC1.1",
			framework:     nil,
			mappedControl: &model.ControlInfo{RefCode: "CC1.1", ReferenceFramework: nil},
			expected:      true,
		},
		{
			name:          "same ref code same non-nil framework",
			refCode:       "CC1.1",
			framework:     &soc2,
			mappedControl: &model.ControlInfo{RefCode: "CC1.1", ReferenceFramework: &soc2},
			expected:      true,
		},
		{
			name:          "same ref code different frameworks",
			refCode:       "CC1.1",
			framework:     &soc2,
			mappedControl: &model.ControlInfo{RefCode: "CC1.1", ReferenceFramework: &iso},
			expected:      false,
		},
		{
			name:          "same ref code one nil one non-nil framework",
			refCode:       "CC1.1",
			framework:     nil,
			mappedControl: &model.ControlInfo{RefCode: "CC1.1", ReferenceFramework: &soc2},
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameControlInfo(tt.refCode, tt.framework, tt.mappedControl)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetControlWherePredicate(t *testing.T) {
	tests := []struct {
		name    string
		where   *generated.ControlWhereInput
		wantNil bool
		wantErr bool
	}{
		{
			name:    "nil input returns nil predicate",
			where:   nil,
			wantNil: true,
		},
		{
			name:    "empty input returns error",
			where:   &generated.ControlWhereInput{},
			wantErr: true,
		},
		{
			name:    "input with filter returns non-nil predicate",
			where:   &generated.ControlWhereInput{RefCode: strPtr("CC1.1")},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getControlWherePredicate(tt.where)

			if tt.wantErr {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err)
			if tt.wantNil {
				assert.Check(t, result == nil)
			} else {
				assert.Check(t, result != nil)
			}
		})
	}
}

func TestConstructWherePredicatesFromStandardRefCodes(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string][]string
		wantLen int
	}{
		{
			name:    "empty map returns only systemOwned predicate",
			input:   map[string][]string{},
			wantLen: 1,
		},
		{
			name:    "single standard returns one match plus systemOwned",
			input:   map[string][]string{"SOC2": {"CC1.1", "CC1.2"}},
			wantLen: 2,
		},
		{
			name:    "multiple standards collapsed to OR clause plus systemOwned",
			input:   map[string][]string{"SOC2": {"CC1.1"}, "ISO27001": {"A.5.1.1"}},
			wantLen: 2,
		},
		{
			name:    "CUSTOM framework uses nil reference framework predicate",
			input:   map[string][]string{customFramework: {"MY-1"}},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constructWherePredicatesFromStandardRefCodes[predicate.Control](context.Background(), tt.input)
			assert.Equal(t, tt.wantLen, len(result))
		})
	}
}
