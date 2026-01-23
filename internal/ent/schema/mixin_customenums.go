package schema

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/internal/ent/hooks"
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
	// WorkflowEdgeEligible marks the enum edge as workflow-eligible
	WorkflowEdgeEligible bool
	// GlobalEnum marks the enum as a shared set across schemas
	GlobalEnum bool
}

// newCustomEnumMixin creates a new CustomEnumMixin with the given schema type and options
func newCustomEnumMixin(schemaType any, opts ...customEnumOptions) CustomEnumMixin {
	c := CustomEnumMixin{
		schemaType: schemaType,
		fieldName:  "kind",
	}

	for _, opt := range opts {
		opt(&c)
	}

	sch := toSchemaFuncs(schemaType)

	if c.GlobalEnum {
		hooks.RegisterGlobalEnum(c.fieldName, sch.PluralName())
	} else {
		hooks.RegisterEnumSchema(sch.Name(), sch.PluralName())
	}

	return c
}

// customEnumOptions defines options for the CustomEnumMixin
type customEnumOptions func(*CustomEnumMixin)

// withEnumFieldName sets the field name for the CustomEnumMixin, it will default to "kind" if not set
func withEnumFieldName(fieldName string) customEnumOptions {
	return func(c *CustomEnumMixin) {
		c.fieldName = fieldName
	}
}

// withWorkflowEnumEdges marks the enum edge as workflow-eligible
func withWorkflowEnumEdges() customEnumOptions {
	return func(c *CustomEnumMixin) {
		c.WorkflowEdgeEligible = true
	}
}

// withGlobalEnum marks the enum as global across schemas
func withGlobalEnum() customEnumOptions {
	return func(c *CustomEnumMixin) {
		c.GlobalEnum = true
	}
}

// Fields of the CustomEnumMixin.
func (c CustomEnumMixin) Fields() []ent.Field {
	schema := toSchemaFuncs(c.schemaType)
	fields := []ent.Field{
		field.String(c.getEnumFieldName()).
			Comment("the " + c.fieldName + " of the " + schema.Name()).
			Optional(),
		field.String(c.getEnumEdgeName() + "_id").
			Comment("the " + c.fieldName + " of the " + schema.Name()).
			Optional(),
	}

	return fields
}

// Edges of the CustomEnumMixin
func (c CustomEnumMixin) Edges() []ent.Edge {
	s := c.schemaType

	edgeDef := &edgeDefinition{
		t:          CustomTypeEnum.Type,
		name:       c.getEnumEdgeName(),
		field:      c.getEnumEdgeName() + "_id",
		fromSchema: s,
		annotations: []schema.Annotation{
			accessmap.EdgeNoAuthCheck(),
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

// Hooks of the CustomEnumMixin
func (c CustomEnumMixin) Hooks() []ent.Hook {
	sch := toSchemaFuncs(c.schemaType)
	objectType := sch.Name()

	// subcontrol uses the control enums for its kind field
	if strings.EqualFold(objectType, "subcontrol") {
		objectType = "control"
	}

	in := hooks.CustomEnumFilter{
		ObjectType:      objectType,
		Field:           c.fieldName,
		EdgeFieldName:   c.getEnumEdgeName() + "_id",
		SchemaFieldName: c.getEnumFieldName(),
		AllowGlobal:     c.GlobalEnum,
	}
	return []ent.Hook{
		hooks.HookCustomEnums(in),
	}
}

// getEnumTypeValue returns the value of the enum type for the object the enum applies to
func (c CustomEnumMixin) getEnumEdgeName() string {
	sch := toSchemaFuncs(c.schemaType)

	if c.GlobalEnum {
		return c.fieldName
	}

	return fmt.Sprintf("%s_%s", sch.Name(), c.fieldName)
}

// getEnumEdgeName returns the name of the edge for the enum
func (c CustomEnumMixin) getEnumFieldName() string {
	return c.getEnumEdgeName() + "_name"
}

// getEnumReverseRefName returns the name of the reverse reference on the enum
func (c CustomEnumMixin) getEnumReverseRefName() string {
	sch := toSchemaFuncs(c.schemaType)

	if c.GlobalEnum || c.fieldName == "kind" {
		return ""
	}

	return fmt.Sprintf("%s_%s", sch.Name(), c.fieldName)
}

// validateObjectType validates the object type field
func validateObjectType(t string) error {
	// empty value is valid for global enums
	if t == "" {
		return nil
	}

	// normalize to snake case for comparison
	t = strcase.SnakeCase(t)

	if hooks.IsValidObjectType(t) {
		return nil
	}

	return rout.InvalidField("object_type")
}
