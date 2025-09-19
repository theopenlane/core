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
	return []entgql.SchemaHook{hook}
}

var hook = func(_ *gen.Graph, s *ast.Schema) error {
	for _, t := range s.Types {
		object := s.Types[getInputObjectName(t.Name)]
		if object == nil {
			continue
		}

		// if the type is an input object, we need to check its fields for directives
		if t.Kind == ast.InputObject {
			for _, f := range t.Fields {
				setReadOnlyDirective(f, object, t)
			}
		}
	}
	return nil
}

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

	return
}

// getInputObjectName returns the input object name by stripping the CRUD operation from the resolver name
// for example UpdateTaskInput will return Task
func getInputObjectName(objectName string) string {
	// replace all operations
	objectName = strings.ReplaceAll(objectName, "Create", "")
	objectName = strings.ReplaceAll(objectName, "Update", "")

	return strings.ReplaceAll(objectName, "Input", "")
}

var _ entc.Extension = (*Extension)(nil)
