package server

import (
	"reflect"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

// SchemaRegistry manages OpenAPI schemas dynamically
type SchemaRegistry struct {
	mu      sync.RWMutex
	refs    map[string]*openapi3.SchemaRef
	spec    *openapi3.T
	genOpts []openapi3gen.Option
}

// NewSchemaRegistry creates a new schema registry
func NewSchemaRegistry(spec *openapi3.T, opts ...openapi3gen.Option) *SchemaRegistry {
	return &SchemaRegistry{
		refs:    make(map[string]*openapi3.SchemaRef),
		spec:    spec,
		genOpts: opts,
	}
}

// GetOrRegister gets an existing schema reference or reflects the value into the spec's component
// schemas and returns a reference to it
func (r *SchemaRegistry) GetOrRegister(v any) (*openapi3.SchemaRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	typeName := t.Name()
	if typeName == "" {
		// Handle anonymous types
		typeName = t.String()
	}

	// Check if already registered
	if ref, exists := r.refs[typeName]; exists {
		return ref, nil
	}

	// Generate schema using a generator
	generator := openapi3gen.NewGenerator(r.genOpts...)

	schemaRef, err := generator.NewSchemaRefForValue(v, r.spec.Components.Schemas)
	if err != nil {
		return nil, err
	}

	// Add to components if not inline
	if r.spec.Components.Schemas == nil {
		r.spec.Components.Schemas = make(openapi3.Schemas)
	}

	r.spec.Components.Schemas[typeName] = schemaRef
	ref := &openapi3.SchemaRef{
		Ref: "#/components/schemas/" + typeName,
	}

	r.refs[typeName] = ref

	return ref, nil
}
