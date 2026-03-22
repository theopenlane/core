package models

import (
	"encoding/json"
	"io"

	"github.com/rs/zerolog/log"
)

// CredentialSet is an opaque provider-owned persisted credential JSON envelope
type CredentialSet struct {
	// Data is the opaque provider-owned persisted credential JSON
	Data json.RawMessage `json:"data,omitempty"`
}

// MarshalGQL implements the graphql.Marshaler interface for gqlgen scalar serialization
func (c CredentialSet) MarshalGQL(w io.Writer) {
	byteData, err := json.Marshal(c)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshalling credential set")
	}

	if _, err = w.Write(byteData); err != nil {
		log.Fatal().Err(err).Msg("error writing credential set")
	}
}

// UnmarshalGQL implements the graphql.Unmarshaler interface for gqlgen scalar deserialization
func (c *CredentialSet) UnmarshalGQL(v any) error {
	byteData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, c)
}
