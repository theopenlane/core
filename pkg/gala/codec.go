package gala

import (
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
)

// Codec encodes and decodes a topic payload type
type Codec[T any] interface {
	// Encode serializes the typed payload
	Encode(T) ([]byte, error)
	// Decode deserializes payload bytes into the typed payload
	Decode([]byte) (T, error)
}

// JSONCodec is the default JSON implementation of Codec
type JSONCodec[T any] struct{}

// Encode serializes payload data using JSON
func (JSONCodec[T]) Encode(payload T) ([]byte, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, ErrPayloadEncodeFailed
	}

	return encoded, nil
}

// Decode deserializes payload data using JSON
func (JSONCodec[T]) Decode(payload []byte) (T, error) {
	var decoded T

	if len(payload) == 0 {
		return decoded, ErrEnvelopePayloadRequired
	}

	if err := jsonx.RoundTrip(payload, &decoded); err != nil {
		return decoded, ErrPayloadDecodeFailed
	}

	return decoded, nil
}
