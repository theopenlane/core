package schema

import (
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/mixin"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/utils/rout"
)

// CustomEnumMixin holds the schema definition for the custom enums
type CustomEnumMixin struct {
	mixin.Schema

	// schemaType is the type of the schema the enum applies to
	schemaType any

	// fieldName is the name of the field on the object the enum applies to, defaults to "kind" but could be customized
	// to support different enum fields like "category"
	fieldName string
}

func newCustomEnumMixin(schemaType any, opts ...customEnumOptions) CustomEnumMixin {
	c := CustomEnumMixin{
		schemaType: schemaType,
		fieldName:  "kind",
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

type customEnumOptions func(*CustomEnumMixin)

func withEnumFieldName(fieldName string) customEnumOptions {
	return func(c *CustomEnumMixin) {
		c.fieldName = fieldName
	}
}

// Fields of the CustomEnumMixin.
func (c CustomEnumMixin) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the CustomEnumMixin.
func (c CustomEnumMixin) Edges() []ent.Edge {
	s := c.schemaType

	edgeDef := &edgeDefinition{
		t:          CustomTypeEnum.Type,
		name:       c.getEnumFieldName(),
		fromSchema: s,
		annotations: []schema.Annotation{
			accessmap.EdgeViewCheck(CustomTypeEnum{}.Name()),
		},
	}

	ref := c.getEnumReverseRefName()
	if ref != "" {
		edgeDef.ref = ref
	}

	return []ent.Edge{
		uniqueEdgeTo(edgeDef),
	}
}

func (CustomEnumMixin) Hooks() []ent.Hook {
	return []ent.Hook{}
}

func (c CustomEnumMixin) getEnumFieldName() string {
	sch := toSchemaFuncs(c.schemaType)

	return fmt.Sprintf("%s_%s", sch.Name(), c.fieldName)
}

func (c CustomEnumMixin) getEnumReverseRefName() string {
	sch := toSchemaFuncs(c.schemaType)

	if c.fieldName == "kind" {
		return ""
	}

	return fmt.Sprintf("%s_%s", sch.PluralName(), c.fieldName)
}

// validObjectTypes is a set of valid object types for CustomTypeEnum
var validObjectTypes = map[string]struct{}{
	Task{}.Name():           {},
	Control{}.Name():        {},
	Subcontrol{}.Name():     {},
	Risk{}.Name():           {},
	InternalPolicy{}.Name(): {},
	Procedure{}.Name():      {},
	ActionPlan{}.Name():     {},
	Program{}.Name():        {},
}

// validateObjectType validates the object type field
func validateObjectType(t string) error {
	// check for empty value
	if t == "" {
		return rout.InvalidField("object_type")
	}

	// normalize to snake case for comparison
	t = strcase.SnakeCase(t)

	if _, ok := validObjectTypes[t]; ok {
		return nil
	}

	return rout.InvalidField("object_type")
}
