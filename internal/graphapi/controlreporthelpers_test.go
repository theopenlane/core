package graphapi

import (
	"testing"

	"entgo.io/contrib/entgql"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/model"
	"gotest.tools/v3/assert"
)

func TestWorstEvidenceStatus(t *testing.T) {
	rejected := enums.EvidenceStatusRejected
	approved := enums.EvidenceStatusAuditorApproved
	submitted := enums.EvidenceStatusSubmitted

	tests := []struct {
		name      string
		evidences []*generated.Evidence
		expected  *enums.EvidenceStatus
	}{
		{
			name:      "empty slice returns nil",
			evidences: []*generated.Evidence{},
			expected:  nil,
		},
		{
			name:      "single item returns its status",
			evidences: []*generated.Evidence{{Status: enums.EvidenceStatusSubmitted}},
			expected:  &submitted,
		},
		{
			name: "returns most severe status",
			evidences: []*generated.Evidence{
				{Status: enums.EvidenceStatusAuditorApproved},
				{Status: enums.EvidenceStatusRejected},
				{Status: enums.EvidenceStatusSubmitted},
			},
			expected: &rejected,
		},
		{
			name: "all same status returns that status",
			evidences: []*generated.Evidence{
				{Status: enums.EvidenceStatusAuditorApproved},
				{Status: enums.EvidenceStatusAuditorApproved},
			},
			expected: &approved,
		},
		{
			name: "missing artifact is worse than needs renewal",
			evidences: []*generated.Evidence{
				{Status: enums.EvidenceStatusNeedsRenewal},
				{Status: enums.EvidenceStatusMissingArtifact},
			},
			expected: func() *enums.EvidenceStatus { s := enums.EvidenceStatusMissingArtifact; return &s }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worstEvidenceStatus(tt.evidences)
			if tt.expected == nil {
				assert.Check(t, result == nil)
				return
			}
			assert.Assert(t, result != nil)
			assert.Equal(t, *tt.expected, *result)
		})
	}
}

