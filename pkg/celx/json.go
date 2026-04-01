package celx

import (
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

// BoolResult extracts a boolean value from a CEL ref.Val.
// Returns ErrNilOutput if val is nil and ErrTypeMismatch if val is not a boolean type.
func BoolResult(val ref.Val) (bool, error) {
	if val == nil {
		return false, ErrNilOutput
	}

	if val.Type() != types.BoolType {
		return false, ErrTypeMismatch
	}

	result, ok := val.Value().(bool)
	if !ok {
		result = val.Equal(types.True) == types.True
	}

	return result, nil
}

// ToJSON converts a CEL value to a native JSON-compatible value
func ToJSON(val ref.Val) (any, error) {
	if val == nil {
		return nil, nil
	}

	native, err := val.ConvertToNative(types.JSONValueType)
	if err != nil {
		return nil, err
	}

	if jsonVal, ok := native.(*structpb.Value); ok {
		return jsonVal.AsInterface(), nil
	}

	return native, nil
}

// ToJSONMap converts a CEL value to a JSON object map
func ToJSONMap(val ref.Val) (map[string]any, error) {
	jsonVal, err := ToJSON(val)
	if err != nil {
		return nil, err
	}

	if jsonVal == nil {
		return map[string]any{}, nil
	}

	obj, ok := jsonVal.(map[string]any)
	if !ok {
		return nil, ErrJSONMapExpected
	}

	return obj, nil
}
