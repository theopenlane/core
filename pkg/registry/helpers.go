package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"dario.cat/mergo"
	"github.com/barkimedes/go-deepcopy"
	"github.com/danielgtaylor/huma/v2"
	genjs "github.com/invopop/jsonschema"
	loadjs "github.com/santhosh-tekuri/jsonschema/v6"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

var ErrNilValue = errors.New("nil value")

var ErrUnsupportedFormat = errors.New("unsupported format")

type schemaMeta struct {
	genSchema  *genjs.Schema
	loadSchema *loadjs.Schema
	jsonFormat []byte
}

var (
	reflector = &genjs.Reflector{
		AllowAdditionalProperties:  true,
		RequiredFromJSONSchemaTags: true,
		DoNotReference:             true,
	}
	schemaMetasMutex = sync.Mutex{}
	schemaMetas      = map[reflect.Type]*schemaMeta{}
)

// GetSchema returns the json schema of t.
func GetSchema(t reflect.Type) (*genjs.Schema, error) {
	if t == nil {
		return nil, fmt.Errorf("%w: nil type", ErrNilValue)
	}

	sm, err := getSchemaMeta(t)
	if err != nil {
		return nil, err
	}

	return sm.genSchema, nil
}

func getSchemaMeta(t reflect.Type) (*schemaMeta, error) {
	// return early if the we get a nil type
	if t == nil {
		return nil, fmt.Errorf("%w: nil type", ErrNilValue)
	}

	schemaMetasMutex.Lock()
	defer schemaMetasMutex.Unlock()

	sm, exists := schemaMetas[t]
	if exists {
		return sm, nil
	}

	var err error

	sm = &schemaMeta{}
	sm.genSchema = reflector.ReflectFromType(t)

	if _, ok := getFormatFunc(sm.genSchema.Format); !ok {
		return nil, fmt.Errorf("%w: %v got unsupported format: %s", ErrUnsupportedFormat, t, sm.genSchema.Format)
	}

	for _, definition := range sm.genSchema.Definitions {
		if _, ok := getFormatFunc(definition.Format); !ok {
			return nil, fmt.Errorf("%w: %v got unsupported format: %s", ErrUnsupportedFormat, t, definition.Format)
		}
	}

	sm.jsonFormat, err = json.Marshal(sm.genSchema)
	if err != nil {
		return nil, fmt.Errorf("%w: marshal %#v to json failed", err, sm.loadSchema)
	}

	sm.loadSchema, err = loadMemorySchema(sm.jsonFormat)
	if err != nil {
		return nil, fmt.Errorf("%w: new schema from %s failed", err, sm.jsonFormat)
	}

	schemaMetas[t] = sm

	return sm, nil
}

// Validate validates by json schema rules, custom formats and general methods.
func Validate(v interface{}) *ValidateRecorder {
	vr := &ValidateRecorder{}

	if v == nil {
		vr.recordSystem(ErrNilValue)
	}

	jsonBuff, err := json.Marshal(v)
	if err != nil {
		vr.recordSystem(fmt.Errorf("%w: marshal %#v to json failed", err, v))
		return vr
	}

	var rawValue interface{}

	err = json.Unmarshal(jsonBuff, &rawValue)
	if err != nil {
		vr.recordSystem(fmt.Errorf("%w: unmarshal json %s failed", err, jsonBuff))
		return vr
	}

	sm, err := getSchemaMeta(reflect.TypeOf(v))
	if err != nil {
		vr.recordSystem(fmt.Errorf("%w: get schema meta for %T failed", err, v))
		return vr
	}

	err = sm.loadSchema.Validate(rawValue)
	vr.recordJSONSchema(err)

	val := reflect.ValueOf(v)
	traverseGo(&val, nil, vr.record)

	return vr
}

