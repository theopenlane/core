package openapi

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/rs/zerolog/log"
)

var (
	// ErrInvalidType indicates an unexpected type was provided during enum unmarshalling
	ErrInvalidType = errors.New("unexpected type for enum unmarshal")
)

// marshalGQLJSON marshals the given type into JSON and writes it to the given writer
func marshalGQLJSON[T any](w io.Writer, a T) {
	byteData, err := json.Marshal(a)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling json object")
	}

	_, err = w.Write(byteData)
	if err != nil {
		log.Fatal().Err(err).Msg("error writing json object")
	}
}

// unmarshalGQLJSON unmarshals a JSON object into the given type
func unmarshalGQLJSON[T any](v any, a T) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteData, &a)
	if err != nil {
		return err
	}

	return nil
}
