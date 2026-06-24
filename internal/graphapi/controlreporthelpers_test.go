package graphapi

import (
	"context"
	"testing"

	"entgo.io/contrib/entgql"
	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/model"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
			assert.Check(t, is.Equal(*tt.expected, *result))
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
			control:         &model.ControlInfo{ReferenceFramework: lo.ToPtr("")},
			frameworksInOrg: []string{"SOC2"},
			expected:        true,
		},
		{
			name:            "framework present in org list",
			control:         &model.ControlInfo{ReferenceFramework: lo.ToPtr("SOC2")},
			frameworksInOrg: []string{"SOC2", "ISO27001"},
			expected:        true,
		},
		{
			name:            "framework absent from org list",
			control:         &model.ControlInfo{ReferenceFramework: lo.ToPtr("NIST800-53")},
			frameworksInOrg: []string{"SOC2", "ISO27001"},
			expected:        false,
		},
		{
			name:            "non-nil framework with empty org list",
			control:         &model.ControlInfo{ReferenceFramework: lo.ToPtr("SOC2")},
			frameworksInOrg: []string{},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldCheckForControl(tt.control, tt.frameworksInOrg)
			assert.Check(t, is.Equal(tt.expected, result))
		})
	}
}

func TestGroupControlReportsByCategory(t *testing.T) {
	c1 := &model.ControlReport{ID: "1", RefCode: "CC1.1", Category: lo.ToPtr("Access")}
	c2 := &model.ControlReport{ID: "2", RefCode: "CC1.2", Category: lo.ToPtr("Access")}
	c3 := &model.ControlReport{ID: "3", RefCode: "CC2.1", Category: lo.ToPtr("Availability")}
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
			assert.Check(t, is.Equal(len(tt.wantCategories), len(result)))
			for i, cat := range result {
				assert.Check(t, is.Equal(tt.wantCategories[i], cat.Category))
				assert.Check(t, is.Equal(tt.wantControlCount[i], cat.TotalCount))
				assert.Check(t, is.Equal(tt.wantControlCount[i], len(cat.Controls)))
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

	assert.Check(t, is.Equal("ctrl-1", result.ID))
	assert.Check(t, is.Equal("CC1.1", result.RefCode))
	assert.Check(t, is.Equal(&fw, result.ReferenceFramework))
	assert.Check(t, is.Equal(false, result.IsSubcontrol))
}

func TestSubcontrolEdgeToControlInfo(t *testing.T) {
	fw := "ISO27001"
	sc := &generated.Subcontrol{
		ID:                 "sc-1",
		RefCode:            "A.5.1.1",
		ReferenceFramework: &fw,
	}

	result := subcontrolEdgeToControlInfo(sc)

	assert.Check(t, is.Equal("sc-1", result.ID))
	assert.Check(t, is.Equal("A.5.1.1", result.RefCode))
	assert.Check(t, is.Equal(&fw, result.ReferenceFramework))
	assert.Check(t, is.Equal(true, result.IsSubcontrol))
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
			assert.Check(t, is.Equal(tt.wantLen, len(result)))
			for i, r := range result {
				assert.Check(t, is.Equal(tt.controls[i].ID, r.ID))
				assert.Check(t, is.Equal(tt.controls[i].RefCode, r.RefCode))
				assert.Check(t, r.RelatedControls != nil)
				assert.Check(t, r.EvidenceStatus != nil)
				assert.Check(t, r.LinkedPolicies != nil)
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
			assert.Check(t, is.Equal(tt.wantLen, len(result)))
			for i, r := range result {
				assert.Check(t, is.Equal(tt.controls[i].ID, r.ID))
				assert.Check(t, is.Equal(tt.controls[i].RefCode, r.RefCode))
				assert.Check(t, is.Equal(tt.controls[i].ReferenceFramework, r.ReferenceFramework))
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

	assert.Check(t, is.Equal(2, result.TotalCount))
	assert.Check(t, is.Equal(2, len(result.Edges)))
	assert.Check(t, is.Equal("ctrl-1", result.Edges[0].Node.ID))
	assert.Check(t, is.Equal("CC1.1", result.Edges[0].Node.RefCode))
	assert.Check(t, is.Equal("ctrl-2", result.Edges[1].Node.ID))
	assert.Assert(t, result.Edges[0].Node.RelatedControls != nil)
}

func TestCollectAllEntityIDs(t *testing.T) {
	tests := []struct {
		name              string
		reports           []*model.ControlReport
		wantControlIDs    []string
		wantSubcontrolIDs []string
	}{
		{
			name:              "empty reports",
			reports:           []*model.ControlReport{},
			wantControlIDs:    nil,
			wantSubcontrolIDs: nil,
		},
		{
			name: "control with no subcontrols or related",
			reports: []*model.ControlReport{
				{ID: "ctrl-1"},
			},
			wantControlIDs:    []string{"ctrl-1"},
			wantSubcontrolIDs: nil,
		},
		{
			name: "control with subcontrols",
			reports: []*model.ControlReport{
				{
					ID: "ctrl-1",
					Subcontrols: []*model.ControlReport{
						{ID: "sc-1"},
						{ID: "sc-2"},
					},
				},
			},
			wantControlIDs:    []string{"ctrl-1"},
			wantSubcontrolIDs: []string{"sc-1", "sc-2"},
		},
		{
			name: "control related controls included",
			reports: []*model.ControlReport{
				{
					ID: "ctrl-1",
					RelatedControls: []*model.ControlInfo{
						{ID: "ctrl-2", IsSubcontrol: false},
						{ID: "sc-rel-1", IsSubcontrol: true},
					},
				},
			},
			wantControlIDs:    []string{"ctrl-1", "ctrl-2"},
			wantSubcontrolIDs: []string{"sc-rel-1"},
		},
		{
			name: "duplicate IDs across reports deduplicated",
			reports: []*model.ControlReport{
				{ID: "ctrl-1"},
				{ID: "ctrl-1"},
				{ID: "ctrl-2"},
			},
			wantControlIDs:    []string{"ctrl-1", "ctrl-2"},
			wantSubcontrolIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotControls, gotSubcontrols := collectAllEntityIDs(tt.reports)
			assert.Check(t, is.Equal(len(tt.wantControlIDs), len(gotControls)))
			assert.Check(t, is.Equal(len(tt.wantSubcontrolIDs), len(gotSubcontrols)))
		})
	}
}

func TestComputeEvidenceStatus(t *testing.T) {
	e1 := &generated.Evidence{ID: "e1", Status: enums.EvidenceStatusAuditorApproved}
	e2 := &generated.Evidence{ID: "e2", Status: enums.EvidenceStatusRejected}
	e3 := &generated.Evidence{ID: "e3", Status: enums.EvidenceStatusSubmitted}

	tests := []struct {
		name            string
		evidenceMap     map[string][]*generated.Evidence
		id              string
		relatedControls []*model.ControlInfo
		wantCount       int
		wantInherited   int
		wantWorst       *enums.EvidenceStatus
	}{
		{
			name:        "no evidence",
			evidenceMap: map[string][]*generated.Evidence{},
			id:          "ctrl-1",
			wantCount:   0,
			wantWorst:   nil,
		},
		{
			name:        "evidence directly on entity",
			evidenceMap: map[string][]*generated.Evidence{"ctrl-1": {e1, e3}},
			id:          "ctrl-1",
			wantCount:   2,
			wantWorst:   func() *enums.EvidenceStatus { s := enums.EvidenceStatusSubmitted; return &s }(),
		},
		{
			name:        "evidence from related control included",
			evidenceMap: map[string][]*generated.Evidence{"ctrl-1": {e1}, "ctrl-rel": {e2}},
			id:          "ctrl-1",
			relatedControls: []*model.ControlInfo{
				{ID: "ctrl-rel", IsSubcontrol: false},
			},
			wantCount:     2,
			wantInherited: 1, // e2 from ctrl-rel
			wantWorst:     func() *enums.EvidenceStatus { s := enums.EvidenceStatusRejected; return &s }(),
		},
		{
			name:        "duplicate evidence across entity and related control deduplicated",
			evidenceMap: map[string][]*generated.Evidence{"ctrl-1": {e1, e2}, "ctrl-rel": {e2, e3}},
			id:          "ctrl-1",
			relatedControls: []*model.ControlInfo{
				{ID: "ctrl-rel", IsSubcontrol: false},
			},
			wantCount:     3, // e1, e2, e3 — e2 counted once
			wantInherited: 1, // e3 only; e2 is direct so it wins
			wantWorst:     func() *enums.EvidenceStatus { s := enums.EvidenceStatusRejected; return &s }(),
		},
		{
			// buildEvidenceMap scopes WithControls to controlIDs so out-of-scope controls
			// never get entries in the map; this test asserts that the lookup honours that boundary
			name:        "evidence on out-of-scope related control does not bleed through",
			evidenceMap: map[string][]*generated.Evidence{"ctrl-1": {e1}},
			id:          "ctrl-1",
			relatedControls: []*model.ControlInfo{
				{ID: "ctrl-oos", IsSubcontrol: false}, // ctrl-oos has e2 but is not in the map
			},
			wantCount: 1,
			wantWorst: func() *enums.EvidenceStatus { s := enums.EvidenceStatusAuditorApproved; return &s }(),
		},
		{
			// a control with no direct evidence but related to one that has evidence gets it
			name:        "control with no direct evidence inherits from related control in map",
			evidenceMap: map[string][]*generated.Evidence{"ctrl-rel": {e1, e3}},
			id:          "ctrl-1",
			relatedControls: []*model.ControlInfo{
				{ID: "ctrl-rel", IsSubcontrol: false},
			},
			wantCount:     2,
			wantInherited: 2, // both e1 and e3 come from ctrl-rel
			wantWorst:     func() *enums.EvidenceStatus { s := enums.EvidenceStatusSubmitted; return &s }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeEvidenceStatus(tt.evidenceMap, tt.id, tt.relatedControls)
			assert.Assert(t, result != nil)
			assert.Check(t, is.Equal(tt.wantCount, result.TotalCount))
			assert.Check(t, is.Equal(tt.wantInherited, result.InheritedCount))
			if tt.wantWorst == nil {
				assert.Check(t, result.WorstStatus == nil)
			} else {
				assert.Assert(t, result.WorstStatus != nil)
				assert.Check(t, is.Equal(*tt.wantWorst, *result.WorstStatus))
			}
		})
	}
}

func TestComputeLinkedPolicies(t *testing.T) {
	p1 := &generated.InternalPolicy{ID: "p1", Name: "Policy A", Status: enums.DocumentPublished}
	p2 := &generated.InternalPolicy{ID: "p2", Name: "Policy B", Status: enums.DocumentDraft}

	tests := []struct {
		name              string
		policiesMap       map[string][]*generated.InternalPolicy
		id                string
		relatedControls   []*model.ControlInfo
		wantCount         int
		wantInheritedFrom map[string][]string
	}{
		{
			name:        "no policies",
			policiesMap: map[string][]*generated.InternalPolicy{},
			id:          "ctrl-1",
			wantCount:   0,
		},
		{
			name:        "policies directly on entity",
			policiesMap: map[string][]*generated.InternalPolicy{"ctrl-1": {p1, p2}},
			id:          "ctrl-1",
			wantCount:   2,
		},
		{
			name:        "policies from related control included",
			policiesMap: map[string][]*generated.InternalPolicy{"ctrl-1": {p1}, "ctrl-rel": {p2}},
			id:          "ctrl-1",
			relatedControls: []*model.ControlInfo{
				{ID: "ctrl-rel", IsSubcontrol: false},
			},
			wantCount:         2,
			wantInheritedFrom: map[string][]string{"p2": {"ctrl-rel"}},
		},
		{
			name:        "duplicate policy across entity and related control deduplicated",
			policiesMap: map[string][]*generated.InternalPolicy{"ctrl-1": {p1, p2}, "ctrl-rel": {p2}},
			id:          "ctrl-1",
			relatedControls: []*model.ControlInfo{
				{ID: "ctrl-rel", IsSubcontrol: false},
			},
			wantCount: 2, // p1 and p2, p2 counted once
		},
		{
			// buildPoliciesMap scopes WithControls to controlIDs so out-of-scope controls
			// never get entries in the map; this test asserts that the lookup honours that boundary
			name:        "policy on out-of-scope related control does not bleed through",
			policiesMap: map[string][]*generated.InternalPolicy{"ctrl-1": {p1}},
			id:          "ctrl-1",
			relatedControls: []*model.ControlInfo{
				{ID: "ctrl-oos", IsSubcontrol: false}, // ctrl-oos has p2 but is not in the map
			},
			wantCount: 1, // only p1, not p2
		},
		{
			// a control with no direct policies but related to one that has policies gets them
			name:        "control with no direct policy inherits from related control in map",
			policiesMap: map[string][]*generated.InternalPolicy{"ctrl-rel": {p1, p2}},
			id:          "ctrl-1",
			relatedControls: []*model.ControlInfo{
				{ID: "ctrl-rel", IsSubcontrol: false},
			},
			wantCount:         2,
			wantInheritedFrom: map[string][]string{"p1": {"ctrl-rel"}, "p2": {"ctrl-rel"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeLinkedPolicies(tt.policiesMap, tt.id, tt.relatedControls)
			assert.Assert(t, result != nil)
			assert.Check(t, is.Equal(tt.wantCount, result.TotalCount))
			assert.Check(t, is.Equal(tt.wantCount, len(result.InternalPolicies)))

			for _, ps := range result.InternalPolicies {
				assert.Check(t, is.DeepEqual(tt.wantInheritedFrom[ps.ID], ps.InheritedFromIDs))
			}
		})
	}
}

func TestResolveRawRelated(t *testing.T) {
	fw := "SOC2"
	orgA := &model.ControlInfo{ID: "org-a", RefCode: "CC1.1", ReferenceFramework: &fw}
	orgB := &model.ControlInfo{ID: "org-b", RefCode: "CC1.2", ReferenceFramework: &fw}
	sysA := &model.ControlInfo{ID: "sys-a", RefCode: "CC1.1", ReferenceFramework: &fw}

	keyA := generateMapControlKey(orgA.RefCode, orgA.ReferenceFramework)
	keyB := generateMapControlKey(orgB.RefCode, orgB.ReferenceFramework)

	tests := []struct {
		name        string
		raw         map[string]map[string]*model.ControlInfo
		sysControls map[string]*model.ControlInfo
		orgLookup   map[string]*model.ControlInfo
		wantKeys    []string
		wantIDs     map[string][]string
	}{
		{
			name:        "empty raw returns empty map",
			raw:         map[string]map[string]*model.ControlInfo{},
			sysControls: map[string]*model.ControlInfo{},
			orgLookup:   map[string]*model.ControlInfo{},
			wantKeys:    []string{},
			wantIDs:     map[string][]string{},
		},
		{
			name:        "org-owned entry passes through unchanged",
			raw:         map[string]map[string]*model.ControlInfo{"outer": {keyB: orgB}},
			sysControls: map[string]*model.ControlInfo{},
			orgLookup:   map[string]*model.ControlInfo{},
			wantKeys:    []string{"outer"},
			wantIDs:     map[string][]string{"outer": {orgB.ID}},
		},
		{
			name:        "system-owned entry replaced by org counterpart",
			raw:         map[string]map[string]*model.ControlInfo{"outer": {keyA: sysA}},
			sysControls: map[string]*model.ControlInfo{keyA: sysA},
			orgLookup:   map[string]*model.ControlInfo{keyA: orgA},
			wantKeys:    []string{"outer"},
			wantIDs:     map[string][]string{"outer": {orgA.ID}},
		},
		{
			name:        "system-owned entry with no org counterpart is excluded",
			raw:         map[string]map[string]*model.ControlInfo{"outer": {keyA: sysA}},
			sysControls: map[string]*model.ControlInfo{keyA: sysA},
			orgLookup:   map[string]*model.ControlInfo{},
			wantKeys:    []string{},
			wantIDs:     map[string][]string{},
		},
		{
			name: "outer key with only excluded system entries is omitted from result",
			raw: map[string]map[string]*model.ControlInfo{
				"outer-with-sys": {keyA: sysA},
				"outer-with-org": {keyB: orgB},
			},
			sysControls: map[string]*model.ControlInfo{keyA: sysA},
			orgLookup:   map[string]*model.ControlInfo{},
			wantKeys:    []string{"outer-with-org"},
			wantIDs:     map[string][]string{"outer-with-org": {orgB.ID}},
		},
		{
			name: "mixed system and org in same inner map",
			raw: map[string]map[string]*model.ControlInfo{
				"outer": {keyA: sysA, keyB: orgB},
			},
			sysControls: map[string]*model.ControlInfo{keyA: sysA},
			orgLookup:   map[string]*model.ControlInfo{keyA: orgA},
			wantKeys:    []string{"outer"},
			wantIDs:     map[string][]string{"outer": {orgA.ID, orgB.ID}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveRawRelated(tt.raw, nil, tt.sysControls, tt.orgLookup)
			assert.Check(t, is.Equal(len(tt.wantKeys), len(result)))

			for outerKey, wantIDs := range tt.wantIDs {
				got, ok := result[outerKey]
				assert.Assert(t, ok, "expected key %q in result", outerKey)
				assert.Check(t, is.Equal(len(wantIDs), len(got)))

				gotIDs := map[string]struct{}{}
				for _, info := range got {
					gotIDs[info.ID] = struct{}{}
				}
				for _, id := range wantIDs {
					_, found := gotIDs[id]
					assert.Assert(t, found, "expected ID %q in result[%q]", id, outerKey)
				}
			}
		})
	}
}

func TestRelatedControlMappingReferenceIDs(t *testing.T) {
	fw := "SOC2"
	orgB := &model.ControlInfo{ID: "org-b", RefCode: "CC1.2", ReferenceFramework: &fw}
	keyB := generateMapControlKey(orgB.RefCode, orgB.ReferenceFramework)

	const (
		outerKey = "outer"
		selfKey  = "self"
	)

	type mapping struct {
		id          string
		systemOwned bool
	}

	tests := []struct {
		name       string
		mappings   []mapping
		wantRefIDs []string
	}{
		{
			name:       "system-owned mapping contributes no reference ids",
			mappings:   []mapping{{id: "mc-sys", systemOwned: true}},
			wantRefIDs: nil,
		},
		{
			name:       "single org-owned mapping contributes one reference id",
			mappings:   []mapping{{id: "mc-1", systemOwned: false}},
			wantRefIDs: []string{"mc-1"},
		},
		{
			name: "duplicate mapping ids are deduplicated to two unique reference ids",
			mappings: []mapping{
				{id: "mc-1", systemOwned: false},
				{id: "mc-2", systemOwned: false},
				{id: "mc-1", systemOwned: false},
			},
			wantRefIDs: []string{"mc-1", "mc-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := map[string]map[string]*model.ControlInfo{}
			refIDs := map[string]map[string][]string{}

			for _, m := range tt.mappings {
				indexRelated(raw, refIDs, outerKey, selfKey, m.id, m.systemOwned, []mcPart{{key: keyB, info: orgB}})
			}

			result := resolveRawRelated(raw, refIDs, map[string]*model.ControlInfo{}, map[string]*model.ControlInfo{})

			related := result[outerKey]
			assert.Assert(t, is.Len(related, 1))
			assert.Check(t, is.Equal(orgB.ID, related[0].ID))
			assert.Check(t, is.DeepEqual(tt.wantRefIDs, related[0].MappedControlReferenceIDs))
		})
	}
}

func TestInheritSubcontrolRelatedControls(t *testing.T) {
	type sub struct {
		id         string
		relatedIDs []string
	}

	build := func(directIDs []string, subs []sub) *model.ControlReport {
		r := &model.ControlReport{}
		for _, id := range directIDs {
			r.RelatedControls = append(r.RelatedControls, &model.ControlInfo{ID: id})
		}
		for _, s := range subs {
			scr := &model.ControlReport{ID: s.id}
			for _, rid := range s.relatedIDs {
				scr.RelatedControls = append(scr.RelatedControls, &model.ControlInfo{ID: rid})
			}
			r.Subcontrols = append(r.Subcontrols, scr)
		}
		return r
	}

	t.Run("subcontrol-only related control is inherited", func(t *testing.T) {
		r := build(nil, []sub{{id: "sc1", relatedIDs: []string{"A"}}})
		inheritSubcontrolRelatedControls(r)

		assert.Assert(t, is.Len(r.RelatedControls, 1))
		assert.Check(t, is.Equal("A", r.RelatedControls[0].ID))
		assert.Check(t, is.DeepEqual([]string{"sc1"}, r.RelatedControls[0].InheritedFromSubcontrolIDs))
		// the subcontrol's own related control must never report inheritance
		assert.Check(t, is.Len(r.Subcontrols[0].RelatedControls[0].InheritedFromSubcontrolIDs, 0))
	})

	t.Run("direct control mapping suppresses the inheritance flag", func(t *testing.T) {
		r := build([]string{"A"}, []sub{{id: "sc1", relatedIDs: []string{"A"}}})
		inheritSubcontrolRelatedControls(r)

		assert.Assert(t, is.Len(r.RelatedControls, 1))
		assert.Check(t, is.Equal("A", r.RelatedControls[0].ID))
		assert.Check(t, is.Len(r.RelatedControls[0].InheritedFromSubcontrolIDs, 0))
	})

	t.Run("multiple subcontrols contribute all their ids", func(t *testing.T) {
		r := build(nil, []sub{
			{id: "sc1", relatedIDs: []string{"A"}},
			{id: "sc2", relatedIDs: []string{"A"}},
		})
		inheritSubcontrolRelatedControls(r)

		assert.Assert(t, is.Len(r.RelatedControls, 1))
		assert.Check(t, is.DeepEqual([]string{"sc1", "sc2"}, r.RelatedControls[0].InheritedFromSubcontrolIDs))
	})

	t.Run("shared related-controls backing is not corrupted across reports", func(t *testing.T) {
		shared := make([]*model.ControlInfo, 1, 8)
		shared[0] = &model.ControlInfo{ID: "direct"}

		reportA := &model.ControlReport{
			RelatedControls: shared,
			Subcontrols:     []*model.ControlReport{{ID: "scA", RelatedControls: []*model.ControlInfo{{ID: "TA"}}}},
		}
		reportB := &model.ControlReport{
			RelatedControls: shared,
			Subcontrols:     []*model.ControlReport{{ID: "scB", RelatedControls: []*model.ControlInfo{{ID: "TB"}}}},
		}

		inheritSubcontrolRelatedControls(reportA)
		inheritSubcontrolRelatedControls(reportB)

		a := map[string]*model.ControlInfo{}
		for _, rc := range reportA.RelatedControls {
			a[rc.ID] = rc
		}
		b := map[string]*model.ControlInfo{}
		for _, rc := range reportB.RelatedControls {
			b[rc.ID] = rc
		}

		// each report keeps only its own inherited entry tagged with its own subcontrol id
		assert.Assert(t, a["TA"] != nil)
		assert.Check(t, is.DeepEqual([]string{"scA"}, a["TA"].InheritedFromSubcontrolIDs))
		assert.Check(t, a["TB"] == nil, "report A must not see report B's inherited control")

		assert.Assert(t, b["TB"] != nil)
		assert.Check(t, is.DeepEqual([]string{"scB"}, b["TB"].InheritedFromSubcontrolIDs))
		assert.Check(t, b["TA"] == nil, "report B must not see report A's inherited control")
	})

	t.Run("direct and inherited related controls coexist", func(t *testing.T) {
		r := build([]string{"B"}, []sub{{id: "sc1", relatedIDs: []string{"A", "B"}}})
		inheritSubcontrolRelatedControls(r)

		assert.Assert(t, is.Len(r.RelatedControls, 2))

		byID := map[string]*model.ControlInfo{}
		for _, rc := range r.RelatedControls {
			byID[rc.ID] = rc
		}
		assert.Check(t, is.Len(byID["B"].InheritedFromSubcontrolIDs, 0))
		assert.Check(t, is.DeepEqual([]string{"sc1"}, byID["A"].InheritedFromSubcontrolIDs))
	})
}

func TestProcessMappedControlResultsNeverSetsInheritedField(t *testing.T) {
	ctx := context.Background()
	fw := "SOC2"

	self := &generated.Control{ID: "self", RefCode: "CC1.1", ReferenceFramework: &fw}
	relatedCtrl := &generated.Control{ID: "ctrl-a", RefCode: "A1.1", ReferenceFramework: &fw}
	relatedSub := &generated.Subcontrol{ID: "sub-a", RefCode: "A1.2", ReferenceFramework: &fw, ControlID: "parent-x"}

	mc := &generated.MappedControl{
		ID:          "mc-1",
		SystemOwned: false,
		Edges: generated.MappedControlEdges{
			FromControls:  []*generated.Control{self},
			ToControls:    []*generated.Control{relatedCtrl},
			ToSubcontrols: []*generated.Subcontrol{relatedSub},
		},
	}

	result, err := processMappedControlResults(ctx, []*generated.MappedControl{mc}, "self", "CC1.1", &fw, []string{fw})
	assert.NilError(t, err)
	assert.Assert(t, is.Len(result, 2))

	for _, rc := range result {
		assert.Check(t, is.Len(rc.InheritedFromSubcontrolIDs, 0), "control resolver must never set inheritedFromSubcontrolIDs")
	}
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
			assert.Check(t, is.Equal(tt.wantLen, len(result)))
			assert.Check(t, is.Equal(tt.wantField, result[0].Field))
			assert.Check(t, is.Equal(tt.wantDir, result[0].Direction))
		})
	}
}
