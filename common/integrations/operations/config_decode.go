package operations

import (
	"bytes"
	"encoding/json"
)

// Decode decodes a config map into a new instance of T
func Decode[T any](config map[string]any) (T, error) {
	var result T
	if err := DecodeConfig(config, &result); err != nil {
		return result, err
	}
	return result, nil
}

// DecodeConfig decodes a config map into a target struct, respecting defaults on the target
func DecodeConfig(config map[string]any, target any) error {
	if target == nil {
		return ErrDecodeConfigTargetNil
	}
	if len(config) == 0 {
		return nil
	}

	payload, err := json.Marshal(config)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	return nil
}
