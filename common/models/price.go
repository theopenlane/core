package models

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
)

// Price is a custom type for pricing data
type Price struct {
	// Amount is the dollar amount of the price (e.g 100)
	Amount float64 `json:"amount"`
	// Interval is the interval of the price (e.g monthly, yearly)
	Interval string `json:"interval"`
	// Currency is the currency of the price that is being charged (e.g USD)
	Currency string `json:"currency"`
}

// String returns a string representation of the price
func (p Price) String() string {
	if p.Amount == 0 {
		return "Free"
	}

	return fmt.Sprintf("%v(%s)/%s", p.Amount, p.Currency, p.Interval)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (p Price) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(p)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling json object")
	}

	_, err = w.Write(byteData)
	if err != nil {
		log.Fatal().Err(err).Msg("error writing json object")
	}
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (p *Price) UnmarshalGQL(v interface{}) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteData, &p)
	if err != nil {
		return err
	}

	return err
}
