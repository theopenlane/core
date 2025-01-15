package models

import (
	"encoding/json"
	"io"

	"github.com/rs/zerolog/log"
)

type Actor struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Details any    `json:"details"`
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (a Actor) MarshalGQL(w io.Writer) {
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
func (a *Actor) UnmarshalGQL(v interface{}) error {
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
