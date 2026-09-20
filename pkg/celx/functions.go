package celx

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

var jsonValueType = reflect.TypeOf(&structpb.Value{})

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

	str := string(s)
	if strings.Contains(str, "\n") {
		return s
	}

	return celtypes.String(paragraphSeparatorRe.ReplaceAllString(strings.TrimSpace(str), "\n\n"))
}

// paragraphsFunc is the CEL function declaration for paragraphs
var paragraphsFunc = cel.Function("paragraphs",
	cel.Overload("paragraphs_dyn",
		[]*cel.Type{cel.DynType},
		cel.StringType,
		cel.UnaryBinding(paragraphsBinding),
	),
)
