package customtypes

import (
	"encoding/json"
	"io"

	"github.com/rs/zerolog/log"
)

// Address is a custom type for Address
type Address struct {
	// Line1 is the first line of the address
	Line1 string `json:"line1"`
	// Line2 is the second line of the address
	Line2 string `json:"line2"`
	// City is the city of the address
	City string `json:"city"`
	// State is the state of the address
	State string `json:"state"`
	// PostalCode is the postal code of the address
	PostalCode string `json:"postalCode"`
	// Country is the country of the address
	Country string `json:"country"`
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (a Address) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(a)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling json object")
	}

	_, err = w.Write(byteData)
	if err != nil {
		log.Fatal().Err(err).Msg("error writing json object")
	}
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (a *Address) UnmarshalGQL(v interface{}) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteData, &a)
	if err != nil {
		return err
	}

	return err
}
