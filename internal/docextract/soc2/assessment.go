package soc2

import (
	"slices"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/docextract/schema/custom"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

// RiskFloor is the least severe risk rating a report with the given opinion can carry
func RiskFloor(opinion docextract.Opinion) enums.VendorRiskRating {
	switch opinion {
	case OpinionAdverse, OpinionDisclaimer:
		return enums.VendorRiskRatingHigh
	case OpinionQualified:
		return enums.VendorRiskRatingMedium
	default:
		return enums.VendorRiskRatingLow
	}
}

// ApplyRiskFloor raises a risk rating to the floor the opinion sets and never lowers it; a rating
// outside the scale is treated as below the floor
func ApplyRiskFloor(rating enums.VendorRiskRating, opinion docextract.Opinion) enums.VendorRiskRating {
	floor := RiskFloor(opinion)

	if slices.Index(custom.RiskRatings, rating) < slices.Index(custom.RiskRatings, floor) {
		return floor
	}

	return rating
}

// opinionLabels are the plain-language names of each opinion, for text written for people
var opinionLabels = map[docextract.Opinion]string{
	OpinionUnqualified: "Unqualified (clean)",
	OpinionQualified:   "Qualified",
	OpinionAdverse:     "Adverse",
	OpinionDisclaimer:  "Disclaimer of opinion",
}

// opinionPhrases are each opinion as the object of a sentence, such as issued an unqualified (clean) opinion
var opinionPhrases = map[docextract.Opinion]string{
	OpinionUnqualified: "an unqualified (clean) opinion",
	OpinionQualified:   "a qualified opinion",
	OpinionAdverse:     "an adverse opinion",
	OpinionDisclaimer:  "a disclaimer of opinion",
}

// OpinionPhrase is the opinion written as the object of a sentence, empty when it could not be read
func OpinionPhrase(opinion docextract.Opinion) string {
	return opinionPhrases[opinion]
}

// OpinionLabel is the plain-language name of the opinion, "Not found" when it could not be read
func OpinionLabel(opinion docextract.Opinion) string {
	if label, ok := opinionLabels[opinion]; ok {
		return label
	}

	return "Not found"
}
