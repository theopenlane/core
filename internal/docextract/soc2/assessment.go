package soc2

import (
	"slices"

	"github.com/theopenlane/core/v2/internal/docextract/schema"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

// RiskFloor is the least severe risk rating a report with the given opinion can carry
func RiskFloor(opinion docextract.Opinion) string {
	switch opinion {
	case OpinionAdverse, OpinionDisclaimer:
		return "high"
	case OpinionQualified:
		return "medium"
	default:
		return "low"
	}
}

// ApplyRiskFloor raises a risk rating to the floor the opinion sets and never lowers it; a rating
// outside the scale is treated as below the floor
func ApplyRiskFloor(rating string, opinion docextract.Opinion) string {
	floor := RiskFloor(opinion)

	if slices.Index(schema.RiskRatings, rating) < slices.Index(schema.RiskRatings, floor) {
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

// OpinionLabel is the plain-language name of the opinion, "Not found" when it could not be read
func OpinionLabel(opinion docextract.Opinion) string {
	if label, ok := opinionLabels[opinion]; ok {
		return label
	}

	return "Not found"
}
