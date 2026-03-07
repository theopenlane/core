package operations

import (
	"bytes"
	"encoding/json"
)

// Decode decodes a config document into a new instance of T
func Decode[T any](config json.RawMessage) (T, error) {
	var result T
	if err := DecodeConfig(config, &result); err != nil {
		return result, err
	}

	return result, nil
}

// DecodeConfig decodes a config document into a target struct, respecting defaults on the target
func DecodeConfig(config json.RawMessage, target any) error {
	if target == nil {
		return ErrDecodeConfigTargetNil
	}
	if len(config) == 0 || string(config) == "null" {
		return nil
	}

	decoder := json.NewDecoder(bytes.NewReader(config))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	return nil
}