// traverseGo recursively traverses the golang data structure with the rules below:
//
// 1. It traverses fields of the embedded struct.
// 2. It does not traverse unexposed subfields of the struct.
// 3. It passes nil to the argument StructField when it's not a struct field.
// 4. It stops when encountering nil.
func traverseGo(val *reflect.Value, field *reflect.StructField, fn func(*reflect.Value, *reflect.StructField)) {
	t := val.Type()

	switch t.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Slice, reflect.Ptr:
		if val.IsNil() {
			return
		}
	}

	fn(val, field)

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			subfield, subval := t.Field(i), val.Field(i)
			// unexposed
			if subfield.PkgPath != "" {
				continue
			}

			if subfield.Type.Kind() == reflect.Ptr && subval.IsNil() {
				continue
			}

			traverseGo(&subval, &subfield, fn)
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			subval := val.Index(i)
			traverseGo(&subval, nil, fn)
		}
	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			k, v := iter.Key(), iter.Value()
			traverseGo(&k, nil, fn)
			traverseGo(&v, nil, fn)
		}
	case reflect.Ptr:
		child := val.Elem()
		traverseGo(&child, nil, fn)
	}
}

// SchemaProvider defines an interface named
// `SchemaProvider`. This interface specifies a method signature `Schema` that takes a `huma.Registry`
// as input and returns a pointer to a `jsonschema.Schema`.
type SchemaProvider interface {
	Schema(r huma.Registry) *huma.Schema
}

// CustomSchema struct is defined as a type that implements the `SchemaProvider` interface. This
// means that the `CustomSchema` struct has a method named `Schema` that takes a `huma.Registry` as
// input and returns a pointer to a `huma.Schema`. In the implementation of the `Schema` method for
// `CustomSchema`, it returns a `huma.Schema` with a specific `Type` value of "string".
type CustomSchema struct {
	Genschema  *genjs.Schema
	Loadschema *loadjs.Schema
}

func (c CustomSchema) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: "string",
	}
}

var _ huma.SchemaProvider = (*CustomSchema)(nil)

// SchemaTransformer defines an interface in Go that specifies a method signature
// `Transform` which takes a `*genjs.Schema` as input and returns a modified `*genjs.Schema`. This
// interface is intended to be implemented by types that provide transformation logic for JSON schemas.
// The `TransformSchema` function in the provided code snippet iterates over instances of types that
// implement the `SchemaTransformer` interface and applies the `Transform` method of each transformer
// to the input schema sequentially. This allows for flexible and customizable transformation of JSON
// schemas by chaining multiple transformers together.
type SchemaTransformer interface {
	Transform(s *genjs.Schema) *genjs.Schema
}

// TransformSchema function takes a JSON schema represented by a `genjs.Schema` as input, along
// with one or more `SchemaTransformer` implementations. It then iterates over each transformer
// provided and applies the `Transform` method of each transformer to the input schema sequentially.
func TransformSchema(s *genjs.Schema, transformers ...SchemaTransformer) *genjs.Schema {
	for _, transformer := range transformers {
		s = transformer.Transform(s)
	}

	return s
}

// Validate takes a JSON schema and an instance of any type as input and validates the instance against the schema
func ValidateAgainstSchema(schema *genjs.Schema, instance any) error {
	c := loadjs.NewCompiler()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(schema); err != nil {
		return err
	}

	if err := c.AddResource(string(schema.ID), &buf); err != nil {
		return err
	}

	validator, err := c.Compile(string(schema.ID))
	if err != nil {
		return err
	}

	return validator.Validate(instance)
}

// GetSchemaFromType is a function that takes any type `T` as input and generates a JSON schema
// representation of that type using the `jsonschema.Reflect` function. It then marshals this schema
// into an indented JSON string using `json.MarshalIndent` and returns this string along with any error
// that may occur during the process. This function allows you to easily obtain the JSON schema
// representation of a given type for further processing or validation purposes.
func GetSchemaFromType[T any](t T) (string, error) {
	schema := genjs.Reflect(t)

	schemaData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}

	return string(schemaData), nil
}

// RoundTripThrough round trips input through T. It may be used to understand
// how various types affect JSON marshalling or apply go's defaulting to an
// untyped value.
func RoundTripThrough[T any, K any](input K) (K, error) {
	through, err := UnmarshalInto[T](input)
	if err != nil {
		var zero K

		return zero, err
	}

	return UnmarshalInto[K](through)
}

