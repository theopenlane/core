package models

import (
	"encoding/json"
	"io"

	"github.com/rs/zerolog/log"
)

func marshalGQL[T any](w io.Writer, a T) {
	byteData, err := json.Marshal(a)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling json object")
	}

	_, err = w.Write(byteData)
	if err != nil {
		log.Fatal().Err(err).Msg("error writing json object")
	}
}

func unmarshalGQL[T any](v any, a T) error {
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
