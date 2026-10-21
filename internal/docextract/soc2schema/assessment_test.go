package soc2schema

import (
	"testing"

	"github.com/theopenlane/precis/docextract"
	"github.com/theopenlane/precis/docextract/kinds/soc2"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
)

func TestApplyRiskFloorRaisesBelowFloor(t *testing.T) {
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingLow, soc2.OpinionQualified) == enums.VendorRiskRatingMedium)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingMedium, soc2.OpinionAdverse) == enums.VendorRiskRatingHigh)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingLow, soc2.OpinionDisclaimer) == enums.VendorRiskRatingHigh)
}

func TestApplyRiskFloorKeepsRatingsAtOrAboveFloor(t *testing.T) {
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingLow, soc2.OpinionUnqualified) == enums.VendorRiskRatingLow)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingCritical, soc2.OpinionQualified) == enums.VendorRiskRatingCritical)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingHigh, soc2.OpinionAdverse) == enums.VendorRiskRatingHigh)
	assert.Check(t, ApplyRiskFloor(enums.VendorRiskRatingMedium, docextract.OpinionUnknown) == enums.VendorRiskRatingMedium)
}

func TestApplyRiskFloorReplacesUnknownRating(t *testing.T) {
	assert.Check(t, ApplyRiskFloor("", soc2.OpinionUnqualified) == enums.VendorRiskRatingLow)
	assert.Check(t, ApplyRiskFloor("SEVERE", soc2.OpinionQualified) == enums.VendorRiskRatingMedium)
}
