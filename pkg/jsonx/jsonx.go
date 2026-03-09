package jsonx

import (
	"bytes"
	"encoding/json"

	"github.com/wundergraph/astjson"
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

// ToRawMap converts an arbitrary value into a JSON object map of raw values.
func ToRawMap(value any) (map[string]json.RawMessage, error) {
	if value == nil {
		return map[string]json.RawMessage{}, nil
	}

	var out map[string]json.RawMessage
	err := RoundTrip(value, &out)
	if err == nil {
		if out == nil {
			return map[string]json.RawMessage{}, nil
		}

		return out, nil
	}

	var generic any
	if parseErr := RoundTrip(value, &generic); parseErr != nil {
		return nil, err
	}
	if generic == nil {
		return map[string]json.RawMessage{}, nil
	}
	if _, ok := generic.(map[string]any); !ok {
		return nil, ErrObjectExpected
	}

	return nil, err
}

// CloneRawMessage copies a raw JSON document to avoid accidental aliasing.
func CloneRawMessage(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}

	return append(json.RawMessage(nil), raw...)
}

// IsEmptyRawMessage reports whether a raw JSON message is empty or null.
func IsEmptyRawMessage(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

// UnmarshalIfPresent unmarshals raw JSON when it is non-empty.
func UnmarshalIfPresent(raw json.RawMessage, output any) error {
	if IsEmptyRawMessage(raw) {
		return nil
	}

	return json.Unmarshal(raw, output)
}

// DecodeAnyOrNil decodes raw JSON to an untyped value or returns nil on failure/empty input.
func DecodeAnyOrNil(raw json.RawMessage) any {
	if IsEmptyRawMessage(raw) {
		return nil
	}

	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}

	return out
}

// DeepMerge deep-merges patch into base and reports whether the document changed.
// Both arguments must be JSON objects. Returns the merged document and a boolean
// indicating whether the result differs from base.
func DeepMerge(base, patch json.RawMessage) (json.RawMessage, bool, error) {
	if len(patch) == 0 {
		return base, false, nil
	}

	b, err := astjson.ParseBytes(patch)
	if err != nil {
		return nil, false, err
	}

	var a *astjson.Value
	if len(base) > 0 {
		if a, err = astjson.ParseBytes(base); err != nil {
			return nil, false, err
		}
	}

	merged, _, err := astjson.MergeValues(nil, a, b)
	if err != nil {
		return nil, false, err
	}

	out := json.RawMessage(merged.MarshalTo(nil))
	if bytes.Equal(base, out) {
		return base, false, nil
	}

	return out, true, nil
}

// MergeObjectMap shallow-merges a raw JSON object with the supplied top-level patch map.
func MergeObjectMap(base json.RawMessage, patch map[string]json.RawMessage) (json.RawMessage, bool, error) {
	if len(patch) == 0 {
		return CloneRawMessage(base), false, nil
	}

	baseMap, err := ToRawMap(base)
	if err != nil {
		return nil, false, err
	}

	changed := false
	for key, value := range patch {
		if !bytes.Equal(baseMap[key], value) {
			changed = true
		}

		baseMap[key] = CloneRawMessage(value)
	}

	if !changed {
		return CloneRawMessage(base), false, nil
	}

	out, err := json.Marshal(baseMap)
	if err != nil {
		return nil, false, err
	}
	if bytes.Equal(base, out) {
		return CloneRawMessage(base), false, nil
	}

	return out, true, nil
}

// SetObjectKey sets or replaces one top-level key in a raw JSON object.
func SetObjectKey(base json.RawMessage, key string, value any) (json.RawMessage, bool, error) {
	if key == "" {
		return nil, false, ErrKeyRequired
	}

	raw, err := ToRawMessage(value)
	if err != nil {
		return nil, false, err
	}
	if raw == nil {
		raw = json.RawMessage(`null`)
	}

	return MergeObjectMap(base, map[string]json.RawMessage{key: raw})
}

// ApplyOverlay applies a JSON overlay to an existing typed value.
func ApplyOverlay[T any](base T, overlay any) (T, error) {
	if overlay == nil {
		return base, nil
	}

	out := base
	if err := RoundTrip(overlay, &out); err != nil {
		var zero T
		return zero, err
	}

	return out, nil
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
	if IsEmptyRawMessage(raw) {
		return nil, nil
	}

	return raw, nil
}
