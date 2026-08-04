package operations

import (
	"time"

	"github.com/theopenlane/core/common/enums"
)

// NextCampaignRunAt computes the next run time from the given base time using
// calendar-based frequency and interval arithmetic. All frequencies are
// calendar-relative (month boundaries, not fixed durations) so time.AddDate is
// used rather than time.Add
func NextCampaignRunAt(from time.Time, frequency enums.Frequency, interval int, timezone string) time.Time {
	loc := time.UTC
	if timezone != "" {
		if parsed, err := time.LoadLocation(timezone); err == nil {
			loc = parsed
		}
	}

	base := from.In(loc)

	switch frequency {
	case enums.FrequencyMonthly:
		return base.AddDate(0, interval, 0).In(time.UTC)
	case enums.FrequencyQuarterly:
		return base.AddDate(0, quarterMonths*interval, 0).In(time.UTC)
	case enums.FrequencyBiAnnually:
		return base.AddDate(0, biannualMonths*interval, 0).In(time.UTC)
	case enums.FrequencyYearly:
		return base.AddDate(interval, 0, 0).In(time.UTC)
	default:
		return from
	}
}

const (
	// quarterMonths is the number of months in a quarter
	quarterMonths = 3
	// biannualMonths is the number of months in a half year
	biannualMonths = 6
)
