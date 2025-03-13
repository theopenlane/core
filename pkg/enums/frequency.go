package enums

import (
	"fmt"
	"io"
	"strings"
)

type Frequency string

var (
	FrequencyYearly     Frequency = "YEARLY"
	FrequencyQuarterly  Frequency = "QUARTERLY"
	FrequencyBiAnnually Frequency = "BIANNUALLY"
	FrequencyMonthly    Frequency = "MONTHLY"
)

// Values returns a slice of strings that represents all the possible values of the Frequency enum.
// Possible default values are "YEARLY", "QUARTERLY", "BIANNUALLY", and "MONTHLY"
func (Frequency) Values() (kinds []string) {
	for _, s := range []Frequency{FrequencyYearly, FrequencyQuarterly, FrequencyBiAnnually, FrequencyMonthly} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the frequency as a string
func (r Frequency) String() string {
	return string(r)
}

// ToFrequency returns the frequency enum based on string input
func ToFrequency(r string) *Frequency {
	switch r := strings.ToUpper(r); r {
	case FrequencyYearly.String():
		return &FrequencyYearly
	case FrequencyQuarterly.String():
		return &FrequencyQuarterly
	case FrequencyBiAnnually.String():
		return &FrequencyBiAnnually
	case FrequencyMonthly.String():
		return &FrequencyMonthly
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Frequency) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Frequency) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Frequency, got: %T", v) //nolint:err113
	}

	*r = Frequency(str)

	return nil
}
