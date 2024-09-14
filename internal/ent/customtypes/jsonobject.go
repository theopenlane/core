package customtypes

import (
	"encoding/json"
	"io"

	"github.com/rs/zerolog/log"
)

// JSONObject is a custom type for JSON object templates
type JSONObject map[string]interface{}

// MarshalGQL implement the Marshaler interface for gqlgen
func (j JSONObject) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(j)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling json object")
	}

	_, err = w.Write(byteData)
	if err != nil {
		log.Fatal().Err(err).Msg("error writing json object")
	}
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (j *JSONObject) UnmarshalGQL(v interface{}) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteData, &j)
	if err != nil {
		return err
	}

	return err
}
