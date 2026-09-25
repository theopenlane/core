package soc2

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

func testPrompts() Prompts {
	return Prompts{
		Report:       "REPORT LAYOUT",
		Sections:     map[string]string{"controls": "EXTRACT CONTROLS", "reviews": "EXTRACT REVIEWS"},
		Scope:        "SCOPE: {{ .RefCodes }}",
		Continuation: "CONTINUE AFTER {{ .LastControl }} ({{ .LastCriterion }})",
	}
}

func TestKindName(t *testing.T) {
	kind := NewKind(testPrompts())

	assert.Check(t, kind.Name() == KindName)
	assert.Check(t, kind.Prompts().Report == "REPORT LAYOUT")
}

func TestPlanControls(t *testing.T) {
	plan, err := NewKind(testPrompts()).Plan(docextract.Request{Section: "controls"})

	assert.NilError(t, err)
	assert.Check(t, strings.HasPrefix(plan.Prompt, "REPORT LAYOUT\nEXTRACT CONTROLS"))
	assert.Check(t, strings.HasSuffix(plan.Prompt, "Requested Sections: controls"))
	assert.Check(t, plan.ResponseSchema == sectionSchemas[IncludeControls])
	assert.Check(t, plan.Stream == nil)
}

func TestPlanScopedToControls(t *testing.T) {
	plan, err := NewKind(testPrompts()).Plan(docextract.Request{Section: "controls", Scope: []string{"ACF-01", "ACF-02"}})

	assert.NilError(t, err)
	assert.Check(t, strings.Contains(plan.Prompt, "SCOPE: ACF-01, ACF-02"))
}

func TestPlanReviewsStreams(t *testing.T) {
	plan, err := NewKind(testPrompts()).Plan(docextract.Request{Section: "reviews"})

	assert.NilError(t, err)
	assert.Check(t, plan.Stream != nil)
	assert.Check(t, plan.ResponseSchema == sectionSchemas[IncludeReviews])
}

func TestPlanUnknownSection(t *testing.T) {
	_, err := NewKind(testPrompts()).Plan(docextract.Request{Section: "invoices"})

	assert.ErrorIs(t, err, ErrUnknownPart)
}

func TestPlanRequiresReportPrompt(t *testing.T) {
	prompts := testPrompts()
	prompts.Report = ""

	_, err := NewKind(prompts).Plan(docextract.Request{Section: "controls"})

	assert.ErrorIs(t, err, ErrReportPromptRequired)
}

func TestPlanRequiresSectionPrompt(t *testing.T) {
	_, err := NewKind(testPrompts()).Plan(docextract.Request{Section: "assets"})

	assert.ErrorIs(t, err, ErrPromptNotConfigured)
}

func TestPlanRequiresScopePrompt(t *testing.T) {
	prompts := testPrompts()
	prompts.Scope = ""

	_, err := NewKind(prompts).Plan(docextract.Request{Section: "controls", Scope: []string{"ACF-01"}})

	assert.ErrorIs(t, err, ErrScopePromptRequired)
}

func TestConfiguredPartsInRequestOrder(t *testing.T) {
	prompts := Prompts{Sections: map[string]string{"reviews": "r", "controls": "c", "programs": "p", "assets": ""}}

	assert.DeepEqual(t, prompts.ConfiguredParts(), []string{"programs", "controls", "reviews"})
}

func TestUnknownSections(t *testing.T) {
	prompts := Prompts{Sections: map[string]string{"controls": "c", "invoices": "not a section"}}

	assert.DeepEqual(t, prompts.UnknownSections(), []string{"invoices"})
	assert.Check(t, len(testPrompts().UnknownSections()) == 0)
}

func TestPartNamesRoundTrip(t *testing.T) {
	for _, name := range PartNames() {
		part, ok := PartByName(name)

		assert.Check(t, ok, name)
		assert.Check(t, PartName(part) == name)
	}

	_, ok := PartByName("nope")
	assert.Check(t, !ok)
	assert.Check(t, PartName(0) == "")
}
