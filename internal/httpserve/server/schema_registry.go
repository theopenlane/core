package server

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

// SchemaRegistry manages OpenAPI schemas dynamically
type SchemaRegistry struct {
	mu      sync.RWMutex
	schemas map[string]any
	refs    map[string]*openapi3.SchemaRef
	spec    *openapi3.T
	genOpts []openapi3gen.Option
}

// NewSchemaRegistry creates a new schema registry
func NewSchemaRegistry(spec *openapi3.T, opts ...openapi3gen.Option) *SchemaRegistry {
	return &SchemaRegistry{
		schemas: make(map[string]any),
		refs:    make(map[string]*openapi3.SchemaRef),
		spec:    spec,
		genOpts: opts,
	}
}

// MustRegisterType registers a type and panics on error (for use during initialization)
func (r *SchemaRegistry) MustRegisterType(v any) *openapi3.SchemaRef {
	ref, err := r.RegisterType(v)
	if err != nil {
		panic(fmt.Sprintf("failed to register schema for type %T: %v", v, err))
	}

	return ref
}

// RegisterType registers a type and returns its schema reference
func (r *SchemaRegistry) RegisterType(v any) (*openapi3.SchemaRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
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

	// Register the type
	r.schemas[typeName] = v

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

// GetRef returns a reference to a registered schema by type
func (r *SchemaRegistry) GetRef(v any) *openapi3.SchemaRef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	typeName := t.Name()
	if typeName == "" {
		typeName = t.String()
	}

	return r.refs[typeName]
}

// GetOrRegister gets an existing schema reference or registers the type
func (r *SchemaRegistry) GetOrRegister(v any) (*openapi3.SchemaRef, error) {
	// Fast path: check if already exists
	if ref := r.GetRef(v); ref != nil {
		return ref, nil
	}

	// Slow path: register the type
	return r.RegisterType(v)
}

// MustGetOrRegister gets an existing schema reference or registers the type, panics on error
func (r *SchemaRegistry) MustGetOrRegister(v any) *openapi3.SchemaRef {
	ref, err := r.GetOrRegister(v)
	if err != nil {
		panic(fmt.Sprintf("failed to get or register schema for type %T: %v", v, err))
	}

	return ref
}

// RegisterAll registers multiple types at once
func (r *SchemaRegistry) RegisterAll(models map[string]any) error {
	for _, model := range models {
		if _, err := r.RegisterType(model); err != nil {
			return err
		}
	}

	return nil
}

// MustRegisterAll registers multiple types at once, panics on any error
func (r *SchemaRegistry) MustRegisterAll(models map[string]any) {
	for _, model := range models {
		r.MustRegisterType(model)
	}
}

// GetSchemas returns all registered schemas (for backwards compatibility)
func (r *SchemaRegistry) GetSchemas() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make(map[string]any, len(r.schemas))
	for k, v := range r.schemas {
		result[k] = v
	}

	return result
}
