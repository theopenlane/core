package handlers

import (
	"encoding/json"
	"reflect"
)

// normalizeExampleValue converts strongly-typed example objects into a generic form
// that kin-openapi can validate (maps, slices, primitives). Structs are marshaled
// to JSON and unmarshaled back into map[string]any / []any representations
func normalizeExampleValue(value any) any {
	if value == nil {
		return nil
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}

		return normalizeExampleValue(rv.Elem().Interface())
	}

	switch value.(type) {
	case map[string]any, []any,
		string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		json.Number:
		return value
	}

	data, err := json.Marshal(value)
	if err != nil {
		return value
	}

	var generic any

	if err := json.Unmarshal(data, &generic); err != nil {
		return value
	}

	return generic
}
