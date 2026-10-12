package soc2

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

func TestApplyRiskFloorRaisesBelowFloor(t *testing.T) {
	assert.Check(t, ApplyRiskFloor("low", OpinionQualified) == "medium")
	assert.Check(t, ApplyRiskFloor("medium", OpinionAdverse) == "high")
	assert.Check(t, ApplyRiskFloor("low", OpinionDisclaimer) == "high")
}

func TestApplyRiskFloorKeepsRatingsAtOrAboveFloor(t *testing.T) {
	assert.Check(t, ApplyRiskFloor("low", OpinionUnqualified) == "low")
	assert.Check(t, ApplyRiskFloor("critical", OpinionQualified) == "critical")
	assert.Check(t, ApplyRiskFloor("high", OpinionAdverse) == "high")
	assert.Check(t, ApplyRiskFloor("medium", docextract.OpinionUnknown) == "medium")
}

func TestApplyRiskFloorReplacesUnknownRating(t *testing.T) {
	assert.Check(t, ApplyRiskFloor("", OpinionUnqualified) == "low")
	assert.Check(t, ApplyRiskFloor("severe", OpinionQualified) == "medium")
}

func TestOpinionLabel(t *testing.T) {
	assert.Check(t, OpinionLabel(OpinionUnqualified) == "Unqualified (clean)")
	assert.Check(t, OpinionLabel(OpinionDisclaimer) == "Disclaimer of opinion")
	assert.Check(t, OpinionLabel(docextract.OpinionUnknown) == "Not found")
}
