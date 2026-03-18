package jsonx

import (
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
