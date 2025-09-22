package directives

import (
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/stoewer/go-strcase"
	"github.com/vektah/gqlparser/v2/ast"
)

// Extension is an implementation of entc.Extension
type Extension struct {
	entc.DefaultExtension
}

// ensure Extension implements the entc.Extension interface
var _ entc.Extension = (*Extension)(nil)

// ExtensionOption allow for control over the behavior of the generator
type ExtensionOption func(*Extension) error

// NewExtension returns an entc Extension that allows the entx package to generate
// the schema changes and templates needed to function
func NewExtension(opts ...ExtensionOption) (*Extension, error) {
	e := &Extension{}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	return e, nil
}

// SchemaHooks of the extension to seamlessly edit the final gql interface
func (e *Extension) SchemaHooks() []entgql.SchemaHook {
	return []entgql.SchemaHook{addReadOnlyDirectiveHook}
}

// addReadOnlyDirectiveHook is used to add the @readOnly directive to input fields that are marked as @hidden
// this prevents hidden fields from being set in create and update mutations
// as of today, there is no way to annotate a schema to do this automatically so we use a schema
// addReadOnlyDirectiveHook to modify the generated schema
var addReadOnlyDirectiveHook = func(_ *gen.Graph, s *ast.Schema) error {
	for _, t := range s.Types {
		// if the type is an input object, we want to check its fields for directives
		// otherwise, skip it
		if t.Kind != ast.InputObject {
			continue
		}

		object := s.Types[getInputObjectName(t.Name)]
		if object == nil {
			continue
		}

		for _, f := range t.Fields {
			setReadOnlyDirective(f, object, t)
		}

	}
	return nil
}

// setReadOnlyDirective checks if a field in an input object corresponds to a field in the main object
// that is marked with the @hidden directive. If it is, it adds the @readOnly directive to the input field
// and also to the clear<FieldName> field if it exists
func setReadOnlyDirective(f *ast.FieldDefinition, object *ast.Definition, t *ast.Definition) {
	// get the directives from the corresponding object field
	field := object.Fields.ForName(f.Name)
	if field == nil {
		return
	}

	if field.Directives == nil {
		return
	}

	// if the field is marked as hidden, we need to mark it as readOnly
	// so that it cannot be set in mutations
	for _, d := range field.Directives {
		if d.Name == Hidden {
			f.Directives = append(f.Directives, &ast.Directive{Name: ReadOnly})

			// if the field is marked as read only, we also need to make the clear<FieldName> field read only
			clearField := "clear" + strcase.UpperCamelCase(f.Name)
			if t.Fields.ForName(clearField) != nil {
				t.Fields.ForName(clearField).Directives = append(t.Fields.ForName(clearField).Directives, &ast.Directive{Name: ReadOnly})
			}
		}
	}
}

// getInputObjectName returns the input object name by stripping the CRUD operation from the resolver name
// for example UpdateTaskInput will return Task
func getInputObjectName(objectName string) string {
	// replace all operations
	objectName = strings.ReplaceAll(objectName, "Create", "")
	objectName = strings.ReplaceAll(objectName, "Update", "")

	return strings.ReplaceAll(objectName, "Input", "")
}
