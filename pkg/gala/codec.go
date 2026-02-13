package gala

import (
	"context"
	"encoding/json"
	"fmt"
)

// Codec encodes and decodes a topic payload type.
type Codec[T any] interface {
	// Encode serializes the typed payload.
	Encode(context.Context, T) ([]byte, error)
	// Decode deserializes payload bytes into the typed payload.
	Decode(context.Context, []byte) (T, error)
}

// JSONCodec is the default JSON implementation of Codec.
type JSONCodec[T any] struct{}

// Encode serializes payload data using JSON.
func (JSONCodec[T]) Encode(_ context.Context, payload T) ([]byte, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPayloadEncodeFailed, err)
	}

	return encoded, nil
}

// Decode deserializes payload data using JSON.
func (JSONCodec[T]) Decode(_ context.Context, payload []byte) (T, error) {
	var decoded T

	if len(payload) == 0 {
		return decoded, ErrEnvelopePayloadRequired
	}

	if err := json.Unmarshal(payload, &decoded); err != nil {
		return decoded, fmt.Errorf("%w: %w", ErrPayloadDecodeFailed, err)
	}

	return decoded, nil
}
