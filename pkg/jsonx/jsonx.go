package jsonx

import (
	"bytes"
	"encoding/json"
)

// RoundTrip marshals input to JSON and unmarshals it into output
func RoundTrip(input any, output any) error {
	switch typed := input.(type) {
	case []byte:
		if err := json.Unmarshal(typed, output); err == nil {
			return nil
		}
	case json.RawMessage:
		if err := json.Unmarshal(typed, output); err == nil {
			return nil
		}
	}

	bytes, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, output)
}

// ToMap converts an arbitrary value into a JSON object map
func ToMap(value any) (map[string]any, error) {
	var out any
	if err := RoundTrip(value, &out); err != nil {
		return nil, err
	}

	if out == nil {
		return map[string]any{}, nil
	}

	mapped, ok := out.(map[string]any)
	if !ok {
		return nil, ErrObjectExpected
	}

	return mapped, nil
}

// CloneRawMessage copies a raw JSON document to avoid accidental aliasing.
func CloneRawMessage(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}

	return append(json.RawMessage(nil), raw...)
}

// ToRawMessage converts an arbitrary value into a raw JSON document.
func ToRawMessage(value any) (json.RawMessage, error) {
	if value == nil {
		return nil, nil
	}

	var raw json.RawMessage
	if err := RoundTrip(value, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}

	return raw, nil
}
