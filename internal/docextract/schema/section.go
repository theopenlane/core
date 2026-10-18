package schema

import (
	"strings"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

// Field is one property a section asks the model to return
type Field struct {
	// Name is the property key, an entity field's input key or an extra's own name
	Name string
	// Schema constrains the property's value
	Schema *docextract.Schema
	// required makes the model always return the property
	required bool
}

// From returns the property for an entity field, keyed and typed from its descriptor
func From(field entityops.FieldDescriptor) Field {
	return Field{Name: field.Name, Schema: typeSchema(field.Type)}
}

// Extra returns a property the entity has no field for, carried to link or review the record; a TypeArray
// extra is a list of strings
func Extra(name string, valueType docextract.SchemaType) Field {
	return Field{Name: name, Schema: valueSchema(valueType)}
}

// Objects returns a property holding a list of objects carrying the fields
func Objects(name string, fields ...Field) Field {
	return Field{Name: name, Schema: &docextract.Schema{Type: docextract.TypeArray, Items: object(fields)}}
}

// Nested returns a property holding a single object carrying the fields
func Nested(name string, fields ...Field) Field {
	return Field{Name: name, Schema: object(fields)}
}

// Required makes the model always return the property
func (f Field) Required() Field {
	f.required = true

	return f
}

// Describe sets the guidance the model is given for the property
func (f Field) Describe(description string) Field {
	described := *f.Schema
	described.Description = description
	f.Schema = &described

	return f
}

// Enum restricts the property to the given values
func (f Field) Enum(values ...string) Field {
	restricted := *f.Schema
	restricted.Enum = values
	f.Schema = &restricted

	return f
}

// Section builds the schema for a section the model returns as a list of objects under name
func Section(name string, fields ...Field) *docextract.Schema {
	return wrap(name, &docextract.Schema{Type: docextract.TypeArray, Items: object(fields)})
}

// Object builds the schema for a section the model returns as a single object under name
func Object(name string, fields ...Field) *docextract.Schema {
	return wrap(name, object(fields))
}

// object builds the schema of one object carrying the fields
func object(fields []Field) *docextract.Schema {
	item := &docextract.Schema{Type: docextract.TypeObject, Properties: make(map[string]*docextract.Schema, len(fields))}

	for _, f := range fields {
		item.Properties[f.Name] = f.Schema

		if f.required {
			item.Required = append(item.Required, f.Name)
		}
	}

	return item
}

// wrap places a section's value under its name in the response object
func wrap(name string, value *docextract.Schema) *docextract.Schema {
	return &docextract.Schema{
		Type:       docextract.TypeObject,
		Properties: map[string]*docextract.Schema{name: value},
		Required:   []string{name},
	}
}

// typeSchema maps an entity field's go type to the value schema the model returns for it
func typeSchema(fieldType string) *docextract.Schema {
	switch {
	case fieldType == "bool":
		return valueSchema(docextract.TypeBoolean)
	case strings.HasPrefix(fieldType, "[]"):
		return valueSchema(docextract.TypeArray)
	default:
		return valueSchema(docextract.TypeString)
	}
}

// valueSchema returns the schema of a scalar value, or of a list of strings for TypeArray
func valueSchema(valueType docextract.SchemaType) *docextract.Schema {
	if valueType == docextract.TypeArray {
		return &docextract.Schema{Type: docextract.TypeArray, Items: &docextract.Schema{Type: docextract.TypeString}}
	}

	return &docextract.Schema{Type: valueType}
}
