package celx

import (
	"maps"
	"reflect"
	"regexp"
	"strings"

	"cel.dev/cel-go/cel"
	celtypes "cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
	"github.com/theopenlane/utils/jsonx"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

var jsonValueType = reflect.TypeFor[*structpb.Value]()

// paragraphSeparatorRe matches runs of three or more spaces used as a paragraph separator in
// single-line provider descriptions
var paragraphSeparatorRe = regexp.MustCompile(` {3,}`)

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

// paragraphsBinding implements the paragraphs(s) CEL function, normalizing single-line provider
// descriptions that use runs of three or more spaces as paragraph separators into blank-line
// separated paragraphs; strings that already contain a newline are returned unchanged
func paragraphsBinding(val ref.Val) ref.Val {
	var s celtypes.String

	switch v := val.(type) {
	case celtypes.String:
		s = v
	case celtypes.Null:
		return celtypes.String("")
	default:
		return celtypes.NewErr("paragraphs: argument must be a string")
	}

	return celtypes.String(Paragraphs(string(s)))
}

// Paragraphs normalizes a single-line description that uses runs of three or more spaces as paragraph separators into
// blank-line separated paragraphs, a string that already contains a newline is returned unchanged
func Paragraphs(str string) string {
	if strings.Contains(str, "\n") {
		return str
	}

	return paragraphSeparatorRe.ReplaceAllString(strings.TrimSpace(str), "\n\n")
}

// paragraphsFunc is the CEL function declaration for paragraphs
var paragraphsFunc = cel.Function("paragraphs",
	cel.Overload("paragraphs_dyn",
		[]*cel.Type{cel.DynType},
		cel.StringType,
		cel.UnaryBinding(paragraphsBinding),
	),
)

// mergeBinding implements the merge(base, overrides) CEL function, returning base with every key of overrides set on it,
// so a map expression can keep a payload and change a few fields, a null base yields the overrides
func mergeBinding(lhs, rhs ref.Val) ref.Val {
	base, err := ToJSONMap(lhs)
	if err != nil {
		return celtypes.NewErr("merge: first argument must be a map")
	}

	overrides, err := ToJSONMap(rhs)
	if err != nil {
		return celtypes.NewErr("merge: second argument must be a map")
	}

	maps.Copy(base, overrides)

	return celtypes.DefaultTypeAdapter.NativeToValue(base)
}

// compactBinding implements the compact(map) CEL function, returning the map without the keys whose value is null, an
// empty string, an empty list or an empty map, so a map expression can leave out fields it has no value for
func compactBinding(val ref.Val) ref.Val {
	values, err := ToJSONMap(val)
	if err != nil {
		return celtypes.NewErr("compact: argument must be a map")
	}

	for key, value := range values {
		if jsonx.IsEmptyValue(value) {
			delete(values, key)
		}
	}

	return celtypes.DefaultTypeAdapter.NativeToValue(values)
}

// compactFunc is the CEL function declaration for compact
var compactFunc = cel.Function("compact",
	cel.Overload("compact_map",
		[]*cel.Type{cel.MapType(cel.StringType, cel.DynType)},
		cel.MapType(cel.StringType, cel.DynType),
		cel.UnaryBinding(compactBinding),
	),
)

// mergeFunc is the CEL function declaration for merge, the base is dyn so a null payload reaches the binding
var mergeFunc = cel.Function("merge",
	cel.Overload("merge_dyn_map",
		[]*cel.Type{cel.DynType, cel.MapType(cel.StringType, cel.DynType)},
		cel.MapType(cel.StringType, cel.DynType),
		cel.BinaryBinding(mergeBinding),
	),
)
