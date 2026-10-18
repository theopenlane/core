package soc2

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

func TestApplyRiskFloorRaisesBelowFloor(t *testing.T) {
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingLow, OpinionQualified) == enums.VendorRiskRatingMedium)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingMedium, OpinionAdverse) == enums.VendorRiskRatingHigh)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingLow, OpinionDisclaimer) == enums.VendorRiskRatingHigh)
}

func TestApplyRiskFloorKeepsRatingsAtOrAboveFloor(t *testing.T) {
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingLow, OpinionUnqualified) == enums.VendorRiskRatingLow)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingCritical, OpinionQualified) == enums.VendorRiskRatingCritical)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingHigh, OpinionAdverse) == enums.VendorRiskRatingHigh)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingMedium, docextract.OpinionUnknown) == enums.VendorRiskRatingMedium)
}

func TestApplyRiskFloorReplacesUnknownRating(t *testing.T) {
	assert.Check(t, ApplyRiskFloor("", OpinionUnqualified) == enums.VendorRiskRatingLow)
	assert.Check(t, ApplyRiskFloor("SEVERE", OpinionQualified) == enums.VendorRiskRatingMedium)
}

func TestOpinionPhrase(t *testing.T) {
	assert.Check(t, OpinionPhrase(OpinionUnqualified) == "an unqualified (clean) opinion")
	assert.Check(t, OpinionPhrase(OpinionDisclaimer) == "a disclaimer of opinion")
	assert.Check(t, OpinionPhrase(docextract.OpinionUnknown) == "")
}

func TestOpinionLabel(t *testing.T) {
	assert.Check(t, OpinionLabel(OpinionUnqualified) == "Unqualified (clean)")
	assert.Check(t, OpinionLabel(OpinionDisclaimer) == "Disclaimer of opinion")
	assert.Check(t, OpinionLabel(docextract.OpinionUnknown) == "Not found")
}
