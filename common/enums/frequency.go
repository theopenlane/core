package enums

import "io"

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
)

var frequencyValues = []Frequency{
	FrequencyYearly,
	FrequencyQuarterly,
	FrequencyBiAnnually,
	FrequencyMonthly,
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