func TestShouldCheckForControl(t *testing.T) {
	tests := []struct {
		name            string
		control         *model.ControlInfo
		frameworksInOrg []string
		expected        bool
	}{
		{
			name:            "nil framework always included",
			control:         &model.ControlInfo{ReferenceFramework: nil},
			frameworksInOrg: []string{"SOC2"},
			expected:        true,
		},
		{
			name:            "empty string framework always included",
			control:         &model.ControlInfo{ReferenceFramework: strPtr("")},
			frameworksInOrg: []string{"SOC2"},
			expected:        true,
		},
		{
			name:            "framework present in org list",
			control:         &model.ControlInfo{ReferenceFramework: strPtr("SOC2")},
			frameworksInOrg: []string{"SOC2", "ISO27001"},
			expected:        true,
		},
		{
			name:            "framework absent from org list",
			control:         &model.ControlInfo{ReferenceFramework: strPtr("NIST800-53")},
			frameworksInOrg: []string{"SOC2", "ISO27001"},
			expected:        false,
		},
		{
			name:            "non-nil framework with empty org list",
			control:         &model.ControlInfo{ReferenceFramework: strPtr("SOC2")},
			frameworksInOrg: []string{},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldCheckForControl(tt.control, tt.frameworksInOrg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupControlReportsByCategory(t *testing.T) {
	c1 := &model.ControlReport{ID: "1", RefCode: "CC1.1", Category: strPtr("Access")}
	c2 := &model.ControlReport{ID: "2", RefCode: "CC1.2", Category: strPtr("Access")}
	c3 := &model.ControlReport{ID: "3", RefCode: "CC2.1", Category: strPtr("Availability")}
	c4 := &model.ControlReport{ID: "4", RefCode: "CC3.1", Category: nil}

	tests := []struct {
		name             string
		controls         []*model.ControlReport
		wantCategories   []string
		wantControlCount []int
	}{
		{
			name:             "empty input",
			controls:         []*model.ControlReport{},
			wantCategories:   []string{},
			wantControlCount: []int{},
		},
		{
			name:             "nil category treated as empty string, sorts before named categories",
			controls:         []*model.ControlReport{c3, c4},
			wantCategories:   []string{"", "Availability"},
			wantControlCount: []int{1, 1},
		},
		{
			name:             "sorted alphabetically",
			controls:         []*model.ControlReport{c3, c1, c2},
			wantCategories:   []string{"Access", "Availability"},
			wantControlCount: []int{2, 1},
		},
		{
			name:             "all nil categories collapsed into one group",
			controls:         []*model.ControlReport{c4, c4},
			wantCategories:   []string{""},
			wantControlCount: []int{2},
		},
		{
			name:             "totalCount matches controls length",
			controls:         []*model.ControlReport{c1, c2, c3},
			wantCategories:   []string{"Access", "Availability"},
			wantControlCount: []int{2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := groupControlReportsByCategory(tt.controls)
			assert.Equal(t, len(tt.wantCategories), len(result))
			for i, cat := range result {
				assert.Equal(t, tt.wantCategories[i], cat.Category)
				assert.Equal(t, tt.wantControlCount[i], cat.TotalCount)
				assert.Equal(t, tt.wantControlCount[i], len(cat.Controls))
			}
		})
	}
}

func TestControlEdgeToControlInfo(t *testing.T) {
	fw := "SOC2"
	ctrl := &generated.Control{
		ID:                 "ctrl-1",
		RefCode:            "CC1.1",
		ReferenceFramework: &fw,
	}

	result := controlEdgeToControlInfo(ctrl)

	assert.Equal(t, "ctrl-1", result.ID)
	assert.Equal(t, "CC1.1", result.RefCode)
	assert.Equal(t, &fw, result.ReferenceFramework)
	assert.Equal(t, false, result.IsSubcontrol)
}

func TestSubcontrolEdgeToControlInfo(t *testing.T) {
	fw := "ISO27001"
	sc := &generated.Subcontrol{
		ID:                 "sc-1",
		RefCode:            "A.5.1.1",
		ReferenceFramework: &fw,
	}

	result := subcontrolEdgeToControlInfo(sc)

	assert.Equal(t, "sc-1", result.ID)
	assert.Equal(t, "A.5.1.1", result.RefCode)
	assert.Equal(t, &fw, result.ReferenceFramework)
	assert.Equal(t, true, result.IsSubcontrol)
}

func TestConvertSubcontrolToControlReportEdge(t *testing.T) {
	tests := []struct {
		name     string
		controls []*generated.Subcontrol
		wantLen  int
	}{
		{
			name:     "empty input returns empty slice",
			controls: []*generated.Subcontrol{},
			wantLen:  0,
		},
		{
			name: "maps id and ref code, initializes slices",
			controls: []*generated.Subcontrol{
				{ID: "sc-1", RefCode: "SC-1.1"},
				{ID: "sc-2", RefCode: "SC-1.2"},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSubcontrolToControlReportEdge(tt.controls)
			assert.Equal(t, tt.wantLen, len(result))
			for i, r := range result {
				assert.Equal(t, tt.controls[i].ID, r.ID)
				assert.Equal(t, tt.controls[i].RefCode, r.RefCode)
				assert.Assert(t, r.RelatedControls != nil)
				assert.Assert(t, r.EvidenceStatus != nil)
				assert.Assert(t, r.LinkedPolicies != nil)
			}
		})
	}
}

func TestConvertControlListToControlReports(t *testing.T) {
	fw := "SOC2"

	tests := []struct {
		name     string
		controls []*generated.Control
		wantLen  int
	}{
		{
			name:     "empty input returns empty slice",
			controls: []*generated.Control{},
			wantLen:  0,
		},
		{
			name: "maps fields and initializes enrichment slices",
			controls: []*generated.Control{
				{ID: "ctrl-1", RefCode: "CC1.1", ReferenceFramework: &fw},
				{ID: "ctrl-2", RefCode: "CC2.1"},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertControlListToControlReports(tt.controls)
			assert.Equal(t, tt.wantLen, len(result))
			for i, r := range result {
				assert.Equal(t, tt.controls[i].ID, r.ID)
				assert.Equal(t, tt.controls[i].RefCode, r.RefCode)
				assert.Equal(t, tt.controls[i].ReferenceFramework, r.ReferenceFramework)
				assert.Assert(t, r.RelatedControls != nil)
				assert.Assert(t, r.EvidenceStatus != nil)
				assert.Assert(t, r.LinkedPolicies != nil)
			}
		})
	}
}

func TestConvertControlToControlReportEdge(t *testing.T) {
	fw := "SOC2"

	conn := &generated.ControlConnection{
		TotalCount: 2,
		Edges: []*generated.ControlEdge{
			{Node: &generated.Control{ID: "ctrl-1", RefCode: "CC1.1", ReferenceFramework: &fw}},
			{Node: &generated.Control{ID: "ctrl-2", RefCode: "CC2.1"}},
		},
	}

	result := convertControlToControlReportEdge(conn)

	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, 2, len(result.Edges))
	assert.Equal(t, "ctrl-1", result.Edges[0].Node.ID)
	assert.Equal(t, "CC1.1", result.Edges[0].Node.RefCode)
	assert.Equal(t, "ctrl-2", result.Edges[1].Node.ID)
	assert.Assert(t, result.Edges[0].Node.RelatedControls != nil)
}

func TestConvertReportOrderToControlOrderBy(t *testing.T) {
	tests := []struct {
		name      string
		orderBy   []*model.ControlReportOrder
		wantLen   int
		wantField *generated.ControlOrderField
		wantDir   entgql.OrderDirection
	}{
		{
			name:      "nil returns default created_at desc",
			orderBy:   nil,
			wantLen:   1,
			wantField: generated.ControlOrderFieldCreatedAt,
			wantDir:   entgql.OrderDirectionDesc,
		},
		{
			name:      "created_at ascending",
			orderBy:   []*model.ControlReportOrder{{Field: model.ControlReportOrderFieldCreatedAt, Direction: entgql.OrderDirectionAsc}},
			wantLen:   1,
			wantField: generated.ControlOrderFieldCreatedAt,
			wantDir:   entgql.OrderDirectionAsc,
		},
		{
			name:      "updated_at descending",
			orderBy:   []*model.ControlReportOrder{{Field: model.ControlReportOrderFieldUpdatedAt, Direction: entgql.OrderDirectionDesc}},
			wantLen:   1,
			wantField: generated.ControlOrderFieldUpdatedAt,
			wantDir:   entgql.OrderDirectionDesc,
		},
		{
			name:      "refCode ascending",
			orderBy:   []*model.ControlReportOrder{{Field: model.ControlReportOrderFieldRefCode, Direction: entgql.OrderDirectionAsc}},
			wantLen:   1,
			wantField: generated.ControlOrderFieldRefCode,
			wantDir:   entgql.OrderDirectionAsc,
		},
		{
			name:      "title",
			orderBy:   []*model.ControlReportOrder{{Field: model.ControlReportOrderFieldTitle, Direction: entgql.OrderDirectionAsc}},
			wantLen:   1,
			wantField: generated.ControlOrderFieldTitle,
			wantDir:   entgql.OrderDirectionAsc,
		},
		{
			name:      "referenceFramework",
			orderBy:   []*model.ControlReportOrder{{Field: model.ControlReportOrderFieldReferenceFramework, Direction: entgql.OrderDirectionAsc}},
			wantLen:   1,
			wantField: generated.ControlOrderFieldReferenceFramework,
			wantDir:   entgql.OrderDirectionAsc,
		},
		{
			name: "multiple fields",
			orderBy: []*model.ControlReportOrder{
				{Field: model.ControlReportOrderFieldRefCode, Direction: entgql.OrderDirectionAsc},
				{Field: model.ControlReportOrderFieldTitle, Direction: entgql.OrderDirectionDesc},
			},
			wantLen:   2,
			wantField: generated.ControlOrderFieldRefCode,
			wantDir:   entgql.OrderDirectionAsc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertReportOrderToControlOrderBy(tt.orderBy)
			assert.Equal(t, tt.wantLen, len(result))
			assert.Equal(t, tt.wantField, result[0].Field)
			assert.Equal(t, tt.wantDir, result[0].Direction)
		})
	}
}
