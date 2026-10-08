package soc2

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

func TestScopeFilterDropsOutOfScopeEntries(t *testing.T) {
	sections := docextract.Sections{
		"findings": json.RawMessage(`[
			{"description": "real", "refCodes": ["ACF-01", "CC1.1"]},
			{"description": "fabricated", "refCodes": ["ACF-99"]},
			{"description": "no codes"}
		]`),
		"domains": json.RawMessage(`[{"name": "Security"}]`),
	}

	filtered, dropped := newScopeFilter([]string{"ACF-01", "ACF-02"}).Filter(sections)

	assert.Check(t, dropped == 3)

	var findings []map[string]any

	assert.NilError(t, json.Unmarshal(filtered["findings"], &findings))
	assert.Check(t, len(findings) == 1)
	assert.Check(t, findings[0]["description"] == "real")

	var domains []map[string]any

	assert.NilError(t, json.Unmarshal(filtered["domains"], &domains))
	assert.Check(t, len(domains) == 0)
}

func TestScopeFilterKeepsInScopeSections(t *testing.T) {
	raw := json.RawMessage(`[{"refCodes": ["ACF-01"]}, {"refCodes": ["CC6.1", "ACF-02"]}]`)

	filtered, dropped := newScopeFilter([]string{"ACF-01", "ACF-02"}).Filter(docextract.Sections{"reviews": raw})

	assert.Check(t, dropped == 0)
	assert.Check(t, string(filtered["reviews"]) == string(raw))
}

func TestPlanScopedRequestsFilter(t *testing.T) {
	plan, err := NewKind(testPrompts()).Plan(docextract.Request{Section: "controls", Scope: []string{"ACF-01"}})

	assert.NilError(t, err)
	assert.Check(t, plan.Filter != nil)

	unscoped, err := NewKind(testPrompts()).Plan(docextract.Request{Section: "controls"})

	assert.NilError(t, err)
	assert.Check(t, unscoped.Filter == nil)
}

func TestControlRefCodesKeepsReportOrderAndDedupes(t *testing.T) {
	refCodes := ControlRefCodes(json.RawMessage(`[
		{"refCode": "CC6.1.1"},
		{"refCode": "CC1.1.1"},
		{"refCode": "CC6.1.1"},
		{"refCode": ""},
		{"title": "no ref code"}
	]`))

	assert.DeepEqual(t, refCodes, []string{"CC6.1.1", "CC1.1.1"})
}

func TestControlRefCodesOnMalformedSection(t *testing.T) {
	assert.Check(t, ControlRefCodes(json.RawMessage(`{"controls": []}`)) == nil)
}
