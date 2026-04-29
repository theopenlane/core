package celx

import (
	"reflect"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

var jsonValueType = reflect.TypeOf(&structpb.Value{})

// indexByBinding implements the indexBy(list, key) CEL function, returning a map keyed by the
// named string field of each list element — elements missing the key or with a non-string value
// are silently skipped
func indexByBinding(lhs, rhs ref.Val) ref.Val {
	key, ok := rhs.(celtypes.String)
	if !ok {
		return celtypes.NewErr("indexBy: second argument must be a string")
	}

	iterable, ok := lhs.(traits.Iterable)
	if !ok {
		return celtypes.NewErr("indexBy: first argument must be a list")
	}

	result := map[string]any{}
	it := iterable.Iterator()

	for it.HasNext() == celtypes.True {
		item := it.Next()

		indexer, ok := item.(traits.Indexer)
		if !ok {
			continue
		}

		idVal := indexer.Get(key)
		if celtypes.IsError(idVal) {
			continue
		}

		idStr, ok := idVal.(celtypes.String)
		if !ok {
			continue
		}

		native, err := item.ConvertToNative(jsonValueType)
		if err != nil {
			continue
		}

		sv, ok := native.(*structpb.Value)
		if !ok {
			continue
		}

		result[string(idStr)] = sv.AsInterface()
	}

	return celtypes.DefaultTypeAdapter.NativeToValue(result)
}

// indexByFunc is the CEL function declaration for indexBy
var indexByFunc = cel.Function("indexBy",
	cel.Overload("indexBy_list_string",
		[]*cel.Type{cel.ListType(cel.DynType), cel.StringType},
		cel.MapType(cel.StringType, cel.DynType),
		cel.BinaryBinding(indexByBinding),
	),
)