// UnmarshalInto "converts" input into T by marshalling input to JSON and then
// unmarshalling into T.
func UnmarshalInto[T any](input any) (T, error) {
	var output T

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(input); err != nil {
		return output, err
	}

	if err := json.NewDecoder(&buf).Decode(&output); err != nil {
		return output, err
	}

	return output, nil
}

// BuildDefinitionMapFromSchema takes a JSON schema as input and builds a mapping of
// definitions within that schema. It creates a map where the key is the reference to the definition
// (using JSON schema reference syntax) and the value is the actual schema definition itself
func BuildDefinitionMapFromSchema(s *genjs.Schema) genjs.Definitions {
	mapping := map[string]*genjs.Schema{"#" + s.Anchor: s}

	for def, schema := range s.Definitions {
		mapping["#/$defs/"+def] = schema
	}

	return mapping
}

// CreatePropMap is a helper for constructing inline OrderedMaps
func CreatePropMap(init map[string]*genjs.Schema) *orderedmap.OrderedMap[string, *genjs.Schema] {
	props := genjs.NewProperties()
	for key, value := range init {
		props.Set(key, value)
	}

	return props
}

// you could use the below 2 functions doing something like this:
// func main() {
//	generatedSchema := reflector.Reflect(&config.Config{})
//	modifySchema(generatedSchema, cleanUp)
// }
// func cleanUp(s *jsonschema.Schema) {
// 	if len(s.OneOf) > 0 || len(s.AnyOf) > 0 {
// 		s.Ref = ""
// 		s.Type = ""
// 		s.Items = nil
// 		s.PatternProperties = nil
// 	}
// }

func walk(schema *genjs.Schema, visit func(s *genjs.Schema)) {
	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		visit(pair.Value)
	}

	for _, definition := range schema.Definitions {
		for pair := definition.Properties.Oldest(); pair != nil; pair = pair.Next() {
			visit(pair.Value)
		}
	}
}

// ModifySchema is used to apply one or more visitor functions to a JSON schema. It
// takes a JSON schema as input along with one or more visitor functions. The function `walk` is used
// internally to traverse the properties and definitions of the schema. For each property or definition
// encountered during the traversal, the corresponding visitor function is applied to modify that
// specific part of the schema.
func ModifySchema(schema *genjs.Schema, visitors ...func(s *genjs.Schema)) {
	// Apply visitors
	if len(visitors) > 0 {
		for _, visitor := range visitors {
			walk(schema, visitor)
		}
	}
}

// LoadSchema takes an input of type interface{} and attempts
// to load and parse a JSON schema from it. It first creates a deep copy of the input schema using the
// deepcopy package. It then checks the type of the copied schema - if it is a slice of interfaces,
// it merges all the schemas in the slice into a single map. If it is already a map, it directly
// assigns it to the `schema` variable. If the type is not recognized as either a slice or a map, it
// returns `nil` indicating that no schema could be loaded
func LoadSchema(inputSchema interface{}) (*genjs.Schema, error) {
	var schema map[string]interface{}

	schemaRaw, err := deepcopy.Anything(inputSchema)
	if err != nil {
		return nil, err
	}

	switch schemaRaw := schemaRaw.(type) {
	case []interface{}:
		if err = Merge(&schema, schemaRaw); err != nil {
			return nil, err
		}
	case map[string]interface{}:
		schema = schemaRaw
	default:
		// If we can't detect the schema, we don't have one
		return nil, nil
	}

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	rs := &genjs.Schema{}

	err = json.Unmarshal(schemaBytes, rs)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

// Merge takes a pointer to a map of string to interface as the destination
// (`dst`) and a slice of interfaces (`schemas`) as input. It iterates over each schema in the slice
// and merges it into the destination map using the `mergo.Merge` function
func Merge(dst *map[string]interface{}, schemas []interface{}) error {
	for _, schema := range schemas {
		if err := mergo.Merge(dst, schema, mergo.WithAppendSlice); err != nil {
			return err
		}
	}

	return nil
}
