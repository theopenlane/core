package registry

import (
	"errors"
	"sync"
)

// Registry is an interface that defines methods for managing JSON schema definitions
type Registry interface {
	// GetSchema returns a jsonschema definition by name
	GetSchema(name string) (string, error)

	// SetSchema sets a jsonschema definition by name
	SetSchema(name string, schema string) error

	// DeleteSchema deletes a jsonschema definition by name
	DeleteSchema(name string) error

	// ListSchemas returns a list of all jsonschema definitions
	ListSchemas() ([]string, error)
}

// NewRegistry creates a new jsonschema registry
func NewRegistry() Registry {
	return &registry{
		schemas: make(map[string]string),
	}
}

type registry struct {
	schemas map[string]string
	mu      sync.RWMutex
}

var ErrorSchemaNotFound = errors.New("schema not found")

func (r *registry) GetSchema(name string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schema, ok := r.schemas[name]
	if !ok {
		return "", ErrorSchemaNotFound
	}

	return schema, nil
}

func (r *registry) SetSchema(name string, schema string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.schemas[name] = schema

	return nil
}

func (r *registry) DeleteSchema(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.schemas, name)

	return nil
}

func (r *registry) ListSchemas() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var schemas []string
	for name := range r.schemas {
		schemas = append(schemas, name)
	}

	return schemas, nil
}
