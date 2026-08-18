package enums

import (
	"io"
	"time"
)

// Frequency is a custom type for frequency
type Frequency string

var (
	// FrequencyYearly indicates that the frequency should occur yearly
	FrequencyYearly Frequency = "YEARLY"
	// FrequencyQuarterly indicates that the frequency should occur quarterly
	FrequencyQuarterly Frequency = "QUARTERLY"
	// FrequencyBiAnnually indicates that the frequency should occur bi-annually
	FrequencyBiAnnually Frequency = "BIANNUALLY"
	// FrequencyMonthly indicates that the frequency should occur monthly
	FrequencyMonthly Frequency = "MONTHLY"
	// FrequencyNone indicates that there is no frequency
	FrequencyNone Frequency = "NONE"
	// FrequencyNone indicates that frequency should occur every 2 years
	FrequencyBiennially Frequency = "BIENNIALLY"
	// FrequencyNone indicates that frequency should occur every 3 years
	FrequencyTriennially Frequency = "TRIENNIALLY"
)

var frequencyValues = []Frequency{
	FrequencyYearly,
	FrequencyQuarterly,
	FrequencyBiAnnually,
	FrequencyMonthly,
	FrequencyNone,
	FrequencyBiennially,
	FrequencyTriennially,
}

// Values returns a slice of strings that represents all the possible values of the Frequency enum.
// Possible default values are "YEARLY", "QUARTERLY", "BIANNUALLY", and "MONTHLY"
func (Frequency) Values() []string { return stringValues(frequencyValues) }

// String returns the frequency as a string
func (r Frequency) String() string { return string(r) }

// ToFrequency returns the frequency enum based on string input
func ToFrequency(r string) *Frequency { return parse(r, frequencyValues, nil) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Frequency) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Frequency) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

const (
	// quarterMonths is the number of months in a quarter
	quarterMonths = 3
	// biannualMonths is the number of months in a half year
	biannualMonths = 6
)

// NextOccurrence computes the next occurrence from the given base time using calendar-based
// interval arithmetic in the given timezone (UTC when empty or unparsable). Frequencies are
// calendar-relative (month boundaries, not fixed durations) so time.AddDate is used rather
// than time.Add; frequencies without calendar semantics return the base time unchanged
func (r Frequency) NextOccurrence(from time.Time, interval int, timezone string) time.Time {
	loc := time.UTC
	if timezone != "" {
		if parsed, err := time.LoadLocation(timezone); err == nil {
			loc = parsed
		}
	}

	base := from.In(loc)

	switch r {
	case FrequencyMonthly:
		return base.AddDate(0, interval, 0).In(time.UTC)
	case FrequencyQuarterly:
		return base.AddDate(0, quarterMonths*interval, 0).In(time.UTC)
	case FrequencyBiAnnually:
		return base.AddDate(0, biannualMonths*interval, 0).In(time.UTC)
	case FrequencyYearly:
		return base.AddDate(interval, 0, 0).In(time.UTC)
	default:
		return from
	}
}
